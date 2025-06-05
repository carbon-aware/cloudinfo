package cloudinfo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// IMDSClient is an interface for making HTTP requests to IMDS endpoints.
type IMDSClient interface {
	Do(*http.Request) (*http.Response, error)
}

// DefaultIMDSClient returns a new HTTP client configured for IMDS requests.
func DefaultIMDSClient() *http.Client {
	return &http.Client{
		Timeout: 5 * time.Second,
	}
}

// IMDSConfig holds the configuration for IMDS endpoints.
type IMDSConfig struct {
	AWSEndpoint   string
	AzureEndpoint string
	GCPEndpoint   string
}

// DefaultIMDSConfig returns the default IMDS configuration.
func DefaultIMDSConfig() IMDSConfig {
	return IMDSConfig{
		AWSEndpoint:   "http://169.254.169.254/latest/meta-data/placement/region",
		AzureEndpoint: "http://169.254.169.254/metadata/instance/compute/location?api-version=2021-02-01",
		GCPEndpoint:   "http://metadata.google.internal/computeMetadata/v1/instance/zone",
	}
}

// DetectIMDSCloudInfo detects cloud provider and region using IMDS.
func DetectIMDSCloudInfo(ctx context.Context) (*CloudInfo, error) {
	return DetectIMDSCloudInfoWithClient(ctx, DefaultIMDSClient(), DefaultIMDSConfig())
}

// DetectIMDSCloudInfoWithClient detects cloud provider and region using IMDS with a custom client.
func DetectIMDSCloudInfoWithClient(ctx context.Context, client IMDSClient, config IMDSConfig) (*CloudInfo, error) {
	// Try AWS first
	req, err := http.NewRequestWithContext(ctx, "GET", config.AWSEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS IMDS request: %w", err)
	}
	resp, err := client.Do(req)
	if err == nil && resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()
		region, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read AWS region: %w", err)
		}
		return &CloudInfo{
			Provider: "aws",
			Region:   string(region),
			Source:   "imds",
		}, nil
	}

	// Try Azure next
	req, err = http.NewRequestWithContext(ctx, "GET", config.AzureEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure IMDS request: %w", err)
	}
	req.Header.Set("Metadata", "true")
	resp, err = client.Do(req)
	if err == nil && resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()
		var result struct {
			Location string `json:"location"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode Azure location: %w", err)
		}
		return &CloudInfo{
			Provider: "azure",
			Region:   result.Location,
			Source:   "imds",
		}, nil
	}

	// Try GCP last
	req, err = http.NewRequestWithContext(ctx, "GET", config.GCPEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCP IMDS request: %w", err)
	}
	req.Header.Set("Metadata-Flavor", "Google")
	resp, err = client.Do(req)
	if err == nil && resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()
		zone, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read GCP zone: %w", err)
		}
		// Extract region from zone (e.g., "projects/123456789/zones/us-central1-a" -> "us-central1")
		parts := strings.Split(string(zone), "/")
		if len(parts) < 4 {
			return nil, fmt.Errorf("invalid GCP zone format: %s", zone)
		}
		zoneName := parts[len(parts)-1]
		regionParts := strings.Split(zoneName, "-")
		if len(regionParts) < 2 {
			return nil, fmt.Errorf("invalid GCP zone format: %s", zoneName)
		}
		region := strings.Join(regionParts[:len(regionParts)-1], "-")
		return &CloudInfo{
			Provider: "gcp",
			Region:   region,
			Source:   "imds",
		}, nil
	}

	return nil, fmt.Errorf("failed to detect cloud provider using IMDS")
}
