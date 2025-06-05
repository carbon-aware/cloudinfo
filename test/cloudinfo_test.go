package test

import (
	"context"

	"github.com/carbon-aware/cloudinfo/pkg/cloudinfo"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes/fake"
)

var _ = ginkgo.Describe("CloudInfo", func() {
	var ctx context.Context
	ginkgo.BeforeEach(func() {
		ctx = context.Background()
	})

	ginkgo.Context("when no detection method is specified", func() {
		ginkgo.It("should return an error", func() {
			client := fake.NewSimpleClientset()
			_, err := cloudinfo.DetectCloudInfo(ctx, client, cloudinfo.Options{})
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.Equal("no cloud info detection method specified"))
		})
	})

	ginkgo.Context("when both detection methods are specified", func() {
		ginkgo.It("should prioritize node labels over IMDS", func() {
			client := fake.NewSimpleClientset()
			_, err := cloudinfo.DetectCloudInfo(ctx, client, cloudinfo.Options{UseNodeLabels: true, UseIMDS: true})
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.Equal("no nodes found"))
		})
	})

	ginkgo.Context("when only IMDS is specified", func() {
		ginkgo.It("should attempt IMDS detection", func() {
			client := fake.NewSimpleClientset()
			_, err := cloudinfo.DetectCloudInfo(ctx, client, cloudinfo.Options{UseIMDS: true})
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.Equal("failed to detect cloud provider using IMDS"))
		})
	})

	// Detailed node label and IMDS tests are in their respective test files.
})
