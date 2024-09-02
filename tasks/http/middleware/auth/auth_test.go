package auth_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"gitlab.com/slon/shad-go/middleware/auth"
)

type fakeChecker map[string]struct {
	user *auth.User
	err  error
}

func (c fakeChecker) CheckToken(ctx context.Context, token string) (*auth.User, error) {
	res := c[token]
	return res.user, res.err
}

func TestAuth(t *testing.T) {
	m := chi.NewRouter()

	c := fakeChecker{
		"token0": {
			user: &auth.User{Name: "Fedor", Email: "dartslon@gmail.com"},
		},

		"token1": {
			err: fmt.Errorf("database offline"),
		},

		"token2": {
			err: fmt.Errorf("token expired: %w", auth.ErrInvalidToken),
		},
	}

	m.Use(auth.CheckAuth(c))

	var (
		lastUser   *auth.User
		lastUserOK bool
		called     bool
	)

	m.Get("/path/ok", func(w http.ResponseWriter, r *http.Request) {
		called = true
		lastUser, lastUserOK = auth.ContextUser(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	m.Get("/path/error", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusConflict)
	})

	t.Run("NoToken", func(t *testing.T) {
		called = false

		w := httptest.NewRecorder()
		m.ServeHTTP(w, httptest.NewRequest("GET", "/path/ok", nil))
		require.Equal(t, http.StatusUnauthorized, w.Code)
		require.False(t, called)
	})

	t.Run("InvalidToken", func(t *testing.T) {
		called = false

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/path/ok", nil)
		r.Header.Add("authorization", "Bearer token2")

		m.ServeHTTP(w, r)
		require.Equal(t, http.StatusUnauthorized, w.Code)
		require.False(t, called)
	})

	t.Run("DatabaseError", func(t *testing.T) {
		called = false

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/path/ok", nil)
		r.Header.Add("authorization", "Bearer token1")

		m.ServeHTTP(w, r)
		require.Equal(t, http.StatusInternalServerError, w.Code)
		require.False(t, called)
	})

	t.Run("GoodToken", func(t *testing.T) {
		called = false
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/path/ok", nil)
		r.Header.Add("authorization", "Bearer token0")

		m.ServeHTTP(w, r)
		require.Equal(t, http.StatusOK, w.Code)
		require.True(t, called)
		require.True(t, lastUserOK)
		require.Equal(t, &auth.User{Name: "Fedor", Email: "dartslon@gmail.com"}, lastUser)

		called = false
		w = httptest.NewRecorder()
		r = httptest.NewRequest("GET", "/path/error", nil)
		r.Header.Add("authorization", "Bearer token0")

		m.ServeHTTP(w, r)
		require.Equal(t, http.StatusConflict, w.Code)
		require.True(t, called)
	})
}
