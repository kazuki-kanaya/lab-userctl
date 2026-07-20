package terminal

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/x/term"
)

type Prompter struct {
	reader *bufio.Reader
}

func NewPrompter() *Prompter {
	return &Prompter{
		reader: bufio.NewReader(os.Stdin),
	}
}

func (p *Prompter) Ask(label string) (string, error) {
	fmt.Fprint(os.Stdout, label)
	return p.readLine()
}

func (p *Prompter) AskPassword(label string) ([]byte, error) {
	fmt.Fprint(os.Stdout, label)
	password, err := term.ReadPassword(os.Stdin.Fd())
	fmt.Fprintln(os.Stdout)
	return password, err
}

func (p *Prompter) readLine() (string, error) {
	line, err := p.reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

func (p *Prompter) Confirm(question string, defaultYes bool) (bool, error) {
	choice := "[y/N]"
	if defaultYes {
		choice = "[Y/n]"
	}

	for {
		answer, err := p.Ask(fmt.Sprintf("%s %s:", question, choice))

		if err != nil {
			return false, err
		}

		switch strings.ToLower(strings.TrimSpace(answer)) {
		case "":
			return defaultYes, nil
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		default:
			fmt.Fprintln(os.Stdout, "Please enter yes or no.")
		}
	}
}
