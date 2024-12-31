package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	UserIDContextKey contextKey = "user_id"
	// You should use a secure method to store this key in production
	jwtSecret = "your-secret-key"
)

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func GenerateToken(userID string) (string, error) {
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

func getTokenFromRequest(r *http.Request) string {
	// First check cookie
	cookie, err := r.Cookie("authToken")
	if err == nil && cookie.Value != "" {
		log.Printf("Found auth cookie: %s", cookie.Value)
		return cookie.Value
	}
	log.Printf("No auth cookie found or error: %v", err)

	// Then check Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) == 2 && strings.ToLower(bearerToken[0]) == "bearer" {
			log.Printf("Found auth header token: %s", bearerToken[1])
			return bearerToken[1]
		}
	}
	log.Printf("No auth header found")

	return ""
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Auth middleware: %s %s", r.Method, r.URL.Path)

		// Skip auth for login and register pages
		if r.URL.Path == "/login" || r.URL.Path == "/register" || r.URL.Path == "/api/users/login" || r.URL.Path == "/api/users" {
			log.Printf("Skipping auth for public path: %s", r.URL.Path)
			next.ServeHTTP(w, r)
			return
		}

		// Skip auth for static files
		if strings.HasPrefix(r.URL.Path, "/static/") {
			log.Printf("Skipping auth for static file: %s", r.URL.Path)
			next.ServeHTTP(w, r)
			return
		}

		token := getTokenFromRequest(r)
		if token == "" {
			log.Printf("No token found in request")
			handleAuthError(w, r)
			return
		}

		claims := &Claims{}
		parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil {
			log.Printf("Token validation error: %v", err)
			// Check if the error is due to token expiration
			if err.Error() == "token has expired" {
				// Handle expired token
				handleTokenExpired(w, r)
				return
			}
			handleAuthError(w, r)
			return
		}

		if !parsedToken.Valid {
			log.Printf("Token is invalid")
			handleAuthError(w, r)
			return
		}

		log.Printf("Token is valid for user: %s", claims.UserID)
		ctx := context.WithValue(r.Context(), UserIDContextKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func handleAuthError(w http.ResponseWriter, r *http.Request) {
    // For API requests, return JSON error
    if strings.HasPrefix(r.URL.Path, "/api/") {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(map[string]string{
            "error": "Unauthorized",
        })
        return
    }
    
    // For page requests, redirect to login
    http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func handleTokenExpired(w http.ResponseWriter, r *http.Request) {
    // Clear the expired token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "authToken",
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-24 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

    // For API requests, return JSON error
    if strings.HasPrefix(r.URL.Path, "/api/") {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(map[string]string{
            "error": "Token expired",
        })
        return
    }
    
    // For page requests, redirect to login
    http.Redirect(w, r, "/login?expired=true", http.StatusSeeOther)
}

// GetUserID retrieves the user ID from the context
func GetUserID(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(UserIDContextKey).(string)
	if !ok {
		return "", fmt.Errorf("user ID not found in context")
	}
	return userID, nil
}
