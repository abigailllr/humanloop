package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func AuthGoogle(w http.ResponseWriter, r *http.Request) {
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

	token, err := issueJWT("google:"+info.Sub, info.Email, info.Name)
	if err != nil {
		http.Error(w, "could not issue token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func AuthApple(w http.ResponseWriter, r *http.Request) {
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

	if len(strings.Split(body.IdentityToken, ".")) != 3 {
		http.Error(w, "invalid apple token", http.StatusUnauthorized)
		return
	}
	if body.UserID == "" {
		http.Error(w, "missing user_id", http.StatusBadRequest)
		return
	}

	token, err := issueJWT("apple:"+body.UserID, body.Email, body.Name)
	if err != nil {
		http.Error(w, "could not issue token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

type googleTokenInfo struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Aud   string `json:"aud"`
	Error string `json:"error_description"`
}

func verifyGoogleToken(idToken string) (*googleTokenInfo, error) {
	resp, err := http.Get("https://oauth2.googleapis.com/tokeninfo?id_token=" + idToken)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	var info googleTokenInfo
	if err := json.Unmarshal(raw, &info); err != nil {
		return nil, err
	}
	if info.Error != "" || info.Sub == "" {
		return nil, http.ErrNoCookie
	}
	return &info, nil
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
		"exp":   time.Now().Add(30 * 24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
