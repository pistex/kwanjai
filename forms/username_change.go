package forms

type UsernameChangeForm struct {
	OldUsername           string    `json:"old_username" binding:"required"`
	NewUsername           string    `json:"new_username" binding:"required,ne=anonymous"`
}
