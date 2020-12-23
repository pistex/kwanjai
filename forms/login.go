package forms

type LoginForm struct {
	ID       string `json:"id" binding:"required,ne=anonymous"`
	Password string `json:"password" binding:"required,min=8"`
}
