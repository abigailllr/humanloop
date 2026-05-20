package middleware

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserIDKey contextKey = "user_id"
const UserEmailKey contextKey = "user_email"
const UserNameKey contextKey = "user_name"
const BuyerDatasetKey contextKey = "buyer_dataset_id"

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		secret := jwtSecret()

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, UserIDKey, stringClaim(claims, "sub"))
		ctx = context.WithValue(ctx, UserEmailKey, stringClaim(claims, "email"))
		ctx = context.WithValue(ctx, UserNameKey, stringClaim(claims, "name"))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func APIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-API-Key")
		expected := os.Getenv("API_KEY")
		if expected == "" || subtle.ConstantTimeCompare([]byte(key), []byte(expected)) != 1 {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type BuyerKeyLookup func(ctx context.Context, hash string) (datasetID string, ok bool)

var BuyerKeyFunc BuyerKeyLookup

func BuyerKeyOrAPIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if raw := r.Header.Get("X-Buyer-Key"); raw != "" && BuyerKeyFunc != nil {
			h := sha256.Sum256([]byte(raw))
			hash := hex.EncodeToString(h[:])
			if datasetID, ok := BuyerKeyFunc(r.Context(), hash); ok {
				ctx := context.WithValue(r.Context(), BuyerDatasetKey, datasetID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		key := r.Header.Get("X-API-Key")
		expected := os.Getenv("API_KEY")
		if expected == "" || subtle.ConstantTimeCompare([]byte(key), []byte(expected)) != 1 {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func jwtSecret() string {
	s := os.Getenv("JWT_SECRET")
	if s == "" {
		return "dev-secret-change-in-production"
	}
	return s
}

func stringClaim(claims jwt.MapClaims, key string) string {
	v, _ := claims[key].(string)
	return v
}
