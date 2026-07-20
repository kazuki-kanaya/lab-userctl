package register

import (
	"bytes"
	"fmt"
	"regexp"

	"github.com/kazuki-kanaya/lab-userctl/internal/account"
)

var usernamePattern = regexp.MustCompile(
	`^[a-z][a-z0-9_-]{0,31}$`,
)

func (s *Service) resolveUser() (*account.User, error) {
	username, err := s.prompt.Ask("Username: ")
	if err != nil {
		return nil, fmt.Errorf("read username: %w", err)
	}

	if !usernamePattern.MatchString(username) {
		return nil, fmt.Errorf("invalid username: %q", username)
	}

	user, err := account.Lookup(username)
	if err != nil {
		return nil, err
	}

	if user != nil {
		fmt.Printf("User already exists: %s\n", user.Username)
		return user, nil
	}

	password, err := s.prompt.AskPassword("Password: ")
	if err != nil {
		return nil, fmt.Errorf("read password: %w", err)
	}
	defer clear(password)

	confirmation, err := s.prompt.AskPassword("Confirm password: ")
	if err != nil {
		return nil, fmt.Errorf("read password confirmation: %w", err)
	}
	defer clear(confirmation)

	if !bytes.Equal(password, confirmation) {
		return nil, fmt.Errorf("passwords do not match")
	}

	user, err = account.Create(username, password)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Created user: %s\n", user.Username)
	return user, nil
}

func (s *Service) configureSudo(user *account.User) error {
	hasSudo, err := user.HasSudo()
	if err != nil {
		return err
	}

	if hasSudo {
		fmt.Println("Sudo access already configured")
		return nil
	}

	grantSudo, err := s.prompt.Confirm(
		"Grant sudo access?",
		false,
	)
	if err != nil {
		return fmt.Errorf("confirm sudo access: %w", err)
	}

	if !grantSudo {
		fmt.Println("Skipped sudo access")
		return nil
	}

	if err := user.GrantSudo(); err != nil {
		return err
	}

	fmt.Println("Granted sudo access")
	return nil
}
