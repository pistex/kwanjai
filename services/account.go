package services

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"kwanjai/forms"
	"kwanjai/models"
)

type AccountServiceInterface interface {
	ChangeUsername(form *forms.UsernameChangeForm) error
	FindByUsernameOrEmail(id string, user *models.User) error
}

type accountService struct {
	ginContext *gin.Context
}

func NewAccountService(ginContext *gin.Context) AccountServiceInterface {
	return &accountService{ginContext: ginContext}
}

func (s *accountService) ChangeUsername(form *forms.UsernameChangeForm) error {
	if db, exist := s.ginContext.Get("db"); exist {
		err := db.(*gorm.DB).Model(&models.User{}).Where("username = ?", form.OldUsername).Update("username", form.NewUsername).Error
		return err
	} else {
		return errors.New("no database set in gin context")
	}
}

func (s *accountService) FindByUsernameOrEmail(id string, user *models.User) error {
	if db, exist := s.ginContext.Get("db"); exist {
		err := db.(*gorm.DB).
			Where("username = ?", id).Or("email = ?", id).Take(user).Error
		return err
	} else {
		return errors.New("no database set in gin context")
	}
}
