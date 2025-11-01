package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

// TrivyResult represents the scan result from Trivy
type TrivyResult struct {
	Vulnerabilities []struct {
		VulnerabilityID  string `json:"VulnerabilityID"`
		PkgName          string `json:"PkgName"`
		InstalledVersion string `json:"InstalledVersion"`
		FixedVersion     string `json:"FixedVersion"`
		Severity         string `json:"Severity"`
		Description      string `json:"Description"`
	} `json:"Vulnerabilities"`
}

// ScanImage scans a container image using Trivy
func ScanImage(ctx context.Context, imageRef string) (*TrivyResult, error) {
	cmd := exec.CommandContext(ctx, "trivy", "image", "-f", "json", imageRef)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to scan image: %w", err)
	}

	var result TrivyResult
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse scan result: %w", err)
	}

	return &result, nil
}

// PatchImage creates a new image with security patches applied
func PatchImage(ctx context.Context, sourceImage, targetImage string) error {
	// Create a temporary script for building with Buildah
	script := fmt.Sprintf(`#!/bin/bash
set -e

# Start a new container build
container=$(buildah from %s)

# Apply updates based on the base image
buildah run $container /bin/sh -c '
    # Alpine
    if command -v apk >/dev/null 2>&1; then
        apk update && apk upgrade --no-cache
    # Debian/Ubuntu
    elif command -v apt-get >/dev/null 2>&1; then
        apt-get update && DEBIAN_FRONTEND=noninteractive apt-get upgrade -y
    # RHEL/CentOS
    elif command -v yum >/dev/null 2>&1; then
        yum update -y && yum upgrade -y
    fi'

# Commit the changes
buildah commit $container %s

# Clean up
buildah rm $container
`, sourceImage, targetImage)

	// Create a temporary file for the script
	tmpFile, err := os.CreateTemp("", "build-*.sh")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write the script to the file
	if err := os.WriteFile(tmpFile.Name(), []byte(script), 0755); err != nil {
		return fmt.Errorf("failed to write build script: %w", err)
	}

	// Execute the build script
	cmd := exec.CommandContext(ctx, "/bin/bash", tmpFile.Name())
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to build patched image: %s, error: %w", string(output), err)
	}

	return nil
}
