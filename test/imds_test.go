package test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/carbon-aware/cloudinfo/pkg/cloudinfo"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("IMDS Detection", func() {
	var server *httptest.Server
	var config cloudinfo.IMDSConfig

	ginkgo.BeforeEach(func() {
		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			// Default to 404 for unknown paths
			w.WriteHeader(http.StatusNotFound)
		}))

		// Configure endpoints to use the test server
		config = cloudinfo.IMDSConfig{
			AWSEndpoint:   server.URL + "/latest/meta-data/placement/region",
			AzureEndpoint: server.URL + "/metadata/instance/compute/location?api-version=2021-02-01",
			GCPEndpoint:   server.URL + "/computeMetadata/v1/instance/zone",
		}
	})

	ginkgo.AfterEach(func() {
		server.Close()
	})

	ginkgo.Context("when running on AWS", func() {
		ginkgo.BeforeEach(func() {
			server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.HasSuffix(r.URL.Path, "/latest/meta-data/placement/region") {
					_, err := w.Write([]byte("us-west-2"))
					gomega.Expect(err).NotTo(gomega.HaveOccurred())
					return
				}
				w.WriteHeader(http.StatusNotFound)
			})
		})

		ginkgo.It("should detect AWS region", func() {
			info, err := cloudinfo.DetectIMDSCloudInfoWithClient(context.Background(), server.Client(), config)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(info.Provider).To(gomega.Equal("aws"))
			gomega.Expect(info.Region).To(gomega.Equal("us-west-2"))
			gomega.Expect(info.Source).To(gomega.Equal("imds"))
		})
	})

	ginkgo.Context("when running on Azure", func() {
		ginkgo.BeforeEach(func() {
			server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.URL.Path, "/metadata/instance/compute/location") {
					w.Header().Set("Content-Type", "application/json")
					err := json.NewEncoder(w).Encode(map[string]string{
						"location": "eastus",
					})
					gomega.Expect(err).NotTo(gomega.HaveOccurred())
					return
				}
				w.WriteHeader(http.StatusNotFound)
			})
		})

		ginkgo.It("should detect Azure region", func() {
			info, err := cloudinfo.DetectIMDSCloudInfoWithClient(context.Background(), server.Client(), config)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(info.Provider).To(gomega.Equal("azure"))
			gomega.Expect(info.Region).To(gomega.Equal("eastus"))
			gomega.Expect(info.Source).To(gomega.Equal("imds"))
		})
	})

	ginkgo.Context("when running on GCP", func() {
		ginkgo.BeforeEach(func() {
			server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.URL.Path, "/computeMetadata/v1/instance/zone") {
					_, err := w.Write([]byte("projects/123456789/zones/us-central1-a"))
					gomega.Expect(err).NotTo(gomega.HaveOccurred())
					return
				}
				w.WriteHeader(http.StatusNotFound)
			})
		})

		ginkgo.It("should detect GCP region", func() {
			info, err := cloudinfo.DetectIMDSCloudInfoWithClient(context.Background(), server.Client(), config)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(info.Provider).To(gomega.Equal("gcp"))
			gomega.Expect(info.Region).To(gomega.Equal("us-central1"))
			gomega.Expect(info.Source).To(gomega.Equal("imds"))
		})
	})

	ginkgo.Context("when no IMDS is available", func() {
		ginkgo.It("should return an error", func() {
			_, err := cloudinfo.DetectIMDSCloudInfoWithClient(context.Background(), server.Client(), config)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.Equal("failed to detect cloud provider using IMDS"))
		})
	})
})
