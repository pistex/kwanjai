package forms

type RegistrationForm struct {
	Username           string    `json:"username" binding:"required,ne=anonymous"`
	Password           string    `json:"password" binding:"required,min=8"`
	Email              string    `json:"email" binding:"required,email"`
}
