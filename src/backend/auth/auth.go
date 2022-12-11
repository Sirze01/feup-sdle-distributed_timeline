package auth

import (
	"errors"

	"git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/core/dht"
	"git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/core/dht/record/account"
	"golang.org/x/crypto/bcrypt"
)

var ErrUsernameAlreadyExists = errors.New("register: username already exists")
var ErrUsernameDoesntExist = errors.New("login: username doesn't exist")
var ErrIncorrectPassword = errors.New("login: incorrect password")

func Login(dht dht.DHT, username string, password string) error {
	usernameExists, err := dht.KeyExists(account.AccountNS + username)

	if err != nil {
		return err
	}

	if !usernameExists {
		return ErrUsernameDoesntExist
	}

	hashedPassword, err := dht.GetValue(account.AccountNS + username)

	if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))

	if err == bcrypt.ErrMismatchedHashAndPassword {
		return ErrIncorrectPassword
	} else if err != nil {
		return err
	}

	return nil
}

func Register(dht dht.DHT, username string, password string) error {
	usernameAlreadyExists, err := dht.KeyExists(account.AccountNS + username)

	if err != nil {
		return err
	}

	if usernameAlreadyExists {
		return ErrUsernameAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	_, err = dht.PutValue(account.AccountNS+username, hashedPassword)

	if err != nil {
		return err
	}

	return nil
}
