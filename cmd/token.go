package cmd

import (
	"context"
	"os"

	"github.com/skevetter/devpod-provider-gcloud/pkg/gcloud"
	"github.com/spf13/cobra"
)

// TokenCmd holds the cmd flags.
type TokenCmd struct{}

// NewTokenCmd defines a command.
func NewTokenCmd() *cobra.Command {
	cmd := &TokenCmd{}
	return &cobra.Command{
		Use:   "token",
		Short: "Prints an access token",
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			return cmd.Run(cobraCmd.Context())
		},
	}
}

// Run runs the command logic.
func (cmd *TokenCmd) Run(ctx context.Context) error {
	tok, err := gcloud.GetToken(ctx)
	if err != nil {
		return err
	}

	_, err = os.Stdout.Write(tok)
	return err
}
