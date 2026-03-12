package cmd

import (
	"context"
	"os"

	"github.com/skevetter/devpod-provider-gcloud/pkg/gcloud"
	"github.com/skevetter/devpod-provider-gcloud/pkg/options"
	"github.com/spf13/cobra"
)

// DescribeCmd holds the cmd flags.
type DescribeCmd struct{}

// NewDescribeCmd defines a command.
func NewDescribeCmd() *cobra.Command {
	cmd := &DescribeCmd{}
	return &cobra.Command{
		Use:   "describe",
		Short: "Retrieve description of the virtual machine",
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
func (cmd *DescribeCmd) Run(ctx context.Context, opts *options.Options) error {
	return withGCloudClient(ctx, opts, func(ctx context.Context, c *gcloud.Client) error {
		result, err := c.Describe(ctx, opts.MachineID)
		if err != nil {
			return err
		}

		_, err = os.Stdout.WriteString(result)
		return err
	})
}
