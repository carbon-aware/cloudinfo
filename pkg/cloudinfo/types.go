package cloudinfo

// CloudInfo represents the cloud provider and region of the cluster
type CloudInfo struct {
	Provider string // e.g. "aws", "gcp", "azure", or "unknown"
	Region   string
	Source   string // e.g. "node", "imds", "fallback"
}

// Options represents the options for detecting cloud info
type Options struct {
	// If should use Kubernetes node labels + spec.ProviderID
	UseNodeLabels bool
	// If should use IMDS to detect cloud info
	UseIMDS bool
}
