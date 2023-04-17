package mocks

import "snippetbox.jmorelli.dev/internal/models"

type UserModel struct{}

func (m *UserModel) Insert(name, password, email string) error {
	switch email {
	case "dupe@example.com":
		return models.ErrDuplicateEmail
	default:
		return nil
	}
}

func (m *UserModel) Authenticate(email, password string) (int, error) {
	if email == "jay@email.com" && password == "12345678" {
		return 1, nil
	}

	return 0, models.ErrInvalidCredentials
}

func (m *UserModel) Exists(id int) (bool, error) {
	switch id {
	case 1:
		return true, nil
	default:
		return false, nil
	}
}

func (m *UserModel) Get(id int) (*models.User, error) {
	switch id {
	case 1:
		return &models.User{Name: "John"}, nil
	default:
		return nil, nil
	}
}

func (m *UserModel) UpdatePassword(id int, oldPassword, newPassword string) error {
	if oldPassword == newPassword {
		return nil
	}
	return models.ErrInvalidCredentials
}
