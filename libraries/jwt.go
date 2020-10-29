package libraries

import (
	"errors"
	"fmt"
	"kwanjai/config"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Token object
type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type tokenStatus struct {
	AccessToken  bool
	RefreshToken bool
}

type customClaims struct {
	User string `json:"user"`
	UUID string `json:"uuid,omitempty"`
	*jwt.StandardClaims
}

func getSecretKeyAndLifetime(tokenType string) (string, time.Duration, error) {
	if tokenType == "access" {
		return config.JWTAccessTokenSecretKey, config.JWTAccessTokenLifetime, nil
	} else if tokenType == "refresh" {
		return config.JWTRefreshTokenSecretKey, config.JWTRefreshTokenLifetime, nil
	} else {
		err := errors.New("no token type provide")
		return "", time.Second, err
	}
}
func createToken(tokenType string, username string) (string, error) {
	secretKey, lifetime, err := getSecretKeyAndLifetime(tokenType)
	var tokenUUID string
	if err != nil {
		return "no token created.", err
	}

	tokenUUID = uuid.New().String()
	firestoreClient, err := FirebaseApp().Firestore(config.Context)
	defer firestoreClient.Close()
	if err != nil {
		return "Cannot create token uuid.", err
	}
	getUserToken, err := firestoreClient.Collection("tokenUUID").Doc(username).Get(config.Context)
	if !getUserToken.Exists() {
		_, err = firestoreClient.Collection("tokenUUID").Doc(username).Set(config.Context, map[string]interface{}{
			tokenUUID: tokenType,
		})
	}
	_, err = firestoreClient.Collection("tokenUUID").Doc(username).Update(config.Context, []firestore.Update{
		{
			Path:  tokenUUID,
			Value: tokenType,
		},
	})
	if err != nil {
		return "Cannot create token uuid.", err
	}

	claims := &customClaims{
		username,
		tokenUUID,
		&jwt.StandardClaims{
			ExpiresAt: time.Now().Add(lifetime).Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secretKey))
	return signedToken, err
}

// Initialize token
func (token *Token) Initialize(username string) (int, string) {
	var passed bool
	tokenStatus := new(tokenStatus)
	go tokenStatus.createToken(username, "access", token)
	go tokenStatus.createToken(username, "refresh", token)
	timeout := time.Now().Add(time.Second * 10)
	timer := time.Now()
	for !passed && !timer.Equal(timeout) {
		passed = tokenStatus.AccessToken == true && tokenStatus.RefreshToken == true
		timer = time.Now()
	}
	if !passed {
		return http.StatusInternalServerError, "create token error"
	}
	return http.StatusOK, "Token issued."
}

func (tokenStatus *tokenStatus) createToken(username string, tokenType string, token *Token) {
	var err error
	if tokenType == "access" {
		token.AccessToken, err = createToken("access", username)
		tokenStatus.AccessToken = err == nil
	} else if tokenType == "refresh" {
		token.RefreshToken, err = createToken("refresh", username)
		tokenStatus.RefreshToken = err == nil
	}
}

// VerifyToken with a particular type.
func VerifyToken(tokenString string, tokenType string) (bool, string, string, error) {
	secretKey, _, err := getSecretKeyAndLifetime(tokenType)
	if err != nil {
		return false, "", "", err
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return false, "anonymous", "", err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false, "anonymous", "", errors.New("claim failed")
	}
	tokenUUID, username := claims["uuid"].(string), claims["user"].(string)
	firestoreClient, ferr := FirebaseApp().Firestore(config.Context)
	defer firestoreClient.Close()
	if ferr != nil {
		return false, "", "", ferr
	}
	uuidVerification, ferr := firestoreClient.Collection("tokenUUID").Doc(username).Get(config.Context)
	if ferr != nil {
		tokenPath := uuidVerification.Ref.Path
		tokenNotExist := status.Errorf(codes.NotFound, "%q not found", tokenPath)
		if ferr.Error() == tokenNotExist.Error() {
			return false, "", "", errors.New("user not found")
		}
		return false, "", "", ferr
	}
	if uuidVerification.Data()[tokenUUID] == nil {
		return false, "anonymous", "", errors.New("token is not valid")
	}
	if err != nil {
		if err.Error() == "Token is expired" {
			ferr = DeleteToken(username, tokenUUID)
			if ferr != nil {
				return false, "anonymous", "", ferr
			}
		}
	}
	if ok && token.Valid {
		return true, username, tokenUUID, nil
	}
	return false, "anonymous", "", err
}

// DeleteToken by username and token uuid
func DeleteToken(username string, tokenUUID string) error {
	firestoreClient, err := FirebaseApp().Firestore(config.Context)
	defer firestoreClient.Close()
	_, err = firestoreClient.Collection("tokenUUID").Doc(username).Update(config.Context, []firestore.Update{
		{
			Path:  tokenUUID,
			Value: firestore.Delete,
		},
	})
	if err != nil {
		return err
	}
	return err
}
