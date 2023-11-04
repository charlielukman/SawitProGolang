package entities

const (
	PhoneNumberMinLength = 10
	PhoneNumberMaxLength = 13
	FullNameMinLength    = 3
	FullNameMaxLength    = 60
	PasswordMinLength    = 6
	PasswordMaxLength    = 64
	PhoneNumberPrefix    = "+62"
)

type User struct {
	ID          int
	FullName    string
	PhoneNumber string
	Password    string
}
