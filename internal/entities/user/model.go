package user

type User struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
	BirthDate    string `json:"birth_date"`
	Height       string `json:"height,omitempty"`
}
