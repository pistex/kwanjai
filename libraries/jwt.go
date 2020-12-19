package libraries

import (
	"errors"
	"fmt"
	"kwanjai/configuration"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// Token object.
type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Use for tracking token issue status.
type tokenStatus struct {
	AccessToken  bool
	RefreshToken bool
}

type customClaims struct {
	User string `json:"user"`
	ID   string `json:"id"`
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

// GetTokenPayload function returns payload value (string), token validiation status (bool), error.
func GetTokenPayload(tokenString string, tokenType string, field string) (string, bool, error) {
	secretKey, _, err := getSecretKeyAndLifetime(tokenType)
	if err != nil {
		return "", false, err
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
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
func CreateToken(tokenType string, username string) (string, error) {
	secretKey, lifetime, err := getSecretKeyAndLifetime(tokenType)
	if err != nil {
		return "no token created.", err
	}

	now := time.Now().Truncate(time.Millisecond)

	reference, _, err := FirestoreAdd("tokens",
		map[string]interface{}{
			"user":   username,
			"expire": now,
		})
	if err != nil {
		return "Cannot store token data.", err
	}
	claims := &customClaims{
		username,
		reference.ID,
		&jwt.StandardClaims{
			ExpiresAt: now.Add(lifetime).Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secretKey))
	return signedToken, err
}

// Initialize token method.
func (token *Token) Initialize(username string) (int, string) {

	tokenStatus := new(tokenStatus)
	accessPassed := make(chan bool)
	refreshPassed := make(chan bool)
	go tokenStatus.createToken(username, "access", token, accessPassed)
	go tokenStatus.createToken(username, "refresh", token, refreshPassed)
	passed := true == <-accessPassed && true == <-refreshPassed
	if !passed {
		return http.StatusInternalServerError, "create token error"
	}
	return http.StatusOK, "Token issued."
}

func (tokenStatus *tokenStatus) createToken(username string, tokenType string, token *Token, passed chan bool) {
	var err error
	if tokenType == "access" {
		token.AccessToken, err = CreateToken("access", username)
		passed <- err == nil
	} else if tokenType == "refresh" {
		token.RefreshToken, err = CreateToken("refresh", username)
		passed <- err == nil
	}
}

// VerifyToken function returns token validiation status (bool), username (string), token UUID (string), error.
func VerifyToken(tokenString string, tokenType string) (bool, string, string, error) {
	tokenID, _, _ := GetTokenPayload(tokenString, tokenType, "id")
	username, valid, err := GetTokenPayload(tokenString, tokenType, "user")
	if err != nil {
		if err.Error() == "Token is expired" {
			FirestoreDelete("tokens", tokenID)
		}
		return false, "anonymous", "", err
	}
	tokenVerification, _ := FirestoreFind("tokens", tokenID)
	if !tokenVerification.Exists() {
		return false, "anonymous", "", errors.New("token does not exist in database")
	}
	if valid {
		return true, username, tokenID, nil
	}
	return false, "anonymous", "", err
}
