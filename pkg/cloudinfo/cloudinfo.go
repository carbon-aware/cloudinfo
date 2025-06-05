// Package cloudinfo provides functions to detect cloud provider and region of a Kubernetes cluster.
package cloudinfo

import (
	"context"
	"fmt"

	"k8s.io/client-go/kubernetes"
)

// DetectCloudInfo detects cloud provider and region using the specified methods.
func DetectCloudInfo(ctx context.Context, client kubernetes.Interface, opts Options) (*CloudInfo, error) {
	switch {
	case opts.UseNodeLabels:
		return DetectNodeCloudInfo(ctx, client)
	case opts.UseIMDS:
		return DetectIMDSCloudInfo(ctx)
	default:
		return nil, fmt.Errorf("no cloud info detection method specified")
	}
}
