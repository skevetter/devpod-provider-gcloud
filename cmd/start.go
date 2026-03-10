package cmd

import (
	"context"

	"github.com/skevetter/devpod-provider-gcloud/pkg/gcloud"
	"github.com/skevetter/devpod-provider-gcloud/pkg/options"
	"github.com/spf13/cobra"
)

// StartCmd holds the cmd flags
type StartCmd struct{}

// NewStartCmd defines a command
func NewStartCmd() *cobra.Command {
	cmd := &StartCmd{}
	return &cobra.Command{
		Use:   "start",
		Short: "Start an instance",
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			optionsFromEnv, err := options.FromEnv(true, true)
			if err != nil {
				return err
			}

			return cmd.Run(cobraCmd.Context(), optionsFromEnv)
		},
	}
}

// Run runs the command logic
func (cmd *StartCmd) Run(ctx context.Context, options *options.Options) error {
	client, err := gcloud.NewClient(ctx, options.Project, options.Zone)
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }()

	return client.Start(ctx, options.MachineID)
}
