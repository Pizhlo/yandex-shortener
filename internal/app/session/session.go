package session

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

const TokenExp = time.Hour * 3
const SecretKey = "supersecretkey"

type ContextParamName string

const UserIDKey ContextParamName = "UserIDKey"
const ValidOwnerKey ContextParamName = "ValidOwnerKey"

type Claims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID
}

// CookieMiddleware проверяет куки на валидность; если проверка не пройдена - создает новую куку.
func CookieMiddleware(next http.Handler) http.Handler {
	fmt.Println("CookieMiddleware")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var userID uuid.UUID
		var token string
		var ok bool
		ownerValid := true

		cookie, err := r.Cookie("token")
		if err != nil {
			fmt.Println("CookieMiddleware r.Cookie err = ", err)
			ownerValid = false
			userID := uuid.New()
			token, err = makeToken(userID)
			if err != nil {
				fmt.Println("CookieMiddleware makeToken err = ", err)
			}
		} else if userID, ok = GetUserID(cookie.Value); !ok {
			ownerValid = false
			userID := uuid.New()
			token, err = makeToken(userID)
			if err != nil {
				fmt.Println("CookieMiddleware makeToken !ok err = ", err)
			}
		}

		if token != "" {
			http.SetCookie(w, &http.Cookie{Name: "token", Value: token, HttpOnly: true})
		}

		c := context.WithValue(context.WithValue(r.Context(), UserIDKey, userID), ValidOwnerKey, ownerValid)

		next.ServeHTTP(w, r.WithContext(c))

	})

}

func GetUserID(tokenString string) (uuid.UUID, bool) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(SecretKey), nil
		})

	if err != nil {
		fmt.Println("CookieMiddleware getOwnerID ParseWithClaims err = ", err)
		return uuid.UUID{}, false
	}

	if !token.Valid {
		fmt.Println("CookieMiddleware getOwnerID token is invalid")
		return uuid.UUID{}, false
	}

	// возвращаем ID пользователя в читаемом виде
	return claims.UserID, true
}

func makeToken(userID uuid.UUID) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
