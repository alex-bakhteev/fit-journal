package user

type CreateUserDTO struct {
	Username  string `json:"username,omitempty"`
	Password  string `json:"password,omitempty"`
	BirthDate string `json:"birth_date"`
	Height    string `json:"height,omitempty"`
}
