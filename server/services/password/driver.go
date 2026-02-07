package password

// PasswordSvcDriver defines the interface for password operations
type PasswordSvcDriver interface {
	HashPassword(password string) (string, error)
	VerifyPassword(hashedPassword, password string) error
}

// PasswordSvc implements PasswordSvcDriver interface
type PasswordSvc struct{}

// NewPasswordSvc creates a new password service instance
func NewPasswordSvc() PasswordSvcDriver {
	return &PasswordSvc{}
}
