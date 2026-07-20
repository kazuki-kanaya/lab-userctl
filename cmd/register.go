package cmd

import (
	"fmt"
	"os"

	"github.com/kazuki-kanaya/lab-userctl/internal/register"
	"github.com/kazuki-kanaya/lab-userctl/internal/terminal"
	"github.com/spf13/cobra"
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a user with sudo access and an SSH public key",
	RunE: func(cmd *cobra.Command, args []string) error {
		if os.Geteuid() != 0 {
			return fmt.Errorf("run this command with sudo")
		}

		service := register.New(terminal.NewPrompter())
		return service.Run()
	},
}

func init() {
	rootCmd.AddCommand(registerCmd)
}
