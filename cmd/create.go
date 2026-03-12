package cmd

import (
	"context"
	"encoding/base64"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/skevetter/devpod-provider-gcloud/pkg/gcloud"
	"github.com/skevetter/devpod-provider-gcloud/pkg/options"
	"github.com/skevetter/devpod-provider-gcloud/pkg/ptr"
	"github.com/skevetter/devpod/pkg/ssh"
	"github.com/spf13/cobra"
)

// CreateCmd holds the cmd flags.
type CreateCmd struct{}

// NewCreateCmd defines a command.
func NewCreateCmd() *cobra.Command {
	cmd := &CreateCmd{}
	return &cobra.Command{
		Use:   "create",
		Short: "Create an instance",
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
func (cmd *CreateCmd) Run(ctx context.Context, options *options.Options) error {
	client, err := gcloud.NewClient(ctx, options.Project, options.Zone)
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }()

	instance, err := buildInstance(options)
	if err != nil {
		return err
	}

	return client.Create(ctx, instance)
}

func buildInstance(options *options.Options) (*computepb.Instance, error) {
	diskSize, err := strconv.Atoi(options.DiskSize)
	if err != nil {
		return nil, fmt.Errorf("parse disk size: %w", err)
	}

	publicKey, err := loadPublicKey(options.MachineFolder)
	if err != nil {
		return nil, err
	}

	instance := &computepb.Instance{
		Scheduling:        buildScheduling(options.MachineType),
		Metadata:          buildMetadata(publicKey),
		MachineType:       ptr.Ptr(machineTypeURI(options)),
		Disks:             buildDisks(options, int64(diskSize)),
		Tags:              buildInstanceTags(options),
		NetworkInterfaces: buildNetworkInterfaces(options),
		Name:              ptr.Ptr(options.MachineID),
		ServiceAccounts:   buildServiceAccounts(options),
	}

	return instance, nil
}

func loadPublicKey(machineFolder string) (string, error) {
	publicKeyBase, err := ssh.GetPublicKeyBase(machineFolder)
	if err != nil {
		return "", fmt.Errorf("generate public key: %w", err)
	}

	decoded, err := base64.StdEncoding.DecodeString(publicKeyBase)
	if err != nil {
		return "", err
	}

	return string(decoded), nil
}

func buildScheduling(machineType string) *computepb.Scheduling {
	return &computepb.Scheduling{
		AutomaticRestart:  ptr.Ptr(true),
		OnHostMaintenance: ptr.Ptr(getMaintenancePolicy(machineType)),
	}
}

func buildMetadata(publicKey string) *computepb.Metadata {
	return &computepb.Metadata{
		Items: []*computepb.Items{
			{
				Key:   ptr.Ptr("ssh-keys"),
				Value: ptr.Ptr("devpod:" + publicKey),
			},
		},
	}
}

func machineTypeURI(options *options.Options) string {
	return fmt.Sprintf(
		"projects/%s/zones/%s/machineTypes/%s",
		options.Project, options.Zone, options.MachineType,
	)
}

func buildDisks(options *options.Options, diskSize int64) []*computepb.AttachedDisk {
	return []*computepb.AttachedDisk{
		{
			AutoDelete: ptr.Ptr(true),
			Boot:       ptr.Ptr(true),
			DeviceName: ptr.Ptr(options.MachineID),
			InitializeParams: &computepb.AttachedDiskInitializeParams{
				DiskSizeGb: ptr.Ptr(diskSize),
				DiskType: ptr.Ptr(fmt.Sprintf(
					"projects/%s/zones/%s/diskTypes/pd-balanced",
					options.Project, options.Zone,
				)),
				SourceImage: ptr.Ptr(options.DiskImage),
			},
		},
	}
}

func buildNetworkInterfaces(options *options.Options) []*computepb.NetworkInterface {
	return []*computepb.NetworkInterface{
		{
			Network:       normalizeNetworkID(options),
			Subnetwork:    normalizeSubnetworkID(options),
			AccessConfigs: getAccessConfig(options),
		},
	}
}

func buildServiceAccounts(options *options.Options) []*computepb.ServiceAccount {
	if options.ServiceAccount == "" {
		return []*computepb.ServiceAccount{}
	}

	return []*computepb.ServiceAccount{
		{
			Email: &options.ServiceAccount,
			Scopes: []string{
				"https://www.googleapis.com/auth/cloud-platform",
			},
		},
	}
}

func getAccessConfig(options *options.Options) []*computepb.AccessConfig {
	if options.PublicIP {
		return []*computepb.AccessConfig{
			{
				Name:        ptr.Ptr("External NAT"),
				NetworkTier: ptr.Ptr("STANDARD"),
			},
		}
	}

	return nil
}

func buildInstanceTags(options *options.Options) *computepb.Tags {
	if len(options.Tag) == 0 {
		return nil
	}

	return &computepb.Tags{Items: []string{options.Tag}}
}

func normalizeNetworkID(options *options.Options) *string {
	network := options.Network
	if network == "" {
		return nil
	}

	// projects/{{project}}/global/networks/{{name}}
	if strings.HasPrefix(network, "projects/") {
		return ptr.Ptr(network)
	}

	// {{project}}/{{name}}
	if project, name, ok := strings.Cut(network, "/"); ok {
		return ptr.Ptr(fmt.Sprintf("projects/%s/global/networks/%s", project, name))
	}

	// {{name}}
	return ptr.Ptr(fmt.Sprintf("projects/%s/global/networks/%s", options.Project, network))
}

func normalizeSubnetworkID(options *options.Options) *string {
	sn := strings.TrimSpace(options.Subnetwork)
	if sn == "" {
		return nil
	}

	project := options.Project
	zone := options.Zone
	region := zone[:strings.LastIndex(zone, "-")]

	parts := strings.Split(sn, "/")
	switch len(parts) {
	case 1:
		// {{name}}
		return ptr.Ptr(fmt.Sprintf("projects/%s/regions/%s/subnetworks/%s", project, region, sn))
	case 2:
		// {{region}}/{{name}}
		return ptr.Ptr(
			fmt.Sprintf("projects/%s/regions/%s/subnetworks/%s", project, parts[0], parts[1]),
		)
	case 3:
		// {{project}}/{{region}}/{{name}}
		return ptr.Ptr(
			fmt.Sprintf("projects/%s/regions/%s/subnetworks/%s", parts[0], parts[1], parts[2]),
		)
	default:
		// projects/{{project}}/regions/{{region}}/subnetworks/{{name}} or other full path
		return ptr.Ptr(sn)
	}
}

var gpuInstancePattern = regexp.MustCompile(`^[agn][0-9]`)

func getMaintenancePolicy(machineType string) string {
	if gpuInstancePattern.MatchString(machineType) {
		return "TERMINATE"
	}

	return "MIGRATE"
}
