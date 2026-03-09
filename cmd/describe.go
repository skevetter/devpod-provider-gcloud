package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/skevetter/devpod-provider-gcloud/pkg/gcloud"
	"github.com/skevetter/devpod-provider-gcloud/pkg/options"
	"github.com/skevetter/log"
	"github.com/spf13/cobra"
)

// DescribeCmd holds the cmd flags
type DescribeCmd struct{}

// NewDescribeCmd defines a command
func NewDescribeCmd() *cobra.Command {
	cmd := &DescribeCmd{}
	describeCmd := &cobra.Command{
		Use:   "describe",
		Short: "Retrieve description of the virtual machine",
		RunE: func(_ *cobra.Command, args []string) error {
			optionsFromEnv, err := options.FromEnv(true, true)
			if err != nil {
				return err
			}

			return cmd.Run(context.Background(), optionsFromEnv, log.Default)
		},
	}

	return describeCmd
}

// Run runs the command logic
func (cmd *DescribeCmd) Run(ctx context.Context, options *options.Options, log log.Logger) error {
	client, err := gcloud.NewClient(ctx, options.Project, options.Zone)
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }()

	json, err := client.Describe(ctx, options.MachineID)
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(os.Stdout, json)
	return err
}
