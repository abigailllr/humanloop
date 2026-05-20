package handlers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwk"
	ljwt "github.com/lestrrat-go/jwx/v2/jwt"

	"github.com/abigailtech/humanloop/backend/db"
	"github.com/abigailtech/humanloop/backend/models"
)

const maxAuthBody = 16 * 1024

func AuthGoogle(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxAuthBody)

	var body struct {
		IDToken string `json:"id_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.IDToken == "" {
		http.Error(w, "missing id_token", http.StatusBadRequest)
		return
	}

	info, err := verifyGoogleToken(body.IDToken)
	if err != nil {
		http.Error(w, "invalid google token", http.StatusUnauthorized)
		return
	}

	userID := "google:" + info.Sub
	if db.Pool != nil {
		db.UpsertUser(r.Context(), models.User{ID: userID, Email: info.Email, Name: info.Name})
	}

	resp, err := issueTokenPair(r.Context(), userID, info.Email, info.Name)
	if err != nil {
		http.Error(w, "could not issue token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func AuthApple(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxAuthBody)

	var body struct {
		IdentityToken string `json:"identity_token"`
		UserID        string `json:"user_id"`
		Email         string `json:"email"`
		Name          string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.IdentityToken == "" {
		http.Error(w, "missing identity_token", http.StatusBadRequest)
		return
	}
	if body.UserID == "" {
		http.Error(w, "missing user_id", http.StatusBadRequest)
		return
	}

	claims, err := verifyAppleToken(body.IdentityToken)
	if err != nil {
		http.Error(w, "invalid apple token", http.StatusUnauthorized)
		return
	}

	sub, _ := claims.Get("sub")
	if sub == nil || fmt.Sprint(sub) != body.UserID {
		http.Error(w, "invalid apple token", http.StatusUnauthorized)
		return
	}

	email := body.Email
	if e, ok := claims.Get("email"); ok && e != nil {
		email = fmt.Sprint(e)
	}

	userID := "apple:" + body.UserID
	if db.Pool != nil {
		db.UpsertUser(r.Context(), models.User{ID: userID, Email: email, Name: body.Name})
	}

	resp, err := issueTokenPair(r.Context(), userID, email, body.Name)
	if err != nil {
		http.Error(w, "could not issue token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func AuthRefresh(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxAuthBody)

	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.RefreshToken == "" {
		http.Error(w, "missing refresh_token", http.StatusBadRequest)
		return
	}

	if db.Pool == nil {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
		return
	}

	hash := hashToken(body.RefreshToken)
	userID, tokenID, err := db.GetRefreshToken(r.Context(), hash)
	if err != nil {
		http.Error(w, "invalid or expired refresh token", http.StatusUnauthorized)
		return
	}

	user, err := db.GetUser(r.Context(), userID)
	if err != nil {
		http.Error(w, "user not found", http.StatusUnauthorized)
		return
	}

	if err := db.DeleteRefreshToken(r.Context(), tokenID); err != nil {
		http.Error(w, "token rotation failed", http.StatusInternalServerError)
		return
	}

	resp, err := issueTokenPair(r.Context(), userID, user.Email, user.Name)
	if err != nil {
		http.Error(w, "could not issue token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func AuthRevoke(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxAuthBody)

	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.RefreshToken == "" {
		http.Error(w, "missing refresh_token", http.StatusBadRequest)
		return
	}

	if db.Pool != nil {
		db.DeleteRefreshTokenByHash(r.Context(), hashToken(body.RefreshToken))
	}

	w.WriteHeader(http.StatusNoContent)
}

func issueTokenPair(ctx context.Context, userID, email, name string) (map[string]any, error) {
	accessToken, err := issueJWT(userID, email, name)
	if err != nil {
		return nil, err
	}

	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return nil, err
	}
	rawRefresh := hex.EncodeToString(raw)

	if db.Pool != nil {
		id := uuid.New().String()
		if err := db.CreateRefreshToken(ctx, id, userID, hashToken(rawRefresh)); err != nil {
			return nil, err
		}
	}

	return map[string]any{
		"access_token":  accessToken,
		"refresh_token": rawRefresh,
		"token_type":    "Bearer",
		"expires_in":    900,
	}, nil
}

func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

type googleTokenInfo struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Aud   string `json:"aud"`
	Error string `json:"error_description"`
}

var googleClient = &http.Client{Timeout: 10 * time.Second}

func verifyGoogleToken(idToken string) (*googleTokenInfo, error) {
	resp, err := googleClient.Get("https://oauth2.googleapis.com/tokeninfo?id_token=" + idToken)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 16*1024))
	var info googleTokenInfo
	if err := json.Unmarshal(raw, &info); err != nil {
		return nil, err
	}
	if info.Error != "" || info.Sub == "" {
		return nil, fmt.Errorf("google token invalid")
	}
	return &info, nil
}

func verifyAppleToken(identityToken string) (ljwt.Token, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	set, err := jwk.Fetch(ctx, "https://appleid.apple.com/auth/keys")
	if err != nil {
		return nil, fmt.Errorf("fetch apple keys: %w", err)
	}

	token, err := ljwt.Parse([]byte(identityToken),
		ljwt.WithKeySet(set),
		ljwt.WithValidate(true),
		ljwt.WithIssuer("https://appleid.apple.com"),
	)
	if err != nil {
		return nil, fmt.Errorf("apple token invalid: %w", err)
	}
	return token, nil
}

func issueJWT(sub, email, name string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret-change-in-production"
	}

	claims := jwt.MapClaims{
		"sub":   sub,
		"email": email,
		"name":  name,
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(15 * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
