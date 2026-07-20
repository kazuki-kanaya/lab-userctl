package sshkey

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

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

	keysFD, err := openOrCreateAuthorizedKeys(
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

func authorizedKeyLine(key PublicKey) string {
	line := strings.TrimSuffix(
		string(ssh.MarshalAuthorizedKey(key.key)),
		"\n",
	)

	if key.comment == "" {
		return line + "\n"
	}

	return line + " " + key.comment + "\n"
}

func openOrCreateSSHDir(
	homeFD int,
	uid int,
	gid int,
) (int, error) {
	sshFD, err := unix.Openat(
		homeFD,
		".ssh",
		unix.O_RDONLY|
			unix.O_DIRECTORY|
			unix.O_CLOEXEC|
			unix.O_NOFOLLOW,
		0,
	)
	if err != nil {
		if !errors.Is(err, unix.ENOENT) {
			return -1, fmt.Errorf("open .ssh directory: %w", err)
		}

		if err := unix.Mkdirat(homeFD, ".ssh", 0o700); err != nil && !errors.Is(err, unix.EEXIST) {
			return -1, fmt.Errorf("create .ssh directory: %w", err)
		}

		sshFD, err = unix.Openat(
			homeFD,
			".ssh",
			unix.O_RDONLY|
				unix.O_DIRECTORY|
				unix.O_CLOEXEC|
				unix.O_NOFOLLOW,
			0,
		)
		if err != nil {
			return -1, fmt.Errorf("open .ssh owner: %w", err)
		}
	}
	if err := unix.Fchown(sshFD, uid, gid); err != nil {
		unix.Close(sshFD)
		return -1, fmt.Errorf("set .ssh owner: %w", err)
	}

	if err := unix.Fchmod(sshFD, 0o700); err != nil {
		unix.Close(sshFD)
		return -1, fmt.Errorf("set .ssh permission")
	}

	return sshFD, nil
}

func openOrCreateAuthorizedKeys(
	sshFD int,
	uid int,
	gid int,
) (int, error) {
	keysFD, err := unix.Openat(
		sshFD,
		"authorized_keys",
		unix.O_RDWR|
			unix.O_APPEND|
			unix.O_CREAT|
			unix.O_NOFOLLOW,
		0o600,
	)
	if err != nil {
		return -1, fmt.Errorf("open authorized_keys: %w", err)
	}
	if err := unix.Fchown(keysFD, uid, gid); err != nil {
		unix.Close(keysFD)
		return -1, fmt.Errorf(
			"set authorized_keys owner: %w",
			err,
		)
	}

	if err := unix.Fchmod(keysFD, 0o600); err != nil {
		unix.Close(keysFD)
		return -1, fmt.Errorf(
			"set authorized_keys permissions: %w",
			err,
		)
	}

	return keysFD, nil
}
