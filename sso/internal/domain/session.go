package domain

type AuthSession struct {
	AccessToken string
	UserID      int64
	Login       string
	Role        string
}

type AuthCredentials struct {
	Identifier string
	Password   string
	Type       string
}

type OTPConfirm struct {
	Destination string
	Type        string
	Code        int64
}

type PasswordResetConfirm struct {
	Destination string
	Type        string
	Code        int64
	NewPassword string
}

type CreateUser struct {
	Email    string
	Password string
	Name     string
	LastName string
	Login    string
}
