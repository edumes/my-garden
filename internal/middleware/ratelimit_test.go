package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/synctest"
	"time"

	"github.com/gin-gonic/gin"
)

func TestRateLimitMutations(t *testing.T) {
	gin.SetMode(gin.TestMode)
	synctest.Test(t, func(t *testing.T) {
		router := gin.New()
		router.Use(RateLimitMutations(2, time.Minute))
		router.Any("/thing", func(c *gin.Context) { c.Status(http.StatusOK) })

		send := func(method, ip string) *httptest.ResponseRecorder {
			req := httptest.NewRequest(method, "/thing", nil)
			req.RemoteAddr = ip + ":1234"
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			return rec
		}
		expect := func(method, ip string, want int) *httptest.ResponseRecorder {
			t.Helper()
			rec := send(method, ip)
			if rec.Code != want {
				t.Fatalf("%s from %s: got %d, want %d", method, ip, rec.Code, want)
			}
			return rec
		}

		expect(http.MethodPost, "10.0.0.1", http.StatusOK)
		expect(http.MethodPut, "10.0.0.1", http.StatusOK)
		limited := expect(http.MethodDelete, "10.0.0.1", http.StatusTooManyRequests)
		if got := limited.Header().Get("Retry-After"); got != "60" {
			t.Errorf("Retry-After = %q, want 60", got)
		}

		expect(http.MethodGet, "10.0.0.1", http.StatusOK)
		expect(http.MethodPost, "10.0.0.2", http.StatusOK)

		time.Sleep(time.Minute)
		expect(http.MethodPost, "10.0.0.1", http.StatusOK)
	})
}
