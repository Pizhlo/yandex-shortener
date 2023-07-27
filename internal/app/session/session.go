package session

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

const TokenExp = time.Hour * 3
const SecretKey = "supersecretkey"
const UserIDKey = "userID"

type Claims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID
}

// CookieMiddleware проверяет куки на валидность; если проверка не пройдена - создает новую куку.
func CookieMiddleware(next http.Handler) http.Handler {
	fmt.Println("CookieMiddleware")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) { // if there's no cookie
				userID := uuid.New()
				//makeCookie(w, userID)
				ctx := context.WithValue(context.WithValue(r.Context(), UserIDKey, userID), UserIDKey, userID)
				fmt.Println("CookieMiddleware ctx value = ", ctx.Value(UserIDKey))
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				fmt.Println("CookieMiddleware r.Cookie err = ", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else { // if there's any cookie

			valid, err := validToken(cookie.Value)
			if err != nil {
				fmt.Println("CookieMiddleware validToken err = ", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if !valid { // if there was cookie but invalid
				userID := uuid.New()
				makeCookie(w, userID)
			}

			next.ServeHTTP(w, r)
		}

	})

}

func makeCookie(w http.ResponseWriter, userID uuid.UUID) {
	fmt.Println("makeCookie")
	token, err := buildToken(userID)
	if err != nil {
		fmt.Println("makeCookie buildToken err = ", err)
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
	fmt.Println("buildToken")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		fmt.Println("buildToken SignedString err = ", err)
		return "", err
	}

	return tokenString, nil

}

func validToken(tokenString string) (bool, error) {
	fmt.Println("validToken")
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if t.Valid {
			return []byte(SecretKey), nil
		}
		return []byte(SecretKey), nil
	})
	if err != nil {
		fmt.Println("validToken Parse err = ", err)
		return false, err
	}
	if token.Valid {
		return true, nil
	}

	return false, nil
}

func GetUserID(tokenString string) (uuid.UUID, error) {
	fmt.Println("GetUserID")
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(SecretKey), nil
		})
	if err != nil {
		fmt.Println("GetUserID ParseWithClaims err = ", err)
		return uuid.UUID{}, err
	}

	if !token.Valid {
		return uuid.UUID{}, nil
	}

	return claims.UserID, nil
}
