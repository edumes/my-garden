//go:build integration

package main

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRateLimitGuardsSignInAndMutations(t *testing.T) {
	t.Setenv("RATE_LIMIT_REQUESTS", "3")
	a, _ := newTestAPI(t)

	player := a.register("hasty")
	badLogin := map[string]string{"username": "hasty", "password": "guess"}
	a.call(http.MethodPost, "/auth/login", badLogin, http.StatusUnauthorized, nil)
	a.call(http.MethodPost, "/auth/login", badLogin, http.StatusUnauthorized, nil)
	a.call(http.MethodPost, "/auth/login", badLogin, http.StatusTooManyRequests, nil)
	a.as(player).call(http.MethodPost, "/gardens", map[string]string{"name": "Blocked"}, http.StatusTooManyRequests, nil)

	// Reads stay available.
	a.as(player).call(http.MethodGet, "/users/profile", nil, http.StatusOK, nil)

	// No proxy is trusted by default, so a forged X-Forwarded-For does not
	// give the client a fresh budget.
	req, err := http.NewRequest(http.MethodPost, a.base+"/auth/login", strings.NewReader(`{}`))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-For", "203.0.113.7")
	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	res.Body.Close()
	require.Equal(t, http.StatusTooManyRequests, res.StatusCode)
}
