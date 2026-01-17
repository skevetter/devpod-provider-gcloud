package cmd

import (
	"context"

	"github.com/skevetter/devpod-provider-gcloud/pkg/gcloud"
	"github.com/skevetter/devpod-provider-gcloud/pkg/options"
	"github.com/skevetter/log"
	"github.com/spf13/cobra"
)

// StartCmd holds the cmd flags
type StartCmd struct{}

// NewStartCmd defines a command
func NewStartCmd() *cobra.Command {
	cmd := &StartCmd{}
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start an instance",
		RunE: func(_ *cobra.Command, args []string) error {
			options, err := options.FromEnv(true, true)
			if err != nil {
				return err
			}

			return cmd.Run(context.Background(), options, log.Default)
		},
	}

	return startCmd
}

// Run runs the command logic
func (cmd *StartCmd) Run(ctx context.Context, options *options.Options, log log.Logger) error {
	client, err := gcloud.NewClient(ctx, options.Project, options.Zone)
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }()

	return client.Start(ctx, options.MachineID)
}
