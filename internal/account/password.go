package account

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type PasswordStatus string

const (
	PasswordUsable PasswordStatus = "P"
	PasswordLocked PasswordStatus = "L"
	PasswordUnset  PasswordStatus = "NP"
)

func (u *User) PasswordStatus() (PasswordStatus, error) {
	output, err := exec.Command(
		"passwd",
		"--status",
		u.Username,
	).Output()
	if err != nil {
		return "", fmt.Errorf(
			"check password status for %q: %w",
			u.Username,
			err,
		)
	}

	fields := strings.Fields(string(output))
	if len(fields) < 2 {
		return "", fmt.Errorf(
			"unexpected password status for %q",
			u.Username,
		)
	}

	status := PasswordStatus(fields[1])
	switch status {
	case PasswordUsable, PasswordLocked, PasswordUnset:
		return status, nil
	default:
		return "", fmt.Errorf(
			"unknown password status %q for %q",
			status,
			u.Username,
		)
	}
}

func (u *User) SetPassword(password []byte) error {
	if len(password) == 0 {
		return fmt.Errorf("password must not be empty")
	}

	if bytes.ContainsAny(password, "\r\n") {
		return fmt.Errorf("password must not contain a line break")
	}

	input := make([]byte, 0, len(u.Username)+1+len(password)+1)
	input = append(input, u.Username...)
	input = append(input, ':')
	input = append(input, password...)
	input = append(input, '\n')
	defer clear(input)

	command := exec.Command("chpasswd")
	command.Stdin = bytes.NewReader(input)

	if err := command.Run(); err != nil {
		return fmt.Errorf(
			"set password for %q: %w",
			u.Username,
			err,
		)
	}

	return nil
}
