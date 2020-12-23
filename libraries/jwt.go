package libraries

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"kwanjai/configuration"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// Token object.
type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type customClaims struct {
	User string `json:"user"`
	UUID string `json:"uuid"`
	*jwt.StandardClaims
}

// getSecretKeyAndLifetime function returns secret key (string), token lifetime (time.Duration), error.
func getSecretKeyAndLifetime(tokenType string) (string, time.Duration, error) {
	if tokenType == "access" {
		return configuration.JWTAccessTokenSecretKey, configuration.JWTAccessTokenLifetime, nil
	} else if tokenType == "refresh" {
		return configuration.JWTRefreshTokenSecretKey, configuration.JWTRefreshTokenLifetime, nil
	} else {
		err := errors.New("no token type provide")
		return "", time.Second, err
	}
}

// GetTokenPayload function returns payload value (string), token validation status (bool), error.
func GetTokenPayload(tokenString string, tokenType string, field string) (string, bool, error) {
	secretKey, _, err := getSecretKeyAndLifetime(tokenType)
	if err != nil {
		return "", false, err
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})
	if token == nil { // if tokenString is not token, jwt.Parse return nil object.
		return "", false, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", false, errors.New("claim failed")
	}
	var payload string
	if claims[field] != nil {
		payload = claims[field].(string)
	} else {
		payload = ""
	}
	return payload, token.Valid, err // if tokenString is a token but it is not valid, it return token object with token.Valid = false.
}

// CreateToken returns token (string) and error.
func CreateToken(tokenType string, username string) (string, *customClaims, error) {
	secretKey, lifetime, err := getSecretKeyAndLifetime(tokenType)
	if err != nil {
		return "no token created.", nil, err
	}
	claims := &customClaims{
		username,
		uuid.New().String(),
		&jwt.StandardClaims{
			ExpiresAt: time.Now().Add(lifetime).Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secretKey))
	return signedToken, claims, err
}

// VerifyToken function returns token validation status (bool), username (string), token UUID (string), error.
func VerifyToken(tokenString string, tokenType string) (bool, string, string, error) {
	tokenID, valid, err := GetTokenPayload(tokenString, tokenType, "id")
	if !valid || err != nil {
		if err != nil {
			if err.Error() == "Token is expired" {
				// Todo: delete token here.
			}
			return false, "anonymous", "", err
		}
	}
	username, _, _ := GetTokenPayload(tokenString, tokenType, "user")
	// passing first verification ensure no error here.
	return valid, username, tokenID, nil

}
