package client

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"net/http"
	"time"
)

const authCookieName = "auth_shortener"

// AuthToken is interface for client token authentication
type AuthToken interface {
	Create() (string, error)
	Verify(tokenString string) (string, error)
	GetUserID(r *http.Request) string
}

type authToken struct {
	secretKey []byte
}

func NewAuthToken(key []byte) AuthToken {
	return &authToken{secretKey: key}
}

// Create creates a new jwt token with client uuid
func (a *authToken) Create() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"user_id": uuid.New().String(),
			"exp":     time.Now().Add(24 * time.Hour).Unix(),
		})

	return token.SignedString(a.secretKey)
}

// Verify checks token and return client uuid if the token is valid
func (a *authToken) Verify(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return a.secretKey, nil
	})

	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", errors.New("token is not valid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("can not get claims")
	}

	uid, ok := claims["user_id"].(string)
	if !ok {
		return "", errors.New("user_id is not exist")
	}

	return uid, nil
}

// GetUserID return user ID from auth cookie
func (a *authToken) GetUserID(r *http.Request) string {
	cookie, err := r.Cookie(authCookieName)
	if err != nil {
		return ""
	}

	userID, err := a.Verify(cookie.Value)
	if err != nil {
		return ""
	}

	return userID
}
