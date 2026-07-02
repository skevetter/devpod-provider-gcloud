package gcloud

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/googleapis/gax-go/v2/apierror"
	"github.com/skevetter/devpod-provider-gcloud/pkg/options"
	"github.com/skevetter/devpod/pkg/client"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type Client struct {
	InstanceClient *compute.InstancesClient

	Project string
	Zone    string
}

func NewClient(
	ctx context.Context,
	project, zone string,
	opts ...option.ClientOption,
) (*Client, error) {
	err := SetupEnvJson()
	if err != nil {
		return nil, err
	}

	instanceClient, err := compute.NewInstancesRESTClient(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return &Client{
		InstanceClient: instanceClient,
		Project:        project,
		Zone:           zone,
	}, nil
}

func SetupEnvJson() error {
	gcloudKeyFile := options.GetEnv("KEY_FILE")

	if gcloudKeyFile == "" {
		gcloudKey := options.GetEnv("KEY")
		if gcloudKey == "" {
			gcloudKey = os.Getenv("GCLOUD_JSON_AUTH")
		}

		var err error
		if gcloudKey != "" {
			gcloudKeyFile, err = writeKey(gcloudKey)
			if err != nil {
				return err
			}
		}
	}

	if gcloudKeyFile != "" {
		return os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", gcloudKeyFile)
	}

	return nil
}

func writeKey(gcloudKey string) (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	gcloudKeyFile := filepath.Join(filepath.Dir(exePath), "gcloud_auth.json")

	err = os.WriteFile(
		gcloudKeyFile,
		[]byte(gcloudKey),
		0o600,
	) // #nosec G703
	if err != nil {
		return "", err
	}

	return gcloudKeyFile, err
}

func DefaultTokenSource(ctx context.Context) (oauth2.TokenSource, error) {
	scopes := []string{
		"https://www.googleapis.com/auth/cloud-platform",
	}

	return google.DefaultTokenSource(ctx, scopes...)
}

func ParseToken(tok string) (*oauth2.Token, error) {
	oauthToken := &oauth2.Token{}
	err := json.Unmarshal([]byte(tok), oauthToken)
	if err != nil {
		return nil, err
	}

	return oauthToken, nil
}

func GetToken(ctx context.Context) ([]byte, error) {
	err := SetupEnvJson()
	if err != nil {
		return nil, err
	}

	tokSource, err := DefaultTokenSource(ctx)
	if err != nil {
		return nil, err
	}

	t, err := tokSource.Token()
	if err != nil {
		return nil, err
	}

	t.RefreshToken = ""
	t.TokenType = ""
	return json.Marshal(t) //nolint:gosec // AccessToken is intentionally marshaled for provider use
}

func (c *Client) Init(ctx context.Context) error {
	_, err := c.InstanceClient.List(ctx, &computepb.ListInstancesRequest{
		Project: c.Project,
		Zone:    c.Zone,
	}).Next()
	if err != nil && !errors.Is(err, iterator.Done) {
		return fmt.Errorf("cannot list instances: %v", err)
	}

	return nil
}

func (c *Client) Create(ctx context.Context, instance *computepb.Instance) error {
	operation, err := c.InstanceClient.Insert(ctx, &computepb.InsertInstanceRequest{
		InstanceResource: instance,
		Project:          c.Project,
		Zone:             c.Zone,
	})
	if err != nil {
		return err
	}

	return operation.Wait(ctx)
}

func (c *Client) Start(ctx context.Context, name string) error {
	operation, err := c.InstanceClient.Start(ctx, &computepb.StartInstanceRequest{
		Instance: name,
		Project:  c.Project,
		Zone:     c.Zone,
	})
	if err != nil {
		return err
	}

	return operation.Wait(ctx)
}

func (c *Client) Stop(ctx context.Context, name string, async bool) error {
	operation, err := c.InstanceClient.Stop(ctx, &computepb.StopInstanceRequest{
		Instance: name,
		Project:  c.Project,
		Zone:     c.Zone,
	})
	if err != nil {
		return err
	} else if async {
		return nil
	}

	return operation.Wait(ctx)
}

func (c *Client) Delete(ctx context.Context, name string) error {
	operation, err := c.InstanceClient.Delete(ctx, &computepb.DeleteInstanceRequest{
		Instance: name,
		Project:  c.Project,
		Zone:     c.Zone,
	})
	if err != nil {
		return err
	}

	return operation.Wait(ctx)
}

func (c *Client) Get(ctx context.Context, name string) (*computepb.Instance, error) {
	instance, err := c.InstanceClient.Get(ctx, &computepb.GetInstanceRequest{
		Instance: name,
		Project:  c.Project,
		Zone:     c.Zone,
	})
	if err != nil {
		var apiErr *apierror.APIError
		if errors.As(err, &apiErr) {
			var googleErr *googleapi.Error
			if errors.As(apiErr.Unwrap(), &googleErr) && googleErr.Code == 404 {
				return nil, nil
			}
		}

		return nil, err
	}

	return instance, nil
}

func (c *Client) Status(ctx context.Context, name string) (client.Status, error) {
	instance, err := c.Get(ctx, name)
	if err != nil || instance == nil {
		return client.StatusNotFound, err
	}

	status := strings.TrimSpace(strings.ToUpper(*instance.Status))
	switch status {
	case "RUNNING":
		return client.StatusRunning, nil
	case "STOPPING", "SUSPENDING", "REPAIRING", "PROVISIONING", "STAGING":
		return client.StatusBusy, nil
	case "TERMINATED":
		return client.StatusStopped, nil
	}

	return client.StatusNotFound, fmt.Errorf("unexpected status: %v", status)
}

func (c *Client) Describe(ctx context.Context, name string) (string, error) {
	instance, err := c.Get(ctx, name)
	if err != nil || instance == nil {
		return client.DescriptionNotFound, err
	}

	bytes, err := json.MarshalIndent(instance, "", "  ")
	if err != nil {
		return client.DescriptionNotFound, nil
	}
	return string(bytes), nil
}

func (c *Client) Close() error {
	return c.InstanceClient.Close()
}
