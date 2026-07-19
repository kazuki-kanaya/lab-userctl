package account

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"slices"
	"strconv"
	"strings"
)

type User struct {
	Username string
	HomeDir  string
	UID      int
	GID      int
}

func Lookup(username string) (*User, error) {
	output, err := exec.Command(
		"getent",
		"passwd",
		username,
	).Output()

	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 2 {
			return nil, nil
		}

		return nil, fmt.Errorf(
			"look up user %q: %w",
			username,
			err,
		)
	}

	fields := strings.Split(
		strings.TrimRight(string(output), "\r\n"),
		":",
	)

	if len(fields) != 7 {
		return nil, fmt.Errorf(
			"unexpected passwd entry for %q",
			username,
		)
	}

	uid, err := strconv.Atoi(fields[2])
	if err != nil {
		return nil, fmt.Errorf(
			"parse UID for %q: %w",
			username,
			err,
		)
	}

	gid, err := strconv.Atoi(fields[3])
	if err != nil {
		return nil, fmt.Errorf(
			"parse GID for %q: %w",
			username,
			err,
		)
	}

	return &User{
		Username: fields[0],
		HomeDir:  fields[5],
		UID:      uid,
		GID:      gid,
	}, nil
}

func (u *User) HasSudo() (bool, error) {
	output, err := exec.Command(
		"id",
		"-nG",
		u.Username,
	).Output()

	if err != nil {
		return false, fmt.Errorf(
			"list groups for %q: %w",
			u.Username,
			err,
		)
	}

	return slices.Contains(
		strings.Fields(string(output)),
		"sudo",
	), nil
}

func Create(username string, password []byte) (*User, error) {
	err := exec.Command(
		"useradd",
		"--create-home",
		"--shell", "/bin/bash",
		"--",
		username,
	).Run()
	if err != nil {
		return nil, fmt.Errorf(
			"create user %q: %w",
			username,
			err,
		)
	}

	user, err := Lookup(username)
	if err != nil {
		return nil, fmt.Errorf(
			"look up newly created user %q: %w",
			username,
			err,
		)
	}

	if user == nil {
		return nil, fmt.Errorf(
			"newly created user %q was not found",
			username,
		)
	}

	if err := user.SetPassword(password); err != nil {
		return nil, fmt.Errorf(
			"set password for newly created user %q: %w",
			username,
			err,
		)
	}

	return user, nil
}

func (u *User) GrantSudo() error {
	err := exec.Command(
		"usermod",
		"--append",
		"--groups", "sudo",
		"--",
		u.Username,
	).Run()
	if err != nil {
		return fmt.Errorf(
			"grant sudo access to %q: %w",
			u.Username,
			err,
		)
	}

	return nil
}

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
