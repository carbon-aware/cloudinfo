package cloudinfo

import (
	"context"
	"fmt"
	"slices"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	// RegionLabel is the label key for the region of the node
	RegionLabel = "topology.kubernetes.io/region"
)

// NodeAttributes represents the attributes of the nodes in the cluster
type NodeAttributes struct {
	// List of unique regions found on nodes
	Regions []string

	// List of provider IDs found on nodes
	ProviderIDs []string
}

// DetectNodeCloudInfo detects cloud provider and region using node labels and spec.ProviderID.
func DetectNodeCloudInfo(ctx context.Context, client kubernetes.Interface) (*CloudInfo, error) {
	// Get node attributes
	attributes, err := GetNodeAttributes(ctx, client)

	if err != nil {
		return nil, err
	}

	// Parse provider from provider IDs
	provider, err := ParseProviderIDs(attributes.ProviderIDs)

	if err != nil {
		return nil, err
	}

	// Check that only one region is found
	if len(attributes.Regions) != 1 {
		return nil, fmt.Errorf("multiple regions found: %v", attributes.Regions)
	}

	return &CloudInfo{
		Provider: provider,
		Region:   attributes.Regions[0],
		Source:   "node-labels",
	}, nil
}

// GetNodeAttributes retrieves nodes and their attributes from the Kubernetes cluster.
func GetNodeAttributes(ctx context.Context, client kubernetes.Interface) (*NodeAttributes, error) {
	// Get node list
	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	if len(nodes.Items) == 0 {
		return nil, fmt.Errorf("no nodes found")
	}

	attributes := &NodeAttributes{}

	// Get unique regions and provider IDs
	for _, node := range nodes.Items {
		regionLabel := node.Labels[RegionLabel]
		providerID := node.Spec.ProviderID

		if regionLabel != "" && !slices.Contains(attributes.Regions, regionLabel) {
			attributes.Regions = append(attributes.Regions, regionLabel)
		}
		if providerID != "" {
			attributes.ProviderIDs = append(attributes.ProviderIDs, providerID)
		}
	}

	return attributes, nil
}

// ParseProviderIDs parses a list of provider IDs and returns the unique cloud provider names.
func ParseProviderIDs(providerIDs []string) (string, error) {
	providers := make(map[string]struct{})
	for _, providerID := range providerIDs {
		switch {
		case strings.HasPrefix(providerID, "aws://"):
			providers["aws"] = struct{}{}
		case strings.HasPrefix(providerID, "azure://"):
			providers["azure"] = struct{}{}
		case strings.HasPrefix(providerID, "gce://"):
			providers["gcp"] = struct{}{}
		default:
			return "", fmt.Errorf("provider ID format unknown: %s", providerID)
		}
	}
	result := make([]string, 0, len(providers))
	for p := range providers {
		result = append(result, p)
	}
	if len(result) > 1 {
		return "", fmt.Errorf("multiple cloud providers found: %s", strings.Join(result, ", "))
	}
	return result[0], nil
}

// ParseProviderID parses a provider ID and returns the cloud provider name.
func ParseProviderID(providerID string) (string, error) {
	if providerID == "" {
		return "", fmt.Errorf("empty provider ID")
	}

	parts := strings.SplitN(providerID, "://", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("unknown provider ID format: %s", providerID)
	}

	switch parts[0] {
	case "aws":
		return "aws", nil
	case "azure":
		return "azure", nil
	case "gce":
		return "gcp", nil
	default:
		return "", fmt.Errorf("unknown provider ID format: %s", providerID)
	}
}
