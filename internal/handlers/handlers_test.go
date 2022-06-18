package handlers

import (
	"github.com/IgorPestretsov/LoyaltySystem/internal/sqlStorage"
	"github.com/IgorPestretsov/LoyaltySystem/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRegisterUser(t *testing.T) {
	type want struct {
		statusCode int
	}
	tests := []struct {
		name    string
		creds   string
		request string
		want    want
	}{
		{
			name:    "test1",
			request: "/api/user/register",
			creds:   "{\"login\":\"user\",\"password\": \"pass\"}",
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name:    "test2",
			request: "/api/user/register",
			creds:   "{\"login\":\"existed_user\",\"password\": \"pass\"}",
			want: want{
				statusCode: http.StatusConflict,
			},
		},
		{
			name:    "test3",
			request: "/api/user/register",
			creds:   "",
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := sqlStorage.NewSQLStorage("password=P@ssw0rd dbname=loyaltySystem sslmode=disable host=localhost port=5432 user=user ")
			defer s.DropTableUsers()
			s.SaveUser(storage.User{Login: "existed_user", Password: "pass"})
			r := chi.NewRouter()

			r.Post(tt.request, func(rw http.ResponseWriter, r *http.Request) {
				RegisterUser(rw, r, s)
			})

			req := httptest.NewRequest(http.MethodPost, tt.request, strings.NewReader(tt.creds))
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)
			result := w.Result()
			defer result.Body.Close()
			assert.Equal(t, tt.want.statusCode, result.StatusCode)
		})
	}
}
