package services

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"kwanjai/forms"
	"kwanjai/libraries"
	"kwanjai/models"
	"log"
	"strings"
)

type AuthenticationServiceInterface interface {
	RegisterUser(form *forms.RegistrationForm) error
}

type authenticationService struct {
	ginContext *gin.Context
}

func NewAuthenticationService(ginContext *gin.Context) AuthenticationServiceInterface {
	return &authenticationService{ginContext: ginContext}
}

func (s *authenticationService) RegisterUser(form *forms.RegistrationForm) error {
	var err error
	user := new(models.User)
	user.Username = strings.ToLower(form.Username)
	user.Password = form.Password
	user.Email = strings.ToLower(form.Email)
	user.Initialize()
	if err = user.HashPassword(); err != nil {
		return err
	}
	if db, exist := s.ginContext.Get("db"); exist {
		err = db.(*gorm.DB).Create(user).Error
		return err
	} else {
		return errors.New("no database set in gin context")
	}

}

func (s *authenticationService) Login(form *forms.LoginForm) (*libraries.Token, error) {
	user := new(models.User)
	form.ID = strings.ToLower(form.ID)
	accountService := NewAccountService(s.ginContext)
	err := accountService.FindByUsernameOrEmail(form.ID, user)
	if err != nil {
		return nil, err
	}
	passed := libraries.CheckPasswordHash(form.Password, user.Password)
	if !passed {
		return nil, errors.New("password verification failed")
	}
	token := new(libraries.Token)
	accessToken, _, err := libraries.CreateToken("access", user.Username)
	if err != nil {
		log.Panic("generate token failed")
	}
	refreshToken, _, err := libraries.CreateToken("refresh", user.Username)
	if err != nil {
		log.Panic("generate token failed")
	}

	token.AccessToken = accessToken
	token.RefreshToken = refreshToken
	return token, nil
}
