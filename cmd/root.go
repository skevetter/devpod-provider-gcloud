package cmd

import (
	"os"
	"os/exec"

	"github.com/skevetter/log"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

// NewRootCmd returns a new root command
func NewRootCmd() *cobra.Command {
	gcloudCmd := &cobra.Command{
		Use:           "devpod-provider-gcloud",
		Short:         "gcloud Provider commands",
		SilenceErrors: true,
		SilenceUsage:  true,

		PersistentPreRunE: func(cobraCmd *cobra.Command, args []string) error {
			log.Default.MakeRaw()
			return nil
		},
	}

	return gcloudCmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// build the root command
	rootCmd := BuildRoot()

	// execute command
	err := rootCmd.Execute()
	if err != nil {
		if exitErr, ok := err.(*ssh.ExitError); ok {
			os.Exit(exitErr.ExitStatus())
		}
		if exitErr, ok := err.(*exec.ExitError); ok {
			if len(exitErr.Stderr) > 0 {
				log.Default.ErrorStreamOnly().Error(string(exitErr.Stderr))
			}
			os.Exit(exitErr.ExitCode())
		}

		log.Default.Fatal(err)
	}
}

// BuildRoot creates a new root command from the
func BuildRoot() *cobra.Command {
	rootCmd := NewRootCmd()

	rootCmd.AddCommand(NewCreateCmd())
	rootCmd.AddCommand(NewStatusCmd())
	rootCmd.AddCommand(NewDeleteCmd())
	rootCmd.AddCommand(NewStartCmd())
	rootCmd.AddCommand(NewStopCmd())
	rootCmd.AddCommand(NewCommandCmd())
	rootCmd.AddCommand(NewTokenCmd())
	rootCmd.AddCommand(NewInitCmd())
	rootCmd.AddCommand(NewDescribeCmd())
	return rootCmd
}
