//go:build integration

package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/my-garden/api/internal/config"
	"github.com/my-garden/api/internal/database"
	"github.com/my-garden/api/internal/testdb"
	"github.com/my-garden/api/internal/weather"
	"github.com/my-garden/api/pkg/game"
	"github.com/stretchr/testify/require"
)

func weatherCount(t *testing.T, db *database.Database) int64 {
	t.Helper()
	var count int64
	require.NoError(t, db.DB.Model(&weather.Weather{}).Count(&count).Error)
	return count
}

func TestWeatherChangesOncePerInterval(t *testing.T) {
	db, _ := testdb.New(t)
	repo := weather.NewRepository(db)
	// Drawing weather takes a while, so refreshes that start together all
	// find the old weather expired before any new one is saved.
	sunny := func() weather.Weather {
		time.Sleep(50 * time.Millisecond)
		return weather.Weather{Condition: weather.WeatherSunny, Temperature: 20, Humidity: 50, GrowthMultiplier: 1.2}
	}
	t0 := time.Now().Truncate(time.Second)

	// Several instances refresh at once; one draws the weather, all see it.
	ids := make([]uuid.UUID, 8)
	var wg sync.WaitGroup
	for i := range ids {
		wg.Add(1)
		go func() {
			defer wg.Done()
			current, err := repo.RefreshWeather(t0, 5*time.Minute, sunny)
			if err != nil {
				t.Error(err)
				return
			}
			ids[i] = current.ID
		}()
	}
	wg.Wait()
	require.EqualValues(t, 1, weatherCount(t, db))
	for _, id := range ids {
		require.Equal(t, ids[0], id)
	}

	// A restart within the interval keeps the weather.
	current, err := repo.RefreshWeather(t0.Add(4*time.Minute), 5*time.Minute, sunny)
	require.NoError(t, err)
	require.Equal(t, ids[0], current.ID)

	next, err := repo.RefreshWeather(t0.Add(5*time.Minute), 5*time.Minute, sunny)
	require.NoError(t, err)
	require.NotEqual(t, ids[0], next.ID)
	require.EqualValues(t, 2, weatherCount(t, db))
}

func TestEngineTickGrowsPlantsUnderCurrentWeather(t *testing.T) {
	a, db := newTestAPI(t)
	player := a.register("ticker")
	gardenID := uuid.MustParse(a.as(player).createGarden("Tick Garden"))
	t0 := time.Now().Add(-time.Hour).Truncate(time.Second)
	plant := seedling(t, db, gardenID, player.User.ID, plantTypeNamed(t, db, "Lettuce"), t0)

	cfg := &config.Config{Game: config.GameConfig{TickInterval: time.Minute, WeatherUpdateInterval: time.Hour, Speed: 1}}
	engine := game.NewGameEngine(db, cfg)
	require.NoError(t, engine.Tick(t0.Add(15*time.Minute)))
	require.NoError(t, engine.Tick(t0.Add(15*time.Minute)), "a repeated tick changes nothing")

	current, err := weather.NewRepository(db).GetCurrentWeather()
	require.NoError(t, err)
	require.EqualValues(t, 1, weatherCount(t, db))
	require.InDelta(t, 50*current.GrowthMultiplier, reload(t, db, plant).GrowthProgress, 1e-6,
		"15 of Lettuce's 30 minutes, scaled by the weather")

	// A fresh engine has no in-memory tick history, as after an API restart.
	beforeRestart := reload(t, db, plant)
	restarted := game.NewGameEngine(db, cfg)
	require.NoError(t, restarted.Tick(t0.Add(15*time.Minute)))
	require.Equal(t, beforeRestart.GrowthProgress, reload(t, db, plant).GrowthProgress)
	require.EqualValues(t, 1, weatherCount(t, db), "restart preserves valid weather")

	status := engine.Status()
	require.Equal(t, "stopped", status.State, "Tick runs without Start")
	require.NotNil(t, status.LastTick)
}

func TestReadinessFollowsTheDatabase(t *testing.T) {
	a, db := newTestAPI(t)
	root := strings.TrimSuffix(a.base, "/api/v1")

	ready := func() (int, map[string]any) {
		res, err := http.Get(root + "/ready")
		require.NoError(t, err)
		defer res.Body.Close()
		var body map[string]any
		require.NoError(t, json.NewDecoder(res.Body).Decode(&body))
		return res.StatusCode, body
	}

	status, body := ready()
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, "ready", body["status"])
	checks := body["checks"].(map[string]any)
	require.Equal(t, "up", checks["database"].(map[string]any)["status"])
	require.Equal(t, "stopped", checks["game_engine"].(map[string]any)["status"])

	sqlDB, err := db.DB.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())
	status, body = ready()
	require.Equal(t, http.StatusServiceUnavailable, status)
	require.Equal(t, "not_ready", body["status"])
}
