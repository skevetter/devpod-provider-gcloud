package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path"
	"regexp"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/skevetter/devpod-provider-gcloud/pkg/gcloud"
	"github.com/skevetter/devpod-provider-gcloud/pkg/options"
	"github.com/skevetter/devpod/pkg/ssh"
	"github.com/spf13/cobra"
)

var iapPortPattern = regexp.MustCompile(`Listening on port \[(\d+)\]`)

// CommandCmd holds the cmd flags.
type CommandCmd struct{}

// NewCommandCmd defines a command.
func NewCommandCmd() *cobra.Command {
	cmd := &CommandCmd{}
	return &cobra.Command{
		Use:   "command",
		Short: "Run a command on the instance",
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			optionsFromEnv, err := options.FromEnv(true, true)
			if err != nil {
				return err
			}

			return cmd.Run(cobraCmd.Context(), optionsFromEnv)
		},
	}
}

type sshTarget struct {
	host    string
	port    string
	cleanup func()
}

// Run runs the command logic.
func (cmd *CommandCmd) Run(ctx context.Context, options *options.Options) error {
	command := os.Getenv("COMMAND")
	if command == "" {
		return fmt.Errorf("command environment variable is missing")
	}

	privateKey, err := ssh.GetPrivateKeyRawBase(options.MachineFolder)
	if err != nil {
		return fmt.Errorf("load private key: %w", err)
	}

	instance, err := getInstance(ctx, options)
	if err != nil {
		return err
	}

	t, err := resolveTarget(ctx, options, instance)
	if err != nil {
		return err
	}
	if t.cleanup != nil {
		defer t.cleanup()
	}

	sshClient, err := ssh.NewSSHClient("devpod", net.JoinHostPort(t.host, t.port), privateKey)
	if err != nil {
		return fmt.Errorf("create ssh client: %w", err)
	}
	defer func() { _ = sshClient.Close() }()

	return ssh.Run(ctx, ssh.RunOptions{
		Client:  sshClient,
		Command: command,
		Stdin:   os.Stdin,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
		EnvVars: nil,
	})
}

func getInstance(
	ctx context.Context, options *options.Options,
) (*computepb.Instance, error) {
	client, err := gcloud.NewClient(ctx, options.Project, options.Zone)
	if err != nil {
		return nil, err
	}
	defer func() { _ = client.Close() }()

	instance, err := client.Get(ctx, options.MachineID)
	if err != nil {
		return nil, err
	}
	if instance == nil {
		return nil, fmt.Errorf("instance %s doesn't exist", options.MachineID)
	}

	return instance, nil
}

func resolveTarget(
	ctx context.Context,
	options *options.Options,
	instance *computepb.Instance,
) (sshTarget, error) {
	if options.PublicIP {
		return resolvePublicTarget(instance)
	}

	return resolveIAPTarget(ctx, options.Project, instance)
}

func resolvePublicTarget(
	instance *computepb.Instance,
) (sshTarget, error) {
	noExternalIP := len(instance.NetworkInterfaces) == 0 ||
		len(instance.NetworkInterfaces[0].AccessConfigs) == 0 ||
		instance.NetworkInterfaces[0].AccessConfigs[0].NatIP == nil
	if noExternalIP {
		return sshTarget{}, fmt.Errorf(
			"instance %s doesn't have an external nat ip",
			instance.GetName(),
		)
	}

	return sshTarget{
		host: *instance.NetworkInterfaces[0].AccessConfigs[0].NatIP,
		port: "22",
	}, nil
}

func resolveIAPTarget(
	ctx context.Context, project string, instance *computepb.Instance,
) (sshTarget, error) {
	if instance.GetName() == "" || instance.GetZone() == "" {
		return sshTarget{}, fmt.Errorf("instance missing name or zone")
	}

	zoneName := path.Base(instance.GetZone())

	gcloudCmd := exec.CommandContext( //nolint:gosec // args from trusted provider config
		ctx, "gcloud",
		"compute", "start-iap-tunnel",
		instance.GetName(), "22",
		"--local-host-port=localhost:0",
		"--zone="+zoneName,
		"--project="+project,
	)

	stderrPipe, err := gcloudCmd.StderrPipe()
	if err != nil {
		return sshTarget{}, fmt.Errorf("create stderr pipe: %w", err)
	}

	if err = gcloudCmd.Start(); err != nil {
		return sshTarget{}, fmt.Errorf("start tunnel: %w", err)
	}

	waitErr := make(chan error, 1)
	go func() { waitErr <- gcloudCmd.Wait() }()

	port, err := parseIAPPort(ctx, stderrPipe)
	if err != nil {
		_ = gcloudCmd.Process.Kill()
		<-waitErr
		return sshTarget{}, fmt.Errorf("parse IAP tunnel port: %w", err)
	}

	return sshTarget{
		host: "localhost",
		port: port,
		cleanup: func() {
			_ = gcloudCmd.Process.Kill()
			<-waitErr
		},
	}, nil
}

func parseIAPPort(ctx context.Context, r io.Reader) (string, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	portCh := make(chan string, 1)
	errCh := make(chan error, 1)
	go func() {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			if m := iapPortPattern.FindStringSubmatch(scanner.Text()); m != nil {
				portCh <- m[1]
				return
			}
		}
		if err := scanner.Err(); err != nil {
			errCh <- err
		} else {
			errCh <- fmt.Errorf("gcloud exited without reporting a listening port")
		}
	}()

	select {
	case port := <-portCh:
		return port, nil
	case err := <-errCh:
		return "", err
	case <-timeoutCtx.Done():
		return "", timeoutCtx.Err()
	}
}
