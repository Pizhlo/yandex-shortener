package session

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

const TokenExp = time.Hour * 3
const SecretKey = "supersecretkey"

type Claims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID
}

// CookieMiddleware проверяет куки на валидность; если проверка не пройдена - создает новую куку.
func CookieMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				userID := uuid.New()
				makeCookie(w, userID)
				next.ServeHTTP(w, r)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		valid, err := validToken(cookie.Value)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !valid {
			userID := uuid.New()
			makeCookie(w, userID)
		}

		next.ServeHTTP(w, r)

	})

}

func makeCookie(w http.ResponseWriter, userID uuid.UUID) {
	token, err := buildToken(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cookie := &http.Cookie{
		Name:     "token",
		Value:    token,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
	}

	http.SetCookie(w, cookie)
}

func buildToken(userID uuid.UUID) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil

}

func validToken(tokenString string) (bool, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if t.Valid {
			return []byte(SecretKey), nil
		}
		return []byte(SecretKey), nil
	})
	if err != nil {
		return false, err
	}
	if token.Valid {
		return true, nil
	}

	return false, nil
}

func GetUserID(tokenString string) (uuid.UUID, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(SecretKey), nil
		})
	if err != nil {
		return uuid.UUID{}, err
	}

	if !token.Valid {
		return uuid.UUID{}, err
	}

	return claims.UserID, nil
}
