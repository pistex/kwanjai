package models

import (
	"github.com/google/uuid"
	"kwanjai/libraries"
	"log"
	"net/http"
	"strings"
	"time"
)

// User model.
type User struct {
	UID                string     `json:"uid" gorm:"primaryKey"`
	Username           string     `json:"username" gorm:"unique"`
	Email              string     `json:"email" gorm:"unique"`
	Firstname          string     `json:"firstname"`
	Lastname           string     `json:"lastname"`
	Password           string     `json:"password,omitempty"`
	IsSuperUser        bool       `json:"is_superuser"`
	IsVerified         bool       `json:"is_verified"`
	IsActive           bool       `json:"is_active"`
	JoinedDate         *time.Time `json:"joined_date"`
	Plan               string     `json:"plan"`
	Projects           int        `json:"projects"`
	CustomerID         string     `json:",omitempty"`
	SubscriptionID     string     `json:",omitempty"`
	DateOfSubscription int        `json:"date_of_subscription"`
}

// Register user method.
func (user *User) Register() (int, string, *User) {
	status, message, user := user.createUser()
	if status != http.StatusCreated || user == nil {
		return status, message, user
	}
	status, message = user.SendVerificationEmail()
	return status, message, user
}

// Finduser user method.
func (user *User) Finduser() (int, string, *User) {
	status, message, user := user.findUser()
	return status, message, user
}

func (user *User) findUser() (int, string, *User) {
	if user.Username == "" && user.Email == "" {
		return http.StatusNotFound, "User not found.", nil
	}
	getUser, err := libraries.FirestoreFind("users", user.Username)
	if err != nil {
		getEmail, err := libraries.FirestoreSearch("users", "Email", "==", user.Email)
		if err != nil {
			log.Panicln(err)
		}
		if len(getEmail) > 0 {
			_ = getEmail[0].DataTo(user)
			user.Password = ""
			return http.StatusOK, "Get user successfully.", user
		}
		return http.StatusNotFound, "User not found.", nil
	}
	err = getUser.DataTo(user)
	if err != nil {
		log.Panicln(err)
	}
	user.Password = ""
	return http.StatusOK, "Get user successfully.", user
}

func (user *User) createUser() (int, string, *User) {
	user.Username = strings.ToLower(user.Username)
	user.Email = strings.ToLower(user.Email)
	_, _, userFound := user.findUser()
	if userFound != nil {
		return http.StatusConflict, "Provided email or username is already registered.", nil
	}
	user.Initialize()
	_, err := libraries.FirestoreCreateOrSet("users", user.Username, user)
	if err != nil {
		log.Panicln(err)
	}
	user.Password = ""
	return http.StatusCreated, "User created successfully.", user
}

// SendVerificationEmail method for user model.
func (user *User) SendVerificationEmail() (int, string) {
	email := new(VerificationEmail)
	email.Initialize(user.Username, user.Email)
	reference, _, err := libraries.FirestoreAdd("verificationEmail", email)
	if err != nil {
		log.Panicln(err)
	}
	email.ID = reference.ID
	status, message := email.Send()
	return status, message
}

// HashPassword before register
func (user *User) HashPassword() error {
	hashedPassword, err := libraries.HashPassword(user.Password)
	user.Password = hashedPassword
	return err
}

func (user *User) Initialize() {
	user.UID = uuid.New().String()
	user.Plan = "Starter"
	user.IsSuperUser = false
	user.IsVerified = false
	now := time.Now()
	user.JoinedDate = &now
}

// MakeAnonymous user
func (user *User) MakeAnonymous() {
	user.Username = "anonymous"
	user.IsSuperUser = false
	user.IsVerified = false
	user.IsActive = false
	now := time.Now()
	user.JoinedDate = &now
}
