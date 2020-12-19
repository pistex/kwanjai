package services

import (
	"kwanjai/configuration"
	"kwanjai/forms"
	"kwanjai/models"
)

type AuthenticationServiceInterface interface {
	RegisterUser(form *forms.RegistrationForm) error
}

type authenticationService struct {

}

func NewAuthenticationService() AuthenticationServiceInterface {
	return &authenticationService{}
}

func (s *authenticationService) RegisterUser(form *forms.RegistrationForm) error {
	var err error
	user := new(models.User)
	user.Username = form.Username
	user.Password = form.Password
	user.Email = form.Email
	user.Initialize()
	if err = user.HashPassword(); err != nil {
		return err
	}
	err = configuration.SQL.Create(user).Error
	return err
}
