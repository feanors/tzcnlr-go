package auth

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
	"time"
)

type AuthAPI struct {
	adminUsername string
	adminPassword string
	JWTSecretKey  []byte
}

func NewAuthAPI(adminUsername, adminPassword string, JWTSecretKey []byte) *AuthAPI {
	return &AuthAPI{
		adminUsername: strings.ToLower(adminUsername),
		adminPassword: adminPassword,
		JWTSecretKey:  JWTSecretKey,
	}
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func decodeCredentials(body []byte) (Credentials, error) {
	var credentials Credentials
	err := json.Unmarshal(body, &credentials)
	if err != nil {
		return credentials, err
	}
	return credentials, nil
}

func (api *AuthAPI) DecodeCredentialsBodyHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, ok := r.Context().Value("body").([]byte)
		if !ok {
			http.Error(w, "error accessing the body of the request", http.StatusInternalServerError)
			return
		}

		if len(body) == 0 {
			http.Error(w, "empty request body", http.StatusBadRequest)
			return
		}

		credentials, err := decodeCredentials(body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		credentials.Username = strings.ToLower(credentials.Username)

		ctx := context.WithValue(r.Context(), "credentials", credentials)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (api *AuthAPI) LoginHandler(w http.ResponseWriter, r *http.Request) {
	credentials, ok := r.Context().Value("credentials").(Credentials)
	if !ok {
		http.Error(w, "error during json decode", http.StatusInternalServerError)
		return
	}

	credentials.Username = strings.ToLower(credentials.Username)
	if credentials.Username != api.adminUsername || credentials.Password != api.adminPassword {
		http.Error(w, "wrong credentials", http.StatusUnauthorized)
		return
	}

	tokenString, err := api.GenerateJWT()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write([]byte(tokenString))
}

func (api *AuthAPI) GenerateJWT() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(730 * time.Hour)),
	})

	return token.SignedString(api.JWTSecretKey)
}

func (api *AuthAPI) ValidateTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		tokenString := r.Header.Get("Authorization")
		// goofy, fix TODO
		if len(tokenString) < 9 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		tokenString = tokenString[len("Bearer "):]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return api.JWTSecretKey, nil
		})
		if err != nil || !token.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
