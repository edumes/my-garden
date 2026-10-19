//go:build integration

package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/my-garden/api/internal/database"
	"github.com/my-garden/api/internal/garden"
	"github.com/stretchr/testify/require"
)

// createLink creates an access link as the caller and returns its token.
func (a *testAPI) createLink(gardenID string, extra map[string]any) string {
	a.t.Helper()
	body := map[string]any{"garden_id": gardenID, "permissions": []string{"view"}}
	for k, v := range extra {
		body[k] = v
	}
	var created struct {
		AccessLink struct {
			Token string `json:"token"`
		} `json:"access_link"`
	}
	a.call(http.MethodPost, "/garden-shares/access-links", body, http.StatusCreated, &created)
	return created.AccessLink.Token
}

// joinConcurrently has every session redeem token at the same time and
// counts the responses by status.
func joinConcurrently(a *testAPI, sessions []session, token string) map[int]int {
	statuses := make([]int, len(sessions))
	var wg sync.WaitGroup
	for i, s := range sessions {
		wg.Add(1)
		go func() {
			defer wg.Done()
			statuses[i], _ = a.as(s).do(http.MethodPost, "/garden-shares/join", map[string]string{"token": token})
		}()
	}
	wg.Wait()

	counts := map[int]int{}
	for _, status := range statuses {
		counts[status]++
	}
	return counts
}

func linkState(t *testing.T, db *database.Database, token string) (usedCount int, shares int64) {
	t.Helper()
	var link garden.GardenAccessLink
	require.NoError(t, db.DB.Where("token = ?", token).Take(&link).Error)
	require.NoError(t, db.DB.Model(&garden.GardenShare{}).Where("garden_id = ?", link.GardenID).Count(&shares).Error)
	return link.UsedCount, shares
}

func TestAccessLinkMaxUsesHoldsUnderConcurrency(t *testing.T) {
	a, db := newTestAPI(t)
	owner := a.register("link_owner")
	a.token = owner.Token
	token := a.createLink(a.createGarden("Popular Garden"), map[string]any{"max_uses": 3})
	require.GreaterOrEqual(t, len(token), 26, "tokens need at least 128 bits of entropy")

	var players []session
	for i := range 10 {
		players = append(players, a.register(fmt.Sprintf("racer%d", i)))
	}

	counts := joinConcurrently(a, players, token)
	require.Equal(t, map[int]int{http.StatusOK: 3, http.StatusGone: 7}, counts)

	usedCount, shares := linkState(t, db, token)
	require.Equal(t, 3, usedCount)
	require.EqualValues(t, 3, shares)
}

func TestJoiningIsIdempotent(t *testing.T) {
	a, db := newTestAPI(t)
	owner := a.register("twice_owner")
	a.token = owner.Token
	gardenID := a.createGarden("Garden")
	token := a.createLink(gardenID, nil)

	player := a.register("eager")
	counts := joinConcurrently(a, []session{player, player, player, player, player}, token)
	require.Equal(t, map[int]int{http.StatusOK: 5}, counts)

	a.as(player).call(http.MethodPost, "/garden-shares/join", map[string]string{"token": token}, http.StatusOK, nil)
	usedCount, shares := linkState(t, db, token)
	require.Equal(t, 1, usedCount, "repeated joins must not consume uses")
	require.EqualValues(t, 1, shares)

	// The owner cannot redeem a link to their own garden either.
	a.call(http.MethodPost, "/garden-shares/join", map[string]string{"token": token}, http.StatusConflict, nil)
	usedCount, _ = linkState(t, db, token)
	require.Equal(t, 1, usedCount)
}

func TestUnavailableAccessLinksFailClosed(t *testing.T) {
	a, db := newTestAPI(t)
	owner := a.register("closed_owner")
	a.token = owner.Token
	gardenID := a.createGarden("Garden")
	player := a.as(a.register("latecomer"))

	expired := a.createLink(gardenID, nil)
	require.NoError(t, db.DB.Model(&garden.GardenAccessLink{}).Where("token = ?", expired).
		UpdateColumn("expires_at", time.Now().Add(-time.Minute)).Error)
	player.call(http.MethodPost, "/garden-shares/join", map[string]string{"token": expired}, http.StatusGone, nil)

	exhausted := a.createLink(gardenID, map[string]any{"max_uses": 1})
	first := a.as(a.register("first"))
	first.call(http.MethodPost, "/garden-shares/join", map[string]string{"token": exhausted}, http.StatusOK, nil)
	player.call(http.MethodPost, "/garden-shares/join", map[string]string{"token": exhausted}, http.StatusGone, nil)
	// A member reopening a used-up link still gets their access back.
	first.call(http.MethodPost, "/garden-shares/join", map[string]string{"token": exhausted}, http.StatusOK, nil)

	deactivated := a.createLink(gardenID, nil)
	var link garden.GardenAccessLink
	require.NoError(t, db.DB.Where("token = ?", deactivated).Take(&link).Error)
	a.call(http.MethodPost, fmt.Sprintf("/garden-shares/garden/%s/access-links/%s/deactivate", gardenID, link.ID), nil, http.StatusOK, nil)
	player.call(http.MethodPost, "/garden-shares/join", map[string]string{"token": deactivated}, http.StatusNotFound, nil)

	player.call(http.MethodPost, "/garden-shares/join", map[string]string{"token": "not-a-real-token"}, http.StatusNotFound, nil)
}

func TestSharingResponsesHidePrivateProfiles(t *testing.T) {
	a, _ := newTestAPI(t)
	owner := a.register("private_owner")
	a.token = owner.Token
	gardenID := a.createGarden("Garden")
	token := a.createLink(gardenID, nil)

	guest := a.as(a.register("private_guest"))
	guest.call(http.MethodPost, "/garden-shares/join", map[string]string{"token": token}, http.StatusOK, nil)

	for _, path := range []string{"/gardens", "/garden-shares/shared-with-me"} {
		status, raw := guest.do(http.MethodGet, path, nil)
		require.Equal(t, http.StatusOK, status)
		require.NotContains(t, string(raw), "private_owner@example.com", path)
		require.Contains(t, string(raw), `"username":"private_owner"`, path)
	}

	status, raw := a.do(http.MethodGet, "/garden-shares/garden/"+gardenID, nil)
	require.Equal(t, http.StatusOK, status)
	require.False(t, strings.Contains(string(raw), "private_guest@example.com"))
}
