package cmd

import (
	"context"

	"github.com/skevetter/devpod-provider-gcloud/pkg/gcloud"
	"github.com/skevetter/devpod-provider-gcloud/pkg/options"
	"github.com/spf13/cobra"
)

// DeleteCmd holds the cmd flags.
type DeleteCmd struct{}

// NewDeleteCmd defines a command.
func NewDeleteCmd() *cobra.Command {
	cmd := &DeleteCmd{}
	return &cobra.Command{
		Use:   "delete",
		Short: "Delete an instance",
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
func (cmd *DeleteCmd) Run(ctx context.Context, options *options.Options) error {
	client, err := gcloud.NewClient(ctx, options.Project, options.Zone)
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }()

	return client.Delete(ctx, options.MachineID)
}
