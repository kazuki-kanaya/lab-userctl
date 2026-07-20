package register

import (
	"bytes"
	"fmt"
	"regexp"

	"github.com/kazuki-kanaya/lab-userctl/internal/account"
	"github.com/kazuki-kanaya/lab-userctl/internal/sshkey"
	"github.com/kazuki-kanaya/lab-userctl/internal/terminal"
)

type Service struct {
	prompt *terminal.Prompter
}

func New(prompt *terminal.Prompter) *Service {
	return &Service{
		prompt: prompt,
	}
}

var usernamePattern = regexp.MustCompile(
	`^[a-z][a-z0-9_-]{0,31}$`,
)

func (s *Service) Run() error {
	username, err := s.prompt.Ask("")
	if err != nil {
		return fmt.Errorf("read username: %w", err)
	}

	if !usernamePattern.MatchString(username) {
		return fmt.Errorf("invalid username: %q", username)
	}

	user, err := account.Lookup(username)
	if err != nil {
		return err
	}

	if user == nil {
		password, err := s.prompt.AskPassword("Password: ")
		if err != nil {
			return fmt.Errorf("read password: %w", err)
		}
		defer clear(password)

		confirmation, err := s.prompt.AskPassword("Confirm password: ")
		if err != nil {
			return fmt.Errorf("read password confirmation: %w", err)
		}
		defer clear(confirmation)

		if !bytes.Equal(password, confirmation) {
			return fmt.Errorf("passwords do not match")
		}

		user, err = account.Create(username, password)
		if err != nil {
			return err
		}

		fmt.Printf("Created user: %s\n", user.Username)
	} else {
		fmt.Printf("User already exists: %s\n", user.Username)
	}

	hasSudo, err := user.HasSudo()
	if err != nil {
		return err
	}

	if !hasSudo {
		if err := user.GrantSudo(); err != nil {
			return err
		}

		fmt.Println("Granted sudo access")
	} else {
		fmt.Println("Sudo access already configured")
	}

	registerSSHKey, err := s.prompt.Confirm(
		"Register an SSH public key?",
		true,
	)
	if err != nil {
		return fmt.Errorf("confirm SSH public key registration: %w", err)
	}

	if !registerSSHKey {
		return nil
	}

	input, err := s.prompt.Ask("SSH public key: ")
	if err != nil {
		return fmt.Errorf("read SSH public key: %w", err)
	}

	key, err := sshkey.Parse(input)
	if err != nil {
		return err
	}

	added, err := sshkey.Add(user, key)
	if err != nil {
		return err
	}

	if added {
		fmt.Println("Registered SSH public key")
	} else {
		fmt.Println("SSH public key already registered")
	}

	return nil
}
