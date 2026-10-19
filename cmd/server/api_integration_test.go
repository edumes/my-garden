//go:build integration

package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/my-garden/api/internal/database"
	"github.com/my-garden/api/internal/testdb"
	"github.com/my-garden/api/pkg/game"
	"github.com/stretchr/testify/require"
)

// testAPI calls the real router, backed by a fresh database, over HTTP.
type testAPI struct {
	t     *testing.T
	base  string
	token string
}

func newTestAPI(t *testing.T) (*testAPI, *database.Database) {
	gin.SetMode(gin.TestMode)
	db, cfg := testdb.New(t)

	// The engine is not started; tests run its ticks themselves.
	server := httptest.NewServer(newRouter(db, game.NewGameEngine(db, cfg), cfg))
	t.Cleanup(server.Close)

	return &testAPI{t: t, base: server.URL + "/api/v1"}, db
}

// do sends body as JSON and returns the status and raw response. It reports
// transport errors with t.Errorf, so it is safe to use from goroutines.
func (a *testAPI) do(method, path string, body any) (int, []byte) {
	a.t.Helper()

	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			a.t.Errorf("encode %s %s: %v", method, path, err)
			return 0, nil
		}
		reader = bytes.NewReader(encoded)
	}
	req, err := http.NewRequest(method, a.base+path, reader)
	if err != nil {
		a.t.Errorf("build %s %s: %v", method, path, err)
		return 0, nil
	}
	req.Header.Set("Content-Type", "application/json")
	if a.token != "" {
		req.Header.Set("Authorization", "Bearer "+a.token)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		a.t.Errorf("%s %s: %v", method, path, err)
		return 0, nil
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		a.t.Errorf("read %s %s: %v", method, path, err)
	}
	return res.StatusCode, raw
}

// call sends body as JSON, requires wantStatus, and decodes the response into out.
func (a *testAPI) call(method, path string, body any, wantStatus int, out any) {
	a.t.Helper()
	status, raw := a.do(method, path, body)
	require.Equalf(a.t, wantStatus, status, "%s %s: %s", method, path, raw)
	if out != nil {
		require.NoError(a.t, json.Unmarshal(raw, out))
	}
}

// as returns a client for the same server authenticated as s.
func (a *testAPI) as(s session) *testAPI {
	return &testAPI{t: a.t, base: a.base, token: s.Token}
}

func (a *testAPI) createGarden(name string) string {
	a.t.Helper()
	var created struct {
		Garden struct {
			ID string `json:"id"`
		} `json:"garden"`
	}
	a.call(http.MethodPost, "/gardens", map[string]string{"name": name}, http.StatusCreated, &created)
	return created.Garden.ID
}

type session struct {
	Token string `json:"token"`
	User  struct {
		ID       uuid.UUID `json:"id"`
		Username string    `json:"username"`
		Coins    int       `json:"coins"`
	} `json:"user"`
}

func (a *testAPI) register(username string) session {
	a.t.Helper()
	var s session
	a.call(http.MethodPost, "/auth/register", map[string]string{
		"username": username,
		"email":    username + "@example.com",
		"password": "correct-horse",
	}, http.StatusCreated, &s)
	return s
}

func TestAuthentication(t *testing.T) {
	a, _ := newTestAPI(t)

	registered := a.register("gardener")
	require.NotEmpty(t, registered.Token)
	require.Equal(t, "gardener", registered.User.Username)

	a.call(http.MethodPost, "/auth/register", map[string]string{
		"username": "gardener", "email": "other@example.com", "password": "correct-horse",
	}, http.StatusConflict, nil)
	a.call(http.MethodPost, "/auth/login", map[string]string{
		"username": "gardener", "password": "wrong-password",
	}, http.StatusUnauthorized, nil)
	a.call(http.MethodGet, "/users/profile", nil, http.StatusUnauthorized, nil)

	var loggedIn session
	a.call(http.MethodPost, "/auth/login", map[string]string{
		"username": "gardener", "password": "correct-horse",
	}, http.StatusOK, &loggedIn)
	require.NotEmpty(t, loggedIn.Token)

	a.token = loggedIn.Token
	var profile struct {
		User struct {
			Username string `json:"username"`
		} `json:"user"`
	}
	a.call(http.MethodGet, "/users/profile", nil, http.StatusOK, &profile)
	require.Equal(t, "gardener", profile.User.Username)
}

func TestGardenHappyPath(t *testing.T) {
	a, db := newTestAPI(t)
	player := a.register("grower")
	a.token = player.Token

	var created struct {
		Garden struct {
			ID string `json:"id"`
		} `json:"garden"`
	}
	a.call(http.MethodPost, "/gardens", map[string]string{"name": "Test Garden"}, http.StatusCreated, &created)
	gardenID := created.Garden.ID

	type plantType struct {
		ID           string `json:"id"`
		Name         string `json:"name"`
		SeedPrice    int    `json:"seed_price"`
		HarvestValue int    `json:"harvest_value"`
		Yield        int    `json:"yield"`
	}
	var catalog struct {
		PlantTypes []plantType `json:"plant_types"`
	}
	a.call(http.MethodGet, "/plants", nil, http.StatusOK, &catalog)
	var lettuce plantType
	for _, pt := range catalog.PlantTypes {
		if pt.Name == "Lettuce" {
			lettuce = pt
		}
	}
	require.NotEmpty(t, lettuce.ID, "seed data should include Lettuce")

	var purchase struct {
		UserCoins int `json:"user_coins"`
	}
	a.call(http.MethodPost, "/store/buy", map[string]any{
		"plant_type_id": lettuce.ID, "quantity": 1,
	}, http.StatusOK, &purchase)
	require.Equal(t, player.User.Coins-lettuce.SeedPrice, purchase.UserCoins)

	var planted struct {
		Plant struct {
			ID    string `json:"id"`
			Stage string `json:"stage"`
		} `json:"plant"`
	}
	a.call(http.MethodPost, "/gardens/"+gardenID+"/plants", map[string]any{
		"plant_type_id": lettuce.ID, "position": 0,
	}, http.StatusCreated, &planted)
	require.Equal(t, "seed", planted.Plant.Stage)

	// Growth happens on the game engine's tick, which these tests do not run.
	require.NoError(t, db.DB.Exec(
		"UPDATE plants SET stage = 'harvestable', growth_progress = 100 WHERE id = ?", planted.Plant.ID,
	).Error)

	var harvested struct {
		Reward struct {
			UserID     uuid.UUID `json:"user_id"`
			Coins      int       `json:"coins"`
			Experience int       `json:"experience"`
		} `json:"reward"`
	}
	a.call(http.MethodPost, "/gardens/"+gardenID+"/plants/"+planted.Plant.ID+"/harvest", nil, http.StatusOK, &harvested)
	require.Equal(t, player.User.ID, harvested.Reward.UserID)
	require.Equal(t, lettuce.HarvestValue*lettuce.Yield, harvested.Reward.Coins)
	require.Greater(t, harvested.Reward.Coins, lettuce.SeedPrice, "a harvest pays back its seed")

	var profile struct {
		User struct {
			Coins      int `json:"coins"`
			Experience int `json:"experience"`
		} `json:"user"`
	}
	a.call(http.MethodGet, "/users/profile", nil, http.StatusOK, &profile)
	require.Equal(t, purchase.UserCoins+harvested.Reward.Coins, profile.User.Coins)
	require.Equal(t, 2+harvested.Reward.Experience, profile.User.Experience, "planting and harvesting both grant experience")
	require.EqualValues(t, 0, plantsIn(t, db, gardenID), "harvesting frees the plot")
}
