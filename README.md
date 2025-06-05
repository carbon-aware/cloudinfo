# cloudinfo

A Go package to detect the cloud provider and region for a Kubernetes cluster or controller node.

## Features

- Node label inspection for Kubernetes clusters
- Cloud metadata services (AWS, GCP, Azure) support
- Configurable detection methods
- Comprehensive test coverage
- Production-ready error handling

## Installation

```bash
go get github.com/carbon-aware/cloudinfo
```

## Usage

```go
package main

import (
    "context"
    "log"

    "github.com/carbon-aware/cloudinfo"
    "k8s.io/client-go/kubernetes"
)

func main() {
    // Create Kubernetes client
    client, err := kubernetes.NewForConfig(config)
    if err != nil {
        log.Fatal(err)
    }

    // Configure detection options
    opts := cloudinfo.Options{
        UseNodeLabels: true,  // Try to detect from node labels first
        UseIMDS:      false
    }

    // Detect cloud info
    info, err := cloudinfo.DetectCloudInfo(context.Background(), client, opts)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Detected cloud provider: %s, region: %s (source: %s)",
        info.Provider, info.Region, info.Source)
}
```

## Detection Methods

### Node Label Detection

The package can detect cloud provider and region information from Kubernetes node labels and provider IDs. This is the preferred method for Kubernetes clusters.

### IMDS Detection

For non-Kubernetes environments or as a fallback, the package can detect cloud information using cloud provider metadata services:

- AWS: http://169.254.169.254/latest/meta-data/placement/region
- Azure: http://169.254.169.254/metadata/instance/compute/location
- GCP: http://metadata.google.internal/computeMetadata/v1/instance/zone

## Development

### Prerequisites

- Go 1.21 or later
- Make (optional, for using Makefile)

### Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/carbon-aware/cloudinfo.git
   cd cloudinfo
   ```

2. Install development tools:
   ```bash
   make tools
   ```

### Common Tasks

- Run tests:
  ```bash
  make test
  ```

- Run linter:
  ```bash
  make lint
  ```

- Generate coverage report:
  ```bash
  make coverage
  ```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Testing

The project uses Ginkgo and Gomega for testing. The tests are organized into separate files for different functionalities:

- `test/cloudinfo_test.go`: Tests the high-level behavior of the `DetectCloudInfo` function.
- `test/node_label_test.go`: Tests the node label detection functionality.
- `test/imds_test.go`: Tests the IMDS detection functionality.

To run the tests, use the following command:

```bash
make test
```

This will execute all tests and report any failures.

## 📄 `LICENSE`

Apache 2.0
