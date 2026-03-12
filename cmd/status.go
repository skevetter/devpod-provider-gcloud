package cmd

import (
	"context"
	"os"

	"github.com/skevetter/devpod-provider-gcloud/pkg/gcloud"
	"github.com/skevetter/devpod-provider-gcloud/pkg/options"
	"github.com/spf13/cobra"
)

// StatusCmd holds the cmd flags.
type StatusCmd struct{}

// NewStatusCmd defines a command.
func NewStatusCmd() *cobra.Command {
	cmd := &StatusCmd{}
	return &cobra.Command{
		Use:   "status",
		Short: "Retrieve the status of an instance",
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			optionsFromEnv, err := options.FromEnv(true, true)
			if err != nil {
				return err
			}

			return cmd.Run(cobraCmd.Context(), optionsFromEnv)
		},
	}
}

// Run runs the command logic.
func (cmd *StatusCmd) Run(ctx context.Context, opts *options.Options) error {
	return withGCloudClient(ctx, opts, func(ctx context.Context, c *gcloud.Client) error {
		status, err := c.Status(ctx, opts.MachineID)
		if err != nil {
			return err
		}

		_, err = os.Stdout.WriteString(string(status))
		return err
	})
}
