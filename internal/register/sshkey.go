package register

import (
	"fmt"

	"github.com/kazuki-kanaya/lab-userctl/internal/account"
	"github.com/kazuki-kanaya/lab-userctl/internal/sshkey"
)

func (s *Service) registerSSHKey(user *account.User) error {
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
