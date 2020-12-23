package controllers

import (
	"kwanjai/forms"
	"kwanjai/helpers"
	"kwanjai/libraries"
	"kwanjai/models"
	"kwanjai/services"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Login endpoint
func Login() gin.HandlerFunc {
	return func(ginContext *gin.Context) {
		login := new(forms.LoginForm)
		err := ginContext.ShouldBindJSON(login)
		if err != nil {
			ginContext.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		ginContext.JSON(200, gin.H{"message": "Logged in successfully"})
	}
}

// Register
func Register() gin.HandlerFunc {
	return func(ginContext *gin.Context) {
		registerForm := new(forms.RegistrationForm)
		err := ginContext.ShouldBindJSON(registerForm)
		if err != nil {
			ginContext.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
		authenticationService := services.NewAuthenticationService(ginContext)
		err = authenticationService.RegisterUser(registerForm)
		if err != nil {
			if strings.HasPrefix(err.Error(), "Error 1062: Duplicate entry"){
				ginContext.JSON(http.StatusBadRequest, gin.H{"message": "this username is already registered."})
				return
			}
			ginContext.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
		ginContext.JSON(201, gin.H{
			"message": "user registered",
		})

	}
}

// Logout endpoint
func Logout() gin.HandlerFunc {
	return func(ginContext *gin.Context) {
		username := helpers.GetUsername(ginContext)
		logout := new(models.LogoutData)
		token := new(libraries.Token)
		_ = ginContext.ShouldBindJSON(token)
		extractedToken := strings.Split(ginContext.Request.Header.Get("Authorization"), "Bearer ")
		if len(extractedToken) != 2 {
			token.AccessToken = ""
		} else {
			token.AccessToken = extractedToken[1]
		}
		if token.RefreshToken == "" {
			ginContext.JSON(http.StatusBadRequest, gin.H{"message": "No refresh token proved."})
			return
		}
		accessPassed := make(chan bool)
		accessTokenID := make(chan string)
		refreshPassed := make(chan bool)
		refreshTokenID := make(chan string)
		go logout.Verify(token.AccessToken, "access", accessPassed, accessTokenID)
		go logout.Verify(token.RefreshToken, "refresh", refreshPassed, refreshTokenID)
		passed := true == <-accessPassed && true == <-refreshPassed
		if !passed {
			ginContext.JSON(http.StatusUnauthorized, gin.H{"message": "Token verification failed."})
			return
		}
		_, _ = libraries.FirestoreDelete("tokens", <-accessTokenID)
		_, _ = libraries.FirestoreDelete("tokens", <-refreshTokenID)
		tokenSearch, _ := libraries.FirestoreSearch("tokens", "user", "==", username)
		if len(tokenSearch) == 0 {
			_, err := libraries.FirestoreUpdateField("users", username, "IsActive", false)
			if err != nil {
				log.Panicln(err)
			}
		}
		ginContext.JSON(200, gin.H{"message": "User logged out successfully."})
	}
}

// RefreshToken endpoint
func RefreshToken() gin.HandlerFunc {
	return func(ginContext *gin.Context) {
		token := new(libraries.Token)
		_ = ginContext.ShouldBind(token)
		extractedToken := strings.Split(ginContext.Request.Header.Get("Authorization"), "Bearer ")
		if len(extractedToken) != 2 {
			token.AccessToken = ""
		} else {
			token.AccessToken = extractedToken[1]
		}
		if token.RefreshToken == "" {
			ginContext.JSON(http.StatusBadRequest, gin.H{"message": "No refresh token provided."})
			return
		}
		_, refreshUsername, _, err := libraries.VerifyToken(token.RefreshToken, "refresh")
		if err != nil {
			if refreshUsername == "anonymous" {
				ginContext.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
				return
			}
			log.Panicln(err)
		}
		_, accessUsername, tokenID, err := libraries.VerifyToken(token.AccessToken, "access") // if token is expired here, it's got delete.
		if accessUsername != "anonymous" && err == nil {                                      // user != "anonymous" means token is still valid.
			_, err = libraries.FirestoreDelete("tokens", tokenID)
			if err != nil {
				log.Panicln(err)
			}
		}
		newToken, _, err := libraries.CreateToken("access", refreshUsername)
		token.AccessToken = newToken
		ginContext.JSON(http.StatusOK, gin.H{
			"token": token,
		})
	}
}

// TokenVerification endpoint
func TokenVerification() gin.HandlerFunc {
	return func(ginContext *gin.Context) {
		ginContext.Status(http.StatusOK)
	}
}

type passwordUpdate struct {
	OldPassword  string `json:"old_password" binding:"required,min=8"`
	NewPassword1 string `json:"new_password1" binding:"required,min=8"`
	NewPassword2 string `json:"new_password2" binding:"required,min=8"`
}

// PasswordUpdate endpoint
func PasswordUpdate() gin.HandlerFunc {
	return func(ginContext *gin.Context) {
		passwordForm := new(passwordUpdate)
		if err := ginContext.ShouldBindJSON(passwordForm); err != nil {
			ginContext.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		}
		if passwordForm.NewPassword1 != passwordForm.NewPassword2 {
			ginContext.JSON(http.StatusBadRequest, gin.H{"message": "Password confirmation failed."})
		}
		username := helpers.GetUsername(ginContext)
		newPassword, _ := libraries.HashPassword(passwordForm.NewPassword1)
		_, _ = libraries.FirestoreUpdateField("users", username, "HashedPassword", newPassword)
		ginContext.JSON(http.StatusOK, gin.H{"message": "Password updated."})
	}
}
