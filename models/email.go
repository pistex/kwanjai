package models

import (
	"fmt"
	"kwanjai/configuration"
	"kwanjai/libraries"
	"log"
	"math/rand"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// VerificationEmail model.
type VerificationEmail struct {
	User        string
	Email       string `json:"email"`
	Key         string `json:"key"`
	ID          string
	ExpiredDate time.Time
}

// Initialize email objects.
func (email *VerificationEmail) Initialize(user string, emailAddress string) {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	email.User = user
	email.Email = emailAddress
	email.Key = fmt.Sprintf("%06d", random.Intn(999999))
	email.ExpiredDate = time.Now().Add(configuration.EmailVerificationLifetime)
}

// Send method for VerificationEmail object.
func (email *VerificationEmail) Send() (int, string) {
	// Sender data.
	from := "surus.d6101@gmail.com"
	password := configuration.EmailServicePassword
	to := []string{email.Email}
	if strings.Contains(to[0], "@example.com") {
		log.Println("email contain @example.com")
		return http.StatusOK, "Email sent."
	}
	verificationLink := fmt.Sprintf("%v/verify_email/%v/", configuration.FrontendURL, email.ID)
	message := fmt.Sprintf("From: Kwanjai Admin <surus.d6101@gmail.com>\r\n"+
		"To: %v\r\n"+
		"Subject: Verification email.\r\n"+
		"\r\n"+
		"Hi %v.\r\n"+
		"Please verify your email using following link.\r\n"+
		"%v\r\n"+
		"Your verification code is: %v\r\n", to[0], email.User, verificationLink, email.Key)
	auth := smtp.PlainAuth("", from, password, "smtp.gmail.com")
	err := smtp.SendMail("smtp.gmail.com:587", auth, from, to, []byte(message))
	if err != nil {
		log.Panicln(err)
	}
	return http.StatusOK, "Email sent."
}

// Verify method for VerificationEmail object.
// The method set user to be verified if verification is completed.
// If the email is expired, the method delete the email in database.
func (email *VerificationEmail) Verify() (int, string) {
	if email.ID == "" {
		return http.StatusBadRequest, "Bad verification link."
	}
	getEmail, err := libraries.FirestoreFind("verificationEmail", email.ID)
	if !getEmail.Exists() {
		if err != nil {
			emailPath := getEmail.Ref.Path
			emailNotExist := status.Errorf(codes.NotFound, "%q not found", emailPath)
			if err.Error() == emailNotExist.Error() {
				return http.StatusBadRequest, "Bad verification link."
			}
			log.Panicln(err)
		}
	}
	verificationEmail := new(VerificationEmail)
	err = getEmail.DataTo(verificationEmail)
	now := time.Now()
	expriredDate := verificationEmail.ExpiredDate
	exprired := now.After(expriredDate)
	if exprired {
		_, err = libraries.FirestoreDelete("verificationEmail", email.ID)
		if err != nil {
			log.Panicln(err)
		}
		return http.StatusBadRequest, "Link is expired."
	}
	if email.Key == verificationEmail.Key {
		_, err = libraries.FirestoreUpdateField("users", verificationEmail.User, "IsVerified", true)
		if err != nil {
			log.Panicln(err)
		}
		_, err = libraries.FirestoreDelete("verificationEmail", email.ID)
		if err != nil {
			log.Panicln(err)
		}
		return http.StatusOK, "Email verified."
	}
	return http.StatusBadRequest, "Key is invalid."
}
