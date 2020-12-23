package models

import (
	"kwanjai/libraries"
	"log"
	"net/http"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LoginCredential info.
type LoginCredential struct {
	ID       string `json:"id" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LogoutData to track logout process event.
type LogoutData struct {
	RefreshPassed  bool
	AccessPassed   bool
	User           string
	AccessTokenID  string
	RefreshTokenID string
	AccessDeleted  bool
	RefreshDeleted bool
}

// Verify logout credential.
func (logout *LogoutData) Verify(tokenString string, tokenType string, passed chan bool, UUID chan string) {
	if tokenType == "access" {
		logout.AccessPassed, logout.User, logout.AccessTokenID, _ = libraries.VerifyToken(tokenString, "access")
		passed <- logout.AccessPassed
		UUID <- logout.AccessTokenID
	} else if tokenType == "refresh" {
		logout.RefreshPassed, logout.User, logout.RefreshTokenID, _ = libraries.VerifyToken(tokenString, "refresh")
		passed <- logout.RefreshPassed
		UUID <- logout.RefreshTokenID
	} else {
		return
	}
}

// Login method
func (login *LoginCredential) Login() (int, string) {
	hashedPassword := ""
	username := ""
	login.ID = strings.ToLower(login.ID)
	getUser, err := libraries.FirestoreFind("users", login.ID)
	if err != nil {
		if getUser == nil {
			log.Panicln(err)
		}
		userPath := getUser.Ref.Path
		userNotExist := status.Errorf(codes.NotFound, "%q not found", userPath)
		if err.Error() == userNotExist.Error() {
			getEmail, err := libraries.FirestoreSearch("users", "Email", "==", login.ID)
			if err != nil {
				log.Panicln(err)
			}
			if len(getEmail) > 0 {
				hashedPassword = getEmail[0].Data()["HashedPassword"].(string)
				username = getEmail[0].Data()["Username"].(string)
			} else {
				return http.StatusBadRequest, "Cannot login with provided credential."
			}
		} else {
			log.Panicln(err)
		}
	} else {
		hashedPassword = getUser.Data()["HashedPassword"].(string)
		username = getUser.Data()["Username"].(string)
	}
	passwordPass := libraries.CheckPasswordHash(login.Password, hashedPassword)
	if !passwordPass {
		return http.StatusBadRequest, "Cannot login with provided credential."
	}
	_, err = libraries.FirestoreUpdateField("users", username, "IsActive", true)
	if err != nil {
		log.Panicln(err)
	}
	return http.StatusOK, username
}
