//go:build integration

package main

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/my-garden/api/internal/auth"
	"github.com/my-garden/api/internal/database"
	"github.com/my-garden/api/internal/garden"
	"github.com/my-garden/api/internal/marketplace"
	"github.com/stretchr/testify/require"
)

func setCoins(t *testing.T, db *database.Database, userID uuid.UUID, coins int) {
	t.Helper()
	require.NoError(t, db.DB.Model(&auth.User{}).Where("id = ?", userID).Update("coins", coins).Error)
}

// harvestablePlant creates a ready-to-harvest lettuce, grown by the caller, in
// a new garden the caller owns, and returns its garden and plant IDs.
func harvestablePlant(t *testing.T, a *testAPI, db *database.Database, grower session) (gardenID string, plantID uuid.UUID) {
	t.Helper()
	gardenID = a.as(grower).createGarden("Market Garden")
	plant := ripePlant(t, db, uuid.MustParse(gardenID), grower.User.ID, plantTypeNamed(t, db, "Lettuce"), 0)
	return gardenID, plant.ID
}

func (a *testAPI) listPlant(plantID uuid.UUID, price int) string {
	a.t.Helper()
	var listing struct {
		ID string `json:"id"`
	}
	a.call(http.MethodPost, "/marketplace/listings", map[string]any{"plant_id": plantID, "price": price}, http.StatusCreated, &listing)
	return listing.ID
}

func TestPlantSaleMovesCoinsAndPlant(t *testing.T) {
	a, db := newTestAPI(t)
	seller := a.register("seller")
	buyer := a.register("buyer")
	setCoins(t, db, buyer.User.ID, 300)

	a.token = seller.Token
	_, plantID := harvestablePlant(t, a, db, seller)
	listingID := a.listPlant(plantID, 200)

	shopper := a.as(buyer)
	buyerGarden := shopper.createGarden("Buyer Garden")
	shopper.call(http.MethodPost, "/marketplace/listings/"+listingID+"/buy", nil, http.StatusOK, nil)

	require.Equal(t, 100, coinsOf(t, db, buyer.User.ID))
	require.Equal(t, 100+200-1, coinsOf(t, db, seller.User.ID), "0.5% commission is withheld from the seller")

	var plant garden.Plant
	require.NoError(t, db.DB.Take(&plant, "id = ?", plantID).Error)
	require.Equal(t, buyerGarden, plant.GardenID.String(), "the plant moves to the buyer's free plot")
	require.Equal(t, buyer.User.ID, plant.GrowerID, "the buyer now earns its harvest")
	require.Equal(t, auth.TradeExperience, experienceOf(t, db, seller.User.ID))

	var listing marketplace.PlantListing
	require.NoError(t, db.DB.Take(&listing, "id = ?", listingID).Error)
	require.Equal(t, marketplace.ListingStatusSold, listing.Status)
	require.Equal(t, buyer.User.ID, *listing.BuyerID)

	shopper.call(http.MethodPost, "/marketplace/listings/"+listingID+"/buy", nil, http.StatusConflict, nil)
	a.call(http.MethodPost, "/marketplace/listings/"+listingID+"/buy", nil, http.StatusBadRequest, nil)
}

func TestOnlyOneBuyerWinsAListing(t *testing.T) {
	a, db := newTestAPI(t)
	seller := a.register("popular_seller")
	a.token = seller.Token
	_, plantID := harvestablePlant(t, a, db, seller)
	listingID := a.listPlant(plantID, 50)

	var buyers []session
	for i := range 5 {
		buyer := a.register(fmt.Sprintf("bidder%d", i))
		a.as(buyer).createGarden("Bidder Garden")
		buyers = append(buyers, buyer)
	}
	counts := concurrently(len(buyers), func(i int) int {
		status, _ := a.as(buyers[i]).do(http.MethodPost, "/marketplace/listings/"+listingID+"/buy", nil)
		return status
	})
	require.Equal(t, map[int]int{http.StatusOK: 1, http.StatusConflict: 4}, counts)

	total := coinsOf(t, db, seller.User.ID)
	for _, b := range buyers {
		total += coinsOf(t, db, b.User.ID)
	}
	require.Equal(t, 100*(len(buyers)+1), total, "coins are only moved once (commission rounds to 0 here)")
}

func TestBuyerMustAffordListing(t *testing.T) {
	a, db := newTestAPI(t)
	seller := a.register("pricey_seller")
	buyer := a.register("thrifty_buyer")
	setCoins(t, db, buyer.User.ID, 10)

	a.token = seller.Token
	gardenID, plantID := harvestablePlant(t, a, db, seller)
	listingID := a.listPlant(plantID, 50)

	a.as(buyer).createGarden("Thrifty Garden")
	a.as(buyer).call(http.MethodPost, "/marketplace/listings/"+listingID+"/buy", nil, http.StatusBadRequest, nil)

	require.Equal(t, 10, coinsOf(t, db, buyer.User.ID))
	require.Equal(t, 100, coinsOf(t, db, seller.User.ID))
	var listing marketplace.PlantListing
	require.NoError(t, db.DB.Take(&listing, "id = ?", listingID).Error)
	require.Equal(t, marketplace.ListingStatusActive, listing.Status)
	require.EqualValues(t, 1, plantsIn(t, db, gardenID), "the plant stays with the seller")
}

func TestBuyerNeedsAFreePlot(t *testing.T) {
	a, db := newTestAPI(t)
	seller := a.register("plot_seller")
	buyer := a.register("plotless_buyer")

	a.token = seller.Token
	_, plantID := harvestablePlant(t, a, db, seller)
	listingID := a.listPlant(plantID, 50)

	a.as(buyer).call(http.MethodPost, "/marketplace/listings/"+listingID+"/buy", nil, http.StatusConflict, nil)

	require.Equal(t, 100, coinsOf(t, db, buyer.User.ID))
	require.Equal(t, 100, coinsOf(t, db, seller.User.ID))
	var gardens int64
	require.NoError(t, db.DB.Model(&garden.Garden{}).Where("user_id = ?", buyer.User.ID).Count(&gardens).Error)
	require.Zero(t, gardens, "no garden is created for the buyer")
}

func TestListingRules(t *testing.T) {
	a, db := newTestAPI(t)
	seller := a.register("rule_seller")
	a.token = seller.Token
	gardenID, plantID := harvestablePlant(t, a, db, seller)

	firstListing := a.listPlant(plantID, 30)
	a.call(http.MethodPost, "/marketplace/listings", map[string]any{"plant_id": plantID, "price": 40}, http.StatusConflict, nil)

	// Once a listing expires, the plant can be listed again.
	require.NoError(t, db.DB.Model(&marketplace.PlantListing{}).Where("id = ?", firstListing).
		Update("expires_at", time.Now().Add(-time.Minute)).Error)
	a.listPlant(plantID, 40)

	stranger := a.as(a.register("rule_stranger"))
	stranger.call(http.MethodPost, "/marketplace/listings", map[string]any{"plant_id": plantID, "price": 50}, http.StatusForbidden, nil)

	seedling := garden.Plant{
		GardenID: uuid.MustParse(gardenID), PlantTypeID: plantTypeNamed(t, db, "Lettuce").ID, GrowerID: seller.User.ID,
		Position: 1, Stage: garden.PlantStageSeed, PlantedAt: time.Now(),
	}
	require.NoError(t, db.DB.Create(&seedling).Error)
	a.call(http.MethodPost, "/marketplace/listings", map[string]any{"plant_id": seedling.ID, "price": 40}, http.StatusBadRequest, nil)

	// Deleting the garden removes the listing with its plants.
	a.call(http.MethodDelete, "/gardens/"+gardenID, nil, http.StatusOK, nil)
	var listings int64
	require.NoError(t, db.DB.Model(&marketplace.PlantListing{}).Where("plant_id = ?", plantID).Count(&listings).Error)
	require.Zero(t, listings)
}

// newDeliveryRequest creates a request for lettuce worth reward as the caller.
func (a *testAPI) newDeliveryRequest(db *database.Database, reward int) string {
	a.t.Helper()
	var request struct {
		ID string `json:"id"`
	}
	a.call(http.MethodPost, "/marketplace/delivery-requests", map[string]any{
		"plant_type_id": plantTypeNamed(a.t, db, "Lettuce").ID,
		"reward":        reward,
		"deadline":      time.Now().Add(2 * time.Hour).Format(time.RFC3339),
	}, http.StatusCreated, &request)
	return request.ID
}

func TestDeliveryRequestPaysOnce(t *testing.T) {
	a, db := newTestAPI(t)
	requester := a.register("requester")
	a.token = requester.Token
	requestID := a.newDeliveryRequest(db, 50)

	a.call(http.MethodPost, "/marketplace/delivery-requests/"+requestID+"/accept", nil, http.StatusBadRequest, nil)

	courier := a.as(a.register("courier"))
	courier.call(http.MethodPost, "/marketplace/delivery-requests/"+requestID+"/accept", nil, http.StatusOK, nil)
	a.as(a.register("latecomer_courier")).call(http.MethodPost, "/marketplace/delivery-requests/"+requestID+"/accept", nil, http.StatusConflict, nil)

	counts := concurrently(5, func(int) int {
		status, _ := courier.do(http.MethodPost, "/marketplace/delivery-requests/"+requestID+"/complete", nil)
		return status
	})
	require.Equal(t, map[int]int{http.StatusOK: 1, http.StatusConflict: 4}, counts)

	require.Equal(t, 50, coinsOf(t, db, requester.User.ID))
	var courierUser auth.User
	require.NoError(t, db.DB.Take(&courierUser, "username = ?", "courier").Error)
	require.Equal(t, 150, courierUser.Coins)
	require.Equal(t, auth.TradeExperience, courierUser.Experience, "completing a delivery grants trade experience once")
}

func TestLateDeliveryLosesTwentyPercent(t *testing.T) {
	a, db := newTestAPI(t)
	a.token = a.register("patient_requester").Token
	requestID := a.newDeliveryRequest(db, 50)

	courier := a.as(a.register("slow_courier"))
	courier.call(http.MethodPost, "/marketplace/delivery-requests/"+requestID+"/accept", nil, http.StatusOK, nil)
	require.NoError(t, db.DB.Model(&marketplace.DeliveryRequest{}).Where("id = ?", requestID).
		Update("deadline", time.Now().Add(-time.Minute)).Error)

	var completed struct {
		RewardPaid int `json:"reward_paid"`
	}
	courier.call(http.MethodPost, "/marketplace/delivery-requests/"+requestID+"/complete", nil, http.StatusOK, &completed)
	require.Equal(t, 40, completed.RewardPaid)
}

func TestDeliveryNeedsRequesterFunds(t *testing.T) {
	a, db := newTestAPI(t)
	requester := a.register("broke_requester")
	a.token = requester.Token
	requestID := a.newDeliveryRequest(db, 50)

	courier := a.as(a.register("unpaid_courier"))
	courier.call(http.MethodPost, "/marketplace/delivery-requests/"+requestID+"/accept", nil, http.StatusOK, nil)
	setCoins(t, db, requester.User.ID, 0)

	courier.call(http.MethodPost, "/marketplace/delivery-requests/"+requestID+"/complete", nil, http.StatusConflict, nil)
	var request marketplace.DeliveryRequest
	require.NoError(t, db.DB.Take(&request, "id = ?", requestID).Error)
	require.Equal(t, marketplace.RequestStatusAccepted, request.Status, "a failed completion changes nothing")

	// Requests the requester cannot pay today are refused up front.
	a.call(http.MethodPost, "/marketplace/delivery-requests", map[string]any{
		"plant_type_id": plantTypeNamed(t, db, "Lettuce").ID,
		"reward":        20,
		"deadline":      time.Now().Add(2 * time.Hour).Format(time.RFC3339),
	}, http.StatusBadRequest, nil)
}

func TestMarketplaceHidesPrivateProfiles(t *testing.T) {
	a, db := newTestAPI(t)
	publicSeller := a.register("public_seller")
	a.token = publicSeller.Token
	_, plantID := harvestablePlant(t, a, db, publicSeller)
	a.listPlant(plantID, 30)
	a.newDeliveryRequest(db, 20)

	browser := a.as(a.register("browser"))
	for _, path := range []string{"/marketplace/listings", "/marketplace/delivery-requests"} {
		status, raw := browser.do(http.MethodGet, path, nil)
		require.Equal(t, http.StatusOK, status)
		require.NotContains(t, string(raw), "public_seller@example.com", path)
		require.Contains(t, string(raw), `"username":"public_seller"`, path)
	}
}

func TestBlockchainEndpointsAreGone(t *testing.T) {
	a, _ := newTestAPI(t)
	a.token = a.register("legacy_client").Token
	a.call(http.MethodGet, "/wallet/balance", nil, http.StatusNotFound, nil)
	a.call(http.MethodGet, "/blockchain/ledger", nil, http.StatusNotFound, nil)
}
