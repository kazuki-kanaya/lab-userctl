package sshkey

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/kazuki-kanaya/lab-userctl/internal/account"
	"golang.org/x/crypto/ssh"
	"golang.org/x/sys/unix"
)

func Add(user *account.User, key PublicKey) (added bool, err error) {
	if user == nil {
		return false, fmt.Errorf("user must not be nil")
	}

	homeFD, err := unix.Open(
		user.HomeDir,
		unix.O_RDONLY|
			unix.O_DIRECTORY|
			unix.O_CLOEXEC|
			unix.O_NOFOLLOW,
		0,
	)
	if err != nil {
		return false, fmt.Errorf(
			"open home directory for %q: %w",
			user.Username,
			err,
		)
	}
	defer unix.Close(homeFD)

	sshFD, err := openOrCreateSSHDir(
		homeFD,
		user.UID,
		user.GID,
	)
	if err != nil {
		return false, err
	}
	defer unix.Close(sshFD)

	keysFD, err := openOrCrateAuthorizedKeys(
		sshFD,
		user.UID,
		user.GID,
	)
	if err != nil {
		return false, err
	}

	keysFile := os.NewFile(
		uintptr(keysFD),
		"authorized_keys",
	)
	defer keysFile.Close()

	content, err := io.ReadAll(keysFile)
	if err != nil {
		return false, fmt.Errorf(
			"read authorized_keys for %q: %w",
			user.Username,
			err,
		)
	}

	exists, err := containsPublicKey(content, key)
	if err != nil {
		return false, err
	}

	if exists {
		return false, nil
	}

	line := authorizedKeyLine(key)
	if len(content) > 0 && content[len(content)-1] != '\n' {
		line = "\n" + line
	}

	if _, err := io.WriteString(keysFile, line); err != nil {
		return false, fmt.Errorf(
			"append SSH key for %q: %w",
			user.Username,
			err,
		)
	}

	return true, nil
}

func containsPublicKey(
	content []byte,
	target PublicKey,
) (bool, error) {
	scanner := bufio.NewScanner(bytes.NewReader(content))
	scanner.Buffer(make([]byte, 1024), 1024*1024)

	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		parsed, _, _, rest, err := ssh.ParseAuthorizedKey(line)
		if err != nil || len(bytes.TrimSpace(rest)) > 0 {
			continue
		}

		existing := PublicKey{
			key: parsed,
		}
		if target.Equal(existing) {
			return true, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf(
			"scan authorized_keys: %w",
			err,
		)
	}

	return false, nil
}
