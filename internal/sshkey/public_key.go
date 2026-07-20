package sshkey

import (
	"bytes"
	"fmt"
	"strings"

	"golang.org/x/crypto/ssh"
)

type PublicKey struct {
	key     ssh.PublicKey
	comment string
}

func Parse(input string) (PublicKey, error) {
	input = strings.TrimSpace(input)

	if input == "" {
		return PublicKey{}, fmt.Errorf("public key must not be empty")
	}

	if strings.Contains(input, "PRIVATE KEY") {
		return PublicKey{}, fmt.Errorf(
			"private keys must not be submitted",
		)
	}

	key, comment, options, rest, err := ssh.ParseAuthorizedKey([]byte(input))
	if err != nil {
		return PublicKey{}, fmt.Errorf(
			"parse SSH public key: %w",
			err,
		)
	}

	if len(options) > 0 {
		return PublicKey{}, fmt.Errorf(
			"authorized key options are not supported",
		)
	}

	if len(bytes.TrimSpace(rest)) > 0 {
		return PublicKey{}, fmt.Errorf(
			"provide exactly one public key",
		)
	}

	return PublicKey{
		key:     key,
		comment: comment,
	}, nil
}

func (k PublicKey) Equal(other PublicKey) bool {
	return k.key.Type() == other.key.Type() &&
		bytes.Equal(k.key.Marshal(), other.key.Marshal())
}
