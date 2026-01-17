package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const provider = "gcloud"

func run() error {
	if len(os.Args) != 2 {
		return fmt.Errorf("expected version as argument")
	}

	content, err := os.ReadFile("./hack/provider/provider.yaml")
	if err != nil {
		return fmt.Errorf("read template: %w", err)
	}

	checksums, err := parseChecksums("./dist/checksums.txt")
	if err != nil {
		return fmt.Errorf("parse checksums: %w", err)
	}

	placeholders := map[string]string{
		"##CHECKSUM_LINUX_AMD64##":   fmt.Sprintf("devpod-provider-%s-linux-amd64", provider),
		"##CHECKSUM_LINUX_ARM64##":   fmt.Sprintf("devpod-provider-%s-linux-arm64", provider),
		"##CHECKSUM_DARWIN_AMD64##":  fmt.Sprintf("devpod-provider-%s-darwin-amd64", provider),
		"##CHECKSUM_DARWIN_ARM64##":  fmt.Sprintf("devpod-provider-%s-darwin-arm64", provider),
		"##CHECKSUM_WINDOWS_AMD64##": fmt.Sprintf("devpod-provider-%s-windows-amd64.exe", provider),
	}

	result := strings.ReplaceAll(string(content), "##VERSION##", os.Args[1])
	for placeholder, filename := range placeholders {
		checksum, ok := checksums[filename]
		if !ok {
			return fmt.Errorf("checksum not found for %s", filename)
		}
		result = strings.ReplaceAll(result, placeholder, checksum)
	}

	fmt.Print(result)
	return nil
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
		parts := strings.Fields(scanner.Text())
		if len(parts) == 2 {
			checksums[parts[1]] = parts[0]
		}
	}

	return checksums, scanner.Err()
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
