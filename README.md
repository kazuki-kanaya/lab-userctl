# lab-userctl

![Platform: Ubuntu](https://img.shields.io/badge/platform-Ubuntu-E95420?logo=ubuntu&logoColor=white)

[日本語版](README.ja.md)

A simple interactive CLI for setting up Linux users on Ubuntu servers, including sudo access and SSH public keys.

## Install

Install the latest release for Ubuntu:

```bash
curl -fsSL https://raw.githubusercontent.com/kazuki-kanaya/lab-userctl/main/scripts/install.sh | sh
```

The installer verifies the downloaded archive against its published SHA-256 checksum and installs `lab-userctl` to `/usr/local/bin`.

## Why

This tool started as a way to remove repetitive account setup work on a lab GPU server: creating users, granting sudo access, registering SSH public keys, and setting the required ownership and permissions.

## What it does

With a single command, it can perform the following account setup tasks:

- Creates a local user when needed
- Sets a password for new users
- Optionally grants sudo access
- Optionally registers an SSH public key
- Applies secure permissions to `.ssh` and `authorized_keys`
- Avoids duplicate SSH public keys

## Usage

```bash
sudo lab-userctl register
```

The command interactively asks for a username, a password when creating a user, whether to grant sudo access, and whether to register an SSH public key.

Only SSH public keys are accepted. Never enter a private key; private key input is rejected.

This tool changes system accounts. Test it on a disposable Ubuntu VM or test account before using it on a production server.

## Tech stack

- Go
- Cobra
- GoReleaser
- GitHub Actions
