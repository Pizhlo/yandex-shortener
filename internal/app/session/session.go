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
		//var token string
		//var ok bool
		ownerValid := true

		cookie, err := r.Cookie("token")
		if err != nil {
			fmt.Println("CookieMiddleware r.Cookie err = ", err)
			ownerValid = false
			userID = uuid.New()
			makeCookie(w, userID)

		} else {
			userID, err = GetUserID(cookie.Value)
			if err != nil {
				fmt.Println("CookieMiddleware GetUserID err = ", err)
				// w.WriteHeader(http.StatusInternalServerError)
				// return
				ownerValid = false
				userID = uuid.New()
				makeCookie(w, userID)
			}

		}

		// valid, err := validToken(cookie.Value)
		// if err != nil {
		// 	fmt.Println("CookieMiddleware validToken err = ", err)
		// 	w.WriteHeader(http.StatusInternalServerError)
		// 	return
		// }

		// if !valid { // if there was cookie but invalid
		// 	userID := uuid.New()
		// 	makeCookie(w, userID)
		// }

		c := context.WithValue(context.WithValue(r.Context(), UserIDKey, userID), ValidOwnerKey, ownerValid)

		next.ServeHTTP(w, r.WithContext(c))

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
