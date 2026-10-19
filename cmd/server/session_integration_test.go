//go:build integration

package main

import (
	"net/http"
	"testing"
)

func (a *testAPI) login(username string) session {
	a.t.Helper()
	var s session
	a.call(http.MethodPost, "/auth/login", map[string]string{
		"username": username, "password": "correct-horse",
	}, http.StatusOK, &s)
	return s
}

func TestLogoutRevokesEverySession(t *testing.T) {
	a, _ := newTestAPI(t)
	laptop := a.register("traveler")
	phone := a.login("traveler")

	var refreshed session
	a.as(laptop).call(http.MethodPost, "/auth/refresh", nil, http.StatusOK, &refreshed)
	tablet := session{Token: refreshed.Token}
	a.as(tablet).call(http.MethodGet, "/users/profile", nil, http.StatusOK, nil)

	a.as(phone).call(http.MethodPost, "/auth/logout", nil, http.StatusOK, nil)

	for _, s := range []session{laptop, phone, tablet} {
		a.as(s).call(http.MethodGet, "/users/profile", nil, http.StatusUnauthorized, nil)
		a.as(s).call(http.MethodPost, "/auth/refresh", nil, http.StatusUnauthorized, nil)
		a.as(s).call(http.MethodGet, "/gardens", nil, http.StatusUnauthorized, nil)
	}

	a.as(a.login("traveler")).call(http.MethodGet, "/users/profile", nil, http.StatusOK, nil)
}

func TestLogoutOnlyRevokesTheCaller(t *testing.T) {
	a, _ := newTestAPI(t)
	leaving := a.register("leaving")
	staying := a.register("staying")

	a.as(leaving).call(http.MethodPost, "/auth/logout", nil, http.StatusOK, nil)
	a.as(staying).call(http.MethodGet, "/users/profile", nil, http.StatusOK, nil)
}
