package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/skevetter/devpod-provider-gcloud/pkg/gcloud"
	"github.com/skevetter/devpod-provider-gcloud/pkg/options"
	"github.com/spf13/cobra"
)

// DescribeCmd holds the cmd flags
type DescribeCmd struct{}

// NewDescribeCmd defines a command
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

// Run runs the command logic
func (cmd *DescribeCmd) Run(ctx context.Context, options *options.Options) error {
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
