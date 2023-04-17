package models

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

type UserModelInterface interface {
	Insert(name, email, password string) error
	Authenticate(email, password string) (int, error)
	Exists(id int) (bool, error)
	Get(id int) (*User, error)
	UpdatePassword(id int, oldPassword, newPassword string) error
}

type User struct {
	ID             int
	Name           string
	Email          string
	HashedPassword []byte
	Created        time.Time
}

type UserModel struct {
	DB *sql.DB
}

func (m *UserModel) Insert(name, password, email string) error {
	hashPass, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	stmt := `INSERT INTO users (name, email, hashed_password, created)
    VALUES(?, ?, ?, UTC_TIMESTAMP())`

	_, err = m.DB.Exec(stmt, name, email, hashPass)
	if err != nil {
		var mySQLError *mysql.MySQLError
		if errors.As(err, &mySQLError) {
			if mySQLError.Number == 1062 && strings.Contains(mySQLError.Message, "users_uc_email") {
				return ErrDuplicateEmail
			}
		}
		return err
	}

	return nil
}

func (m *UserModel) Authenticate(email, password string) (id int, err error) {
	var hashedPassword []byte

	stmt := `SELECT id, hashed_password FROM users WHERE email = ?`
	err = m.DB.QueryRow(stmt, email).Scan(&id, &hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	return id, nil
}

func (m *UserModel) Exists(id int) (exists bool, err error) {
	stmt := "SELECT EXISTS(SELECT true FROM users WHERE id = ?)"

	err = m.DB.QueryRow(stmt, id).Scan(&exists)

	return exists, err
}

func (m *UserModel) Get(id int) (*User, error) {
	stmt := "SELECT name, email, created FROM users WHERE id = ?"

	usr := &User{}
	err := m.DB.QueryRow(stmt, id).Scan(&usr.Name, &usr.Email, &usr.Created)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}

	return usr, err
}

func (m *UserModel) UpdatePassword(id int, oldPassword, newPassword string) error {
	fetchPassSTMT := "SELECT hashed_password FROM users WHERE id = ?"
	var pass []byte
	err := m.DB.QueryRow(fetchPassSTMT, id).Scan(&pass)
	if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword(pass, []byte(oldPassword))
	if err != nil {
		return ErrInvalidCredentials
	}

	hashPass, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return err
	}

	stmt := `UPDATE users SET hashed_password = ? WHERE id = ?`

	_, err = m.DB.Exec(stmt, hashPass, id)
	if err != nil {
		return err
	}

	return nil
}
