//go:build integration

package main

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/my-garden/api/internal/garden"
	"github.com/stretchr/testify/require"
)

// TestGardenPermissionMatrix checks every garden action against every kind of
// caller. Anyone without unexpired access gets 404 so private gardens stay
// invisible; callers whose access lacks the capability get 403.
func TestGardenPermissionMatrix(t *testing.T) {
	a, db := newTestAPI(t)

	actors := []string{"owner", "manage", "harvest", "plant", "view", "expired", "stranger"}
	sessions := map[string]session{}
	for _, name := range actors {
		sessions[name] = a.register("matrix_" + name)
	}
	owner := sessions["owner"].User.ID

	var lettuce garden.PlantType
	require.NoError(t, db.DB.Where("name = ?", "Lettuce").First(&lettuce).Error)

	// newGarden creates a garden owned by "owner", shared with every other
	// actor according to its name, and with one harvestable plant.
	newGarden := func(t *testing.T) (gardenID, plantID uuid.UUID) {
		g := garden.Garden{UserID: owner, Name: "Matrix Garden"}
		require.NoError(t, db.DB.Create(&g).Error)

		yesterday := time.Now().Add(-24 * time.Hour)
		shares := map[string]struct {
			permissions garden.GardenPermissions
			expiresAt   *time.Time
		}{
			"manage":  {garden.GardenPermissions{garden.GardenPermissionManage}, nil},
			"harvest": {garden.GardenPermissions{garden.GardenPermissionHarvest}, nil},
			"plant":   {garden.GardenPermissions{garden.GardenPermissionPlant}, nil},
			"view":    {garden.GardenPermissions{garden.GardenPermissionView}, nil},
			"expired": {garden.GardenPermissions{garden.GardenPermissionManage}, &yesterday},
		}
		for name, share := range shares {
			require.NoError(t, db.DB.Create(&garden.GardenShare{
				GardenID:    g.ID,
				UserID:      sessions[name].User.ID,
				Permissions: share.permissions,
				SharedBy:    owner,
				ExpiresAt:   share.expiresAt,
			}).Error)
		}

		plant := ripePlant(t, db, g.ID, owner, lettuce, 0)
		return g.ID, plant.ID
	}

	type request struct {
		method string
		path   string
		body   any
	}
	actions := map[string]struct {
		request func(gardenID, plantID uuid.UUID) request
		// Expected status per actor, in the order of actors.
		want []int
	}{
		"view garden": {
			func(g, _ uuid.UUID) request { return request{http.MethodGet, "/gardens/" + g.String(), nil} },
			[]int{200, 200, 200, 200, 200, 404, 404},
		},
		"rename garden": {
			func(g, _ uuid.UUID) request {
				return request{http.MethodPut, "/gardens/" + g.String(), map[string]string{"name": "Renamed"}}
			},
			[]int{200, 200, 403, 403, 403, 404, 404},
		},
		"plant seed": {
			func(g, _ uuid.UUID) request {
				return request{http.MethodPost, "/gardens/" + g.String() + "/plants", map[string]any{
					"plant_type_id": lettuce.ID, "position": 1,
				}}
			},
			[]int{201, 201, 403, 201, 403, 404, 404},
		},
		"harvest plant": {
			func(g, p uuid.UUID) request {
				return request{http.MethodPost, fmt.Sprintf("/gardens/%s/plants/%s/harvest", g, p), nil}
			},
			[]int{200, 200, 200, 403, 403, 404, 404},
		},
		"remove plant": {
			func(g, p uuid.UUID) request {
				return request{http.MethodDelete, fmt.Sprintf("/gardens/%s/plants/%s", g, p), nil}
			},
			[]int{200, 200, 403, 403, 403, 404, 404},
		},
		"delete garden": {
			func(g, _ uuid.UUID) request { return request{http.MethodDelete, "/gardens/" + g.String(), nil} },
			[]int{200, 403, 403, 403, 403, 404, 404},
		},
		"list shares": {
			func(g, _ uuid.UUID) request {
				return request{http.MethodGet, "/garden-shares/garden/" + g.String(), nil}
			},
			[]int{200, 403, 403, 403, 403, 404, 404},
		},
		// Access links carry redeemable tokens, so only the owner sees them.
		"list access links": {
			func(g, _ uuid.UUID) request {
				return request{http.MethodGet, "/garden-shares/garden/" + g.String() + "/access-links", nil}
			},
			[]int{200, 403, 403, 403, 403, 404, 404},
		},
	}

	// Planting needs seeds in the caller's own inventory.
	for _, actor := range actors {
		require.NoError(t, db.DB.Create(&garden.SeedInventory{
			UserID: sessions[actor].User.ID, PlantTypeID: lettuce.ID, Quantity: len(actions),
		}).Error)
	}

	for action, tc := range actions {
		for i, actor := range actors {
			t.Run(action+"/"+actor, func(t *testing.T) {
				gardenID, plantID := newGarden(t)
				req := tc.request(gardenID, plantID)
				caller := &testAPI{t: t, base: a.base, token: sessions[actor].Token}
				caller.call(req.method, req.path, req.body, tc.want[i], nil)
			})
		}
	}
}

// TestAuditLogsAreUserScoped checks that audit endpoints never expose another
// user's actions.
func TestAuditLogsAreUserScoped(t *testing.T) {
	a, _ := newTestAPI(t)
	alice := a.register("alice")
	bob := a.register("bob")

	a.token = alice.Token
	var created struct {
		Garden struct {
			ID string `json:"id"`
		} `json:"garden"`
	}
	a.call(http.MethodPost, "/gardens", map[string]string{"name": "Alice's Garden"}, http.StatusCreated, &created)

	a.token = bob.Token
	var logs struct {
		Logs []struct {
			UserID *uuid.UUID `json:"user_id"`
		} `json:"logs"`
	}
	a.call(http.MethodGet, "/audit/logs?limit=100", nil, http.StatusOK, &logs)
	for _, log := range logs.Logs {
		require.NotNil(t, log.UserID)
		require.Equal(t, bob.User.ID, *log.UserID)
	}

	a.call(http.MethodGet, "/audit/users/"+alice.User.ID.String()+"/activity", nil, http.StatusForbidden, nil)

	var gardenActivity struct {
		Count int `json:"count"`
	}
	a.call(http.MethodGet, "/audit/gardens/"+created.Garden.ID+"/activity", nil, http.StatusOK, &gardenActivity)
	require.Zero(t, gardenActivity.Count)

	var stats struct {
		Stats struct {
			TopActions []struct {
				Action string `json:"action"`
			} `json:"top_actions"`
		} `json:"stats"`
	}
	a.call(http.MethodGet, "/audit/stats", nil, http.StatusOK, &stats)
	for _, action := range stats.Stats.TopActions {
		require.NotEqual(t, "garden_create", action.Action, "Bob never created a garden")
	}
}
