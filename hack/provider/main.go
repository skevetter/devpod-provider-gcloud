package main

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
)

const (
	providerName = "gcloud"
	githubOwner  = "skevetter"
	githubRepo   = "devpod-provider-gcloud"
)

type Provider struct {
	Name         string            `yaml:"name"`
	Version      string            `yaml:"version"`
	Description  string            `yaml:"description"`
	Icon         string            `yaml:"icon"`
	OptionGroups []OptionGroup     `yaml:"optionGroups"`
	Options      Options           `yaml:"options"`
	Agent        Agent             `yaml:"agent"`
	Binaries     Binaries          `yaml:"binaries"`
	Exec         map[string]string `yaml:"exec"`
}

type OptionGroup struct {
	Name    string   `yaml:"name"`
	Options []string `yaml:"options"`
}

type Options map[string]Option

type Option struct {
	Description string   `yaml:"description,omitempty"`
	Required    bool     `yaml:"required,omitempty"`
	Default     string   `yaml:"default,omitempty"`
	Command     string   `yaml:"command,omitempty"`
	Suggestions []string `yaml:"suggestions,omitempty"`
	Local       bool     `yaml:"local,omitempty"`
	Hidden      bool     `yaml:"hidden,omitempty"`
	Cache       string   `yaml:"cache,omitempty"`
}

type Agent struct {
	Path                    string         `yaml:"path"`
	InactivityTimeout       string         `yaml:"inactivityTimeout"`
	InjectGitCredentials    string         `yaml:"injectGitCredentials"`
	InjectDockerCredentials string         `yaml:"injectDockerCredentials"`
	Binaries                map[string]any `yaml:"binaries"`
	Exec                    map[string]any `yaml:"exec"`
}

type Binaries struct {
	GCloudProvider []Binary `yaml:"GCLOUD_PROVIDER"`
}

type Binary struct {
	OS       string `yaml:"os"`
	Arch     string `yaml:"arch"`
	Path     string `yaml:"path"`
	Checksum string `yaml:"checksum"`
}

type buildConfig struct {
	version     string
	projectRoot string
	isRelease   bool
	checksums   map[string]string
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) != 2 {
		return fmt.Errorf("expected version as argument")
	}

	cfg, err := newBuildConfig(os.Args[1])
	if err != nil {
		return err
	}

	provider, err := buildProvider(cfg)
	if err != nil {
		return err
	}

	output, err := yaml.Marshal(provider)
	if err != nil {
		return fmt.Errorf("marshal yaml: %w", err)
	}

	fmt.Print(string(output))
	return nil
}

func newBuildConfig(version string) (*buildConfig, error) {
	checksums, err := parseChecksums("./dist/checksums.txt")
	if err != nil {
		return nil, fmt.Errorf("parse checksums: %w", err)
	}

	projectRoot := os.Getenv("PROJECT_ROOT")
	if projectRoot == "" {
		owner := getEnvOrDefault("GITHUB_OWNER", githubOwner)
		projectRoot = fmt.Sprintf("https://github.com/%s/%s/releases/download/%s", owner, githubRepo, version)
	}

	isRelease := strings.Contains(projectRoot, "github.com") && strings.Contains(projectRoot, "/releases/")

	return &buildConfig{
		version:     version,
		projectRoot: projectRoot,
		isRelease:   isRelease,
		checksums:   checksums,
	}, nil
}

func buildProvider(cfg *buildConfig) (Provider, error) {
	binaries, err := buildBinaries(cfg, allPlatforms())
	if err != nil {
		return Provider{}, err
	}
	agent, err := buildAgent(cfg)
	if err != nil {
		return Provider{}, err
	}
	return Provider{
		Name:         providerName,
		Version:      cfg.version,
		Description:  "DevPod on Google Cloud",
		Icon:         "https://devpod.sh/assets/gcp.svg",
		OptionGroups: buildOptionGroups(),
		Options:      buildOptions(),
		Agent:        agent,
		Binaries:     binaries,
		Exec: map[string]string{
			"init":    "${GCLOUD_PROVIDER} init",
			"command": "${GCLOUD_PROVIDER} command",
			"create":  "${GCLOUD_PROVIDER} create",
			"delete":  "${GCLOUD_PROVIDER} delete",
			"start":   "${GCLOUD_PROVIDER} start",
			"stop":    "${GCLOUD_PROVIDER} stop",
			"status":  "${GCLOUD_PROVIDER} status",
		},
	}, nil
}

func buildOptionGroups() []OptionGroup {
	return []OptionGroup{
		{
			Name:    "GCloud options",
			Options: []string{"DISK_SIZE", "DISK_IMAGE", "MACHINE_TYPE", "NETWORK", "SUBNETWORK", "TAG", "SERVICE_ACCOUNT", "PUBLIC_IP_ENABLED"},
		},
		{
			Name:    "Agent options",
			Options: []string{"AGENT_PATH", "INACTIVITY_TIMEOUT", "INJECT_DOCKER_CREDENTIALS", "INJECT_GIT_CREDENTIALS"},
		},
	}
}

func buildOptions() Options {
	return Options{
		"PROJECT": {
			Description: "The project id to use.",
			Required:    true,
			Command:     `gcloud config list --quiet --verbosity=error --format "value(core.project)" 2>/dev/null || true`,
		},
		"ZONE": {
			Description: "The google cloud zone to create the VM in. E.g. europe-west1-d",
			Required:    true,
			Command: `GCLOUD_ZONE=$(gcloud config list --quiet --verbosity=error --format "value(compute.zone)" 2>/dev/null || true)
if [ -z "$GCLOUD_ZONE" ]; then
  echo "europe-west2-b"
else
  echo $GCLOUD_ZONE
fi`,
			Suggestions: []string{
				"asia-east1-a", "asia-east1-b", "asia-east1-c",
				"asia-east2-a", "asia-east2-b", "asia-east2-c",
				"asia-northeast1-a", "asia-northeast1-c",
				"asia-northeast2-b", "asia-northeast3-b",
				"asia-south1-a", "asia-south1-b",
				"asia-southeast1-a",
				"europe-north1-a", "europe-north1-b", "europe-north1-c",
				"europe-west1-b", "europe-west1-c", "europe-west1-d",
				"europe-west2-a", "europe-west2-b", "europe-west2-c",
				"europe-west3-a", "europe-west3-b", "europe-west3-c",
				"europe-west4-a", "europe-west4-b", "europe-west4-c",
				"europe-west9-a", "europe-west9-b", "europe-west9-c",
				"me-central1-a", "me-central1-b", "me-central1-c",
				"me-west1-a", "me-west1-b", "me-west1-c",
				"northamerica-northeast1-a", "northamerica-northeast1-b", "northamerica-northeast1-c",
				"southamerica-east1-a", "southamerica-east1-b", "southamerica-east1-c",
				"southamerica-west1-a", "southamerica-west1-b", "southamerica-west1-c",
				"us-central1-a", "us-central1-b", "us-central1-f",
				"us-east1-b", "us-east1-c", "us-east1-d",
				"us-east4-a", "us-east4-b", "us-east4-c",
				"us-south1-a", "us-south1-b", "us-south1-c",
				"us-west1-a", "us-west1-b", "us-west1-c",
				"us-west2-a", "us-west2-b", "us-west2-c",
				"us-west4-a", "us-west4-b", "us-west4-c",
			},
		},
		"NETWORK": {
			Description: "The network id to use.",
		},
		"SUBNETWORK": {
			Description: "The subnetwork id to use.",
		},
		"TAG": {
			Description: "A tag to attach to the instance.",
			Default:     "devpod",
		},
		"DISK_SIZE": {
			Description: "The disk size to use (GB).",
			Default:     "40",
		},
		"DISK_IMAGE": {
			Description: "The disk image to use.",
			Default:     "projects/cos-cloud/global/images/cos-101-17162-127-5",
		},
		"SERVICE_ACCOUNT": {
			Description: "A service account to attach",
			Default:     "",
		},
		"PUBLIC_IP_ENABLED": {
			Description: "Use a public ip to access the instance",
			Default:     "true",
		},
		"MACHINE_TYPE": {
			Description: "The machine type to use.",
			Default:     "c2-standard-4",
			Suggestions: []string{
				"f1-micro", "e2-small", "e2-medium",
				"n2-standard-2", "n2-standard-4", "n2-standard-8", "n2-standard-16",
				"n2-highcpu-8", "n2-highcpu-16",
				"c2-standard-4", "c2-standard-8", "c2-standard-16", "c2-standard-30",
				"g2-standard-4", "g2-standard-8", "g2-standard-12", "g2-standard-16",
				"a2-highgpu-1g", "a2-highgpu-2g",
			},
		},
		"INACTIVITY_TIMEOUT": {
			Description: "If defined, will automatically stop the VM after the inactivity period.",
			Default:     "5m",
		},
		"INJECT_GIT_CREDENTIALS": {
			Description: "If DevPod should inject git credentials into the remote host.",
			Default:     "true",
		},
		"INJECT_DOCKER_CREDENTIALS": {
			Description: "If DevPod should inject docker credentials into the remote host.",
			Default:     "true",
		},
		"AGENT_PATH": {
			Description: "The path where to inject the DevPod agent to.",
			Default:     "/var/lib/toolbox/devpod",
		},
		"GCLOUD_PROVIDER_TOKEN": {
			Local:       true,
			Hidden:      true,
			Cache:       "5m",
			Description: "The Google Cloud auth token to use",
			Command:     "${GCLOUD_PROVIDER} token",
		},
	}
}

func buildAgent(cfg *buildConfig) (Agent, error) {
	linuxBins, err := buildBinaries(cfg, linuxPlatforms())
	if err != nil {
		return Agent{}, err
	}
	return Agent{
		Path:                    "${AGENT_PATH}",
		InactivityTimeout:       "${INACTIVITY_TIMEOUT}",
		InjectGitCredentials:    "${INJECT_GIT_CREDENTIALS}",
		InjectDockerCredentials: "${INJECT_DOCKER_CREDENTIALS}",
		Binaries: map[string]any{
			"GCLOUD_PROVIDER": linuxBins.GCloudProvider,
		},
		Exec: map[string]any{
			"shutdown": "${GCLOUD_PROVIDER} stop --raw",
		},
	}, nil
}

func buildBinaries(cfg *buildConfig, platforms []string) (Binaries, error) {
	list, err := buildBinaryList(cfg, platforms)
	if err != nil {
		return Binaries{}, err
	}
	return Binaries{GCloudProvider: list}, nil
}

func buildBinaryList(cfg *buildConfig, platforms []string) ([]Binary, error) {
	result := make([]Binary, 0, len(platforms))
	for _, platform := range platforms {
		binary, err := buildBinary(cfg, platform)
		if err != nil {
			return nil, err
		}
		result = append(result, binary)
	}
	return result, nil
}

func buildBinary(cfg *buildConfig, platform string) (Binary, error) {
	os, arch, ok := strings.Cut(platform, "/")
	if !ok {
		return Binary{}, fmt.Errorf("invalid platform %q", platform)
	}

	path, err := buildBinaryPath(cfg, platform, os, arch)
	if err != nil {
		return Binary{}, err
	}

	filename := buildFilename(os, arch)
	checksum, ok := cfg.checksums[filename]
	if !ok || checksum == "" {
		return Binary{}, fmt.Errorf("missing checksum for %s", filename)
	}

	return Binary{
		OS:       os,
		Arch:     arch,
		Path:     path,
		Checksum: checksum,
	}, nil
}

func buildBinaryPath(cfg *buildConfig, platform, os, arch string) (string, error) {
	dir := buildDir(platform)
	if dir == "" {
		return "", fmt.Errorf("unsupported platform %q", platform)
	}

	basePath, err := resolveBasePath(cfg, dir)
	if err != nil {
		return "", err
	}

	filename := buildFilename(os, arch)
	return joinPath(basePath, filename)
}

func resolveBasePath(cfg *buildConfig, dir string) (string, error) {
	if cfg.isRelease {
		return cfg.projectRoot, nil
	}

	if strings.HasPrefix(cfg.projectRoot, "http://") || strings.HasPrefix(cfg.projectRoot, "https://") {
		return joinURLPath(cfg.projectRoot, dir)
	}

	absPath, err := filepath.Abs(cfg.projectRoot)
	if err != nil {
		return "", fmt.Errorf("abs PROJECT_ROOT: %w", err)
	}
	return filepath.Join(absPath, dir), nil
}

func joinPath(basePath, filename string) (string, error) {
	if strings.HasPrefix(basePath, "http://") || strings.HasPrefix(basePath, "https://") {
		return joinURLPath(basePath, filename)
	}
	return filepath.Join(basePath, filename), nil
}

func joinURLPath(base, elem string) (string, error) {
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("parse URL: %w", err)
	}
	joined, err := url.JoinPath(parsed.String(), elem)
	if err != nil {
		return "", fmt.Errorf("join URL path: %w", err)
	}
	return joined, nil
}

func buildFilename(os, arch string) string {
	filename := fmt.Sprintf("devpod-provider-%s-%s-%s", providerName, os, arch)
	if os == "windows" {
		filename += ".exe"
	}
	return filename
}

func buildDir(platform string) string {
	dirs := map[string]string{
		"linux/amd64":   "build_linux_amd64_v1",
		"linux/arm64":   "build_linux_arm64_v8.0",
		"darwin/amd64":  "build_darwin_amd64_v1",
		"darwin/arm64":  "build_darwin_arm64_v8.0",
		"windows/amd64": "build_windows_amd64_v1",
	}
	return dirs[platform]
}

func allPlatforms() []string {
	return []string{"linux/amd64", "linux/arm64", "darwin/amd64", "darwin/arm64", "windows/amd64"}
}

func linuxPlatforms() []string {
	return []string{"linux/amd64", "linux/arm64"}
}

func parseChecksums(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	checksums := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if checksum, filename, ok := strings.Cut(scanner.Text(), " "); ok {
			checksums[strings.TrimSpace(filename)] = checksum
		}
	}

	return checksums, scanner.Err()
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
