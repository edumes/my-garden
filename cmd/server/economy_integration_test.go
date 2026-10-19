//go:build integration

package main

import (
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/my-garden/api/internal/auth"
	"github.com/my-garden/api/internal/database"
	"github.com/my-garden/api/internal/garden"
	"github.com/stretchr/testify/require"
)

// concurrently runs request n times at once and counts the returned statuses.
func concurrently(n int, request func(i int) int) map[int]int {
	statuses := make([]int, n)
	var wg sync.WaitGroup
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			statuses[i] = request(i)
		}()
	}
	wg.Wait()

	counts := map[int]int{}
	for _, status := range statuses {
		counts[status]++
	}
	return counts
}

func plantTypeNamed(t *testing.T, db *database.Database, name string) garden.PlantType {
	t.Helper()
	var pt garden.PlantType
	require.NoError(t, db.DB.Where("name = ?", name).Take(&pt).Error)
	return pt
}

func coinsOf(t *testing.T, db *database.Database, userID uuid.UUID) int {
	t.Helper()
	var user auth.User
	require.NoError(t, db.DB.Take(&user, "id = ?", userID).Error)
	return user.Coins
}

func giveSeeds(t *testing.T, db *database.Database, userID uuid.UUID, pt garden.PlantType, quantity int) {
	t.Helper()
	require.NoError(t, db.DB.Create(&garden.SeedInventory{
		UserID: userID, PlantTypeID: pt.ID, Quantity: quantity, PurchasedAt: time.Now(),
	}).Error)
}

func seedsOf(t *testing.T, db *database.Database, userID uuid.UUID, pt garden.PlantType) (quantity int, rows int64) {
	t.Helper()
	var inventory []garden.SeedInventory
	require.NoError(t, db.DB.Where("user_id = ? AND plant_type_id = ?", userID, pt.ID).Find(&inventory).Error)
	for _, row := range inventory {
		quantity += row.Quantity
	}
	return quantity, int64(len(inventory))
}

func plantsIn(t *testing.T, db *database.Database, gardenID string) int64 {
	t.Helper()
	var count int64
	require.NoError(t, db.DB.Model(&garden.Plant{}).Where("garden_id = ?", gardenID).Count(&count).Error)
	return count
}

func TestSeedPurchasesNeverOverspend(t *testing.T) {
	a, db := newTestAPI(t)
	player := a.register("shopper")
	tomato := plantTypeNamed(t, db, "Tomato")
	require.Equal(t, 100, player.User.Coins)
	require.Equal(t, 20, tomato.SeedPrice)

	buyer := a.as(player)
	counts := concurrently(10, func(int) int {
		status, _ := buyer.do(http.MethodPost, "/store/buy", map[string]any{"plant_type_id": tomato.ID, "quantity": 1})
		return status
	})
	require.Equal(t, map[int]int{http.StatusOK: 5, http.StatusBadRequest: 5}, counts)

	require.Equal(t, 0, coinsOf(t, db, player.User.ID))
	quantity, rows := seedsOf(t, db, player.User.ID, tomato)
	require.Equal(t, 5, quantity)
	require.EqualValues(t, 1, rows, "purchases of one seed type share an inventory row")
}

func TestPlantingSpendsOneSeedPerPlant(t *testing.T) {
	a, db := newTestAPI(t)
	player := a.register("planter")
	a.token = player.Token
	gardenID := a.createGarden("Garden")
	lettuce := plantTypeNamed(t, db, "Lettuce")
	giveSeeds(t, db, player.User.ID, lettuce, 2)

	counts := concurrently(6, func(i int) int {
		status, _ := a.do(http.MethodPost, "/gardens/"+gardenID+"/plants", map[string]any{
			"plant_type_id": lettuce.ID, "position": i,
		})
		return status
	})
	require.Equal(t, map[int]int{http.StatusCreated: 2, http.StatusBadRequest: 4}, counts)

	quantity, _ := seedsOf(t, db, player.User.ID, lettuce)
	require.Zero(t, quantity)
	require.EqualValues(t, 2, plantsIn(t, db, gardenID))
}

func TestOneCellHoldsOnePlant(t *testing.T) {
	a, db := newTestAPI(t)
	player := a.register("crowder")
	a.token = player.Token
	gardenID := a.createGarden("Garden")
	lettuce := plantTypeNamed(t, db, "Lettuce")
	giveSeeds(t, db, player.User.ID, lettuce, 5)

	counts := concurrently(5, func(int) int {
		status, _ := a.do(http.MethodPost, "/gardens/"+gardenID+"/plants", map[string]any{
			"plant_type_id": lettuce.ID, "position": 4,
		})
		return status
	})
	require.Equal(t, map[int]int{http.StatusCreated: 1, http.StatusConflict: 4}, counts)

	quantity, _ := seedsOf(t, db, player.User.ID, lettuce)
	require.Equal(t, 4, quantity, "a rejected planting must give its seed back")
	require.EqualValues(t, 1, plantsIn(t, db, gardenID))
}

func TestHarvestPaysOnce(t *testing.T) {
	a, db := newTestAPI(t)
	player := a.register("harvester")
	a.token = player.Token
	gardenID := a.createGarden("Garden")
	lettuce := plantTypeNamed(t, db, "Lettuce")

	plant := garden.Plant{
		GardenID:       uuid.MustParse(gardenID),
		PlantTypeID:    lettuce.ID,
		Position:       0,
		Stage:          garden.PlantStageHarvestable,
		GrowthProgress: 100,
		PlantedAt:      time.Now(),
	}
	require.NoError(t, db.DB.Create(&plant).Error)

	counts := concurrently(5, func(int) int {
		status, _ := a.do(http.MethodPost, fmt.Sprintf("/gardens/%s/plants/%s/harvest", gardenID, plant.ID), nil)
		return status
	})
	require.Equal(t, map[int]int{http.StatusOK: 1, http.StatusBadRequest: 4}, counts)
	require.Equal(t, player.User.Coins+lettuce.HarvestValue*lettuce.Yield, coinsOf(t, db, player.User.ID))
}

func TestDatabaseRejectsNegativeBalances(t *testing.T) {
	a, db := newTestAPI(t)
	player := a.register("debtor")
	lettuce := plantTypeNamed(t, db, "Lettuce")
	giveSeeds(t, db, player.User.ID, lettuce, 1)

	require.Error(t, db.DB.Exec("UPDATE users SET coins = -1 WHERE id = ?", player.User.ID).Error)
	require.Error(t, db.DB.Exec("UPDATE seed_inventories SET quantity = -1 WHERE user_id = ?", player.User.ID).Error)
	require.Error(t, db.DB.Create(&garden.SeedInventory{
		UserID: player.User.ID, PlantTypeID: lettuce.ID, Quantity: 1,
	}).Error, "one inventory row per user and seed type")
}
