package options

import (
	"fmt"
	"os"
	"strings"
)

// Options holds all provider configuration.
type Options struct {
	MachineID     string
	MachineFolder string

	Project        string
	Zone           string
	Network        string
	Subnetwork     string
	Tag            string
	DiskSize       string
	DiskImage      string
	MachineType    string
	ServiceAccount string
	PublicIP       bool
}

// FromEnv loads options from environment variables.
func FromEnv(withMachine, withFolder bool) (*Options, error) {
	retOptions := &Options{}

	if err := loadMachineOptions(retOptions, withMachine, withFolder); err != nil {
		return nil, err
	}

	if err := loadRequiredOptions(retOptions); err != nil {
		return nil, err
	}

	retOptions.PublicIP = GetEnv("PUBLIC_IP_ENABLED") == "true"
	retOptions.ServiceAccount = GetEnv("SERVICE_ACCOUNT")
	retOptions.Network = GetEnv("NETWORK")
	retOptions.Subnetwork = GetEnv("SUBNETWORK")
	retOptions.Tag = GetEnv("TAG")

	return retOptions, nil
}

func loadMachineOptions(opts *Options, withMachine, withFolder bool) error {
	var err error
	if withMachine {
		opts.MachineID, err = fromEnvOrError("MACHINE_ID")
		if err != nil {
			return err
		}
		if !strings.HasPrefix(opts.MachineID, "devpod-") {
			opts.MachineID = "devpod-" + opts.MachineID
		}
	}
	if withFolder {
		opts.MachineFolder, err = fromEnvOrError("MACHINE_FOLDER")
		if err != nil {
			return err
		}
	}
	return nil
}

func loadRequiredOptions(opts *Options) error {
	required := []struct {
		dest *string
		name string
	}{
		{&opts.Project, "PROJECT"},
		{&opts.Zone, "ZONE"},
		{&opts.DiskSize, "DISK_SIZE"},
		{&opts.DiskImage, "DISK_IMAGE"},
		{&opts.MachineType, "MACHINE_TYPE"},
	}

	for _, r := range required {
		val, err := fromEnvOrError(r.name)
		if err != nil {
			return err
		}
		*r.dest = val
	}

	return nil
}

func fromEnvOrError(name string) (string, error) {
	val := GetEnv(name)
	if val == "" {
		return "", fmt.Errorf(
			"couldn't find option %s in environment, please make sure %s is defined",
			name,
			name,
		)
	}

	return val, nil
}

func GetEnv(key string) string {
	providerName := os.Getenv("DEVPOD_PROVIDER")
	if providerName == "" {
		providerName = "GCLOUD"
	}
	providerName = strings.ToUpper(strings.ReplaceAll(providerName, "-", "_"))
	val := os.Getenv("DEVPOD_PROVIDER_" + providerName + "_" + key)
	if val != "" {
		return val
	}
	val = os.Getenv("DEVPOD_PROVIDER_" + key)
	if val != "" {
		return val
	}
	return os.Getenv(key)
}
