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

// seedling plants a fresh seed of pt, grown by grower, that last grew at t0.
func seedling(t *testing.T, db *database.Database, gardenID, grower uuid.UUID, pt garden.PlantType, t0 time.Time) garden.Plant {
	t.Helper()
	plant := garden.Plant{
		GardenID: gardenID, PlantTypeID: pt.ID, GrowerID: grower,
		Stage: garden.PlantStageSeed, PlantedAt: t0, GrownAt: t0,
	}
	require.NoError(t, db.DB.Create(&plant).Error)
	return plant
}

func reload(t *testing.T, db *database.Database, plant garden.Plant) garden.Plant {
	t.Helper()
	require.NoError(t, db.DB.Take(&plant, "id = ?", plant.ID).Error)
	return plant
}

func TestGrowthFollowsElapsedTime(t *testing.T) {
	a, db := newTestAPI(t)
	player := a.register("timekeeper")
	gardenID := uuid.MustParse(a.as(player).createGarden("Clock Garden"))
	lettuce := plantTypeNamed(t, db, "Lettuce") // ripens in 30 minutes
	t0 := time.Now().Add(-time.Hour).Truncate(time.Second)
	plant := seedling(t, db, gardenID, player.User.ID, lettuce, t0)

	require.NoError(t, garden.GrowPlants(db.DB, t0.Add(15*time.Minute), 1, 1))
	halfway := reload(t, db, plant)
	require.InDelta(t, 50, halfway.GrowthProgress, 1e-6)
	require.Equal(t, garden.PlantStageGrowing, halfway.Stage)

	// Repeating a tick, as after a restart, adds nothing.
	require.NoError(t, garden.GrowPlants(db.DB, t0.Add(15*time.Minute), 1, 1))
	require.InDelta(t, 50, reload(t, db, plant).GrowthProgress, 1e-6)

	require.NoError(t, garden.GrowPlants(db.DB, t0.Add(30*time.Minute), 1, 1))
	ripe := reload(t, db, plant)
	require.Equal(t, 100.0, ripe.GrowthProgress)
	require.Equal(t, garden.PlantStageHarvestable, ripe.Stage)
}

func TestOverlappingTicksNeverLoseOrRepeatGrowth(t *testing.T) {
	a, db := newTestAPI(t)
	player := a.register("overlapper")
	gardenID := uuid.MustParse(a.as(player).createGarden("Busy Garden"))
	pumpkin := plantTypeNamed(t, db, "Pumpkin") // ripens in 100 minutes
	t0 := time.Now().Add(-time.Hour).Truncate(time.Second)
	plant := seedling(t, db, gardenID, player.User.ID, pumpkin, t0)

	// Ticks of several instances, each with its own clock reading, run at once.
	var wg sync.WaitGroup
	for minutes := 1; minutes <= 20; minutes++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := garden.GrowPlants(db.DB, t0.Add(time.Duration(minutes)*time.Minute), 1, 1); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	require.NoError(t, garden.GrowPlants(db.DB, t0.Add(20*time.Minute), 1, 1))

	require.InDelta(t, 20, reload(t, db, plant).GrowthProgress, 1e-6, "20 minutes of a 100-minute plant, however the ticks interleave")
}

func TestSharedGardenHarvestPaysTheGrower(t *testing.T) {
	a, db := newTestAPI(t)
	owner := a.register("plot_owner")
	member := a.register("plot_member")
	gardenID := a.as(owner).createGarden("Community Plot")
	require.NoError(t, db.DB.Create(&garden.GardenShare{
		GardenID:    uuid.MustParse(gardenID),
		UserID:      member.User.ID,
		Permissions: garden.GardenPermissions{garden.GardenPermissionPlant, garden.GardenPermissionHarvest},
		SharedBy:    owner.User.ID,
	}).Error)

	lettuce := plantTypeNamed(t, db, "Lettuce")
	giveSeeds(t, db, member.User.ID, lettuce, 1)
	var planted struct {
		Plant struct {
			ID       uuid.UUID `json:"id"`
			GrowerID uuid.UUID `json:"grower_id"`
		} `json:"plant"`
	}
	a.as(member).call(http.MethodPost, "/gardens/"+gardenID+"/plants", map[string]any{
		"plant_type_id": lettuce.ID, "position": 0,
	}, http.StatusCreated, &planted)
	require.Equal(t, member.User.ID, planted.Plant.GrowerID)
	require.NoError(t, db.DB.Model(&garden.Plant{}).Where("id = ?", planted.Plant.ID).
		Updates(map[string]any{"stage": garden.PlantStageHarvestable, "growth_progress": 100}).Error)

	// The owner harvests, but the member spent the seed.
	a.as(owner).call(http.MethodPost, fmt.Sprintf("/gardens/%s/plants/%s/harvest", gardenID, planted.Plant.ID), nil, http.StatusOK, nil)

	require.Equal(t, 100, coinsOf(t, db, owner.User.ID))
	require.Zero(t, experienceOf(t, db, owner.User.ID))
	require.Equal(t, 100+lettuce.HarvestValue*lettuce.Yield, coinsOf(t, db, member.User.ID))
	require.Equal(t, auth.PlantingExperience+lettuce.GrowthTime, experienceOf(t, db, member.User.ID))
}

func TestLevelsUnlockSeedsAndGardens(t *testing.T) {
	a, db := newTestAPI(t)
	player := a.register("climber")
	climber := a.as(player)
	tomato := plantTypeNamed(t, db, "Tomato")
	require.Equal(t, 2, tomato.UnlockLevel)

	climber.call(http.MethodPost, "/store/buy", map[string]any{"plant_type_id": tomato.ID, "quantity": 1}, http.StatusForbidden, nil)
	require.Equal(t, 100, coinsOf(t, db, player.User.ID), "a locked purchase charges nothing")

	climber.createGarden("First")
	climber.call(http.MethodPost, "/gardens", map[string]string{"name": "Second"}, http.StatusForbidden, nil)

	require.NoError(t, db.DB.Model(&auth.User{}).Where("id = ?", player.User.ID).Update("level", 3).Error)
	climber.call(http.MethodPost, "/store/buy", map[string]any{"plant_type_id": tomato.ID, "quantity": 1}, http.StatusOK, nil)
	climber.createGarden("Second")
	climber.call(http.MethodPost, "/gardens", map[string]string{"name": "Third"}, http.StatusForbidden, nil)
}

func TestGardenLimitHoldsUnderConcurrency(t *testing.T) {
	a, _ := newTestAPI(t)
	player := a.as(a.register("land_grabber"))
	counts := concurrently(5, func(i int) int {
		status, _ := player.do(http.MethodPost, "/gardens", map[string]string{"name": fmt.Sprintf("Garden %d", i)})
		return status
	})
	require.Equal(t, map[int]int{http.StatusCreated: 1, http.StatusForbidden: 4}, counts)
}

func TestExperienceRaisesLevels(t *testing.T) {
	a, db := newTestAPI(t)
	player := a.register("almost_there")
	a.token = player.Token
	gardenID := a.createGarden("Level Garden")
	lettuce := plantTypeNamed(t, db, "Lettuce")
	giveSeeds(t, db, player.User.ID, lettuce, 1)
	require.NoError(t, db.DB.Model(&auth.User{}).Where("id = ?", player.User.ID).Update("experience", 99).Error)

	var planted struct {
		Reward struct {
			Experience int  `json:"experience"`
			Level      int  `json:"level"`
			LevelUp    bool `json:"level_up"`
		} `json:"reward"`
	}
	a.call(http.MethodPost, "/gardens/"+gardenID+"/plants", map[string]any{
		"plant_type_id": lettuce.ID, "position": 0,
	}, http.StatusCreated, &planted)
	require.Equal(t, auth.PlantingExperience, planted.Reward.Experience)
	require.True(t, planted.Reward.LevelUp)
	require.Equal(t, 2, planted.Reward.Level)

	var profile struct {
		User struct {
			Level               int `json:"level"`
			LevelExperience     int `json:"level_experience"`
			NextLevelExperience int `json:"next_level_experience"`
		} `json:"user"`
	}
	a.call(http.MethodGet, "/users/profile", nil, http.StatusOK, &profile)
	require.Equal(t, 2, profile.User.Level)
	require.Equal(t, 100, profile.User.LevelExperience)
	require.Equal(t, 300, profile.User.NextLevelExperience)
}
