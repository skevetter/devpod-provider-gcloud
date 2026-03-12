package cmd

import (
	"context"

	"github.com/skevetter/devpod-provider-gcloud/pkg/gcloud"
	"github.com/skevetter/devpod-provider-gcloud/pkg/options"
	"github.com/spf13/cobra"
)

// InitCmd holds the cmd flags.
type InitCmd struct{}

// NewInitCmd defines a command.
func NewInitCmd() *cobra.Command {
	cmd := &InitCmd{}
	return &cobra.Command{
		Use:   "init",
		Short: "Init an instance",
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			optionsFromEnv, err := options.FromEnv(false, false)
			if err != nil {
				return err
			}

			return cmd.Run(cobraCmd.Context(), optionsFromEnv)
		},
	}
}

// Run runs the command logic.
func (cmd *InitCmd) Run(ctx context.Context, options *options.Options) error {
	client, err := gcloud.NewClient(ctx, options.Project, options.Zone)
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }()

	return client.Init(ctx)
}
