package cmd

import (
	"context"

	"github.com/skevetter/devpod-provider-gcloud/pkg/gcloud"
	"github.com/skevetter/devpod-provider-gcloud/pkg/options"
)

func withGCloudClient(
	ctx context.Context,
	opts *options.Options,
	fn func(ctx context.Context, c *gcloud.Client) error,
) error {
	client, err := gcloud.NewClient(ctx, opts.Project, opts.Zone)
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }()

	return fn(ctx, client)
}
