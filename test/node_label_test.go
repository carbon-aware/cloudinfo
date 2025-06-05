package test

import (
	"context"

	"github.com/carbon-aware/cloudinfo/pkg/cloudinfo"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

var _ = ginkgo.Describe("Node Label Detection", func() {
	var (
		ctx    context.Context
		client *fake.Clientset
	)

	ginkgo.BeforeEach(func() {
		ctx = context.Background()
		client = fake.NewSimpleClientset()
	})

	ginkgo.Context("when parsing provider IDs", func() {
		ginkgo.It("should parse AWS provider ID", func() {
			provider, err := cloudinfo.ParseProviderID("aws:///us-west-2a/i-1234567890abcdef0")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(provider).To(gomega.Equal("aws"))
		})

		ginkgo.It("should parse Azure provider ID", func() {
			provider, err := cloudinfo.ParseProviderID("azure:///subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/myResourceGroup/providers/Microsoft.Compute/virtualMachines/myVM")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(provider).To(gomega.Equal("azure"))
		})

		ginkgo.It("should parse GCP provider ID", func() {
			provider, err := cloudinfo.ParseProviderID("gce://my-project/us-central1-a/my-instance")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(provider).To(gomega.Equal("gcp"))
		})

		ginkgo.It("should handle empty provider ID", func() {
			_, err := cloudinfo.ParseProviderID("")
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.Equal("empty provider ID"))
		})

		ginkgo.It("should handle unknown provider ID format", func() {
			_, err := cloudinfo.ParseProviderID("unknown://format")
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.Equal("unknown provider ID format: unknown://format"))
		})
	})

	ginkgo.Describe("DetectNodeCloudInfo", func() {
		ginkgo.Context("when multiple cloud providers are found", func() {
			ginkgo.BeforeEach(func() {
				// Create nodes with different provider IDs
				_, err := client.CoreV1().Nodes().Create(ctx, &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node1",
						Labels: map[string]string{
							"topology.kubernetes.io/region": "us-west-2",
						},
						Annotations: map[string]string{
							"csi.volume.kubernetes.io/nodeid": `{"ebs.csi.aws.com":"i-1234567890abcdef0"}`,
						},
					},
					Spec: corev1.NodeSpec{
						ProviderID: "aws:///us-west-2a/i-1234567890abcdef0",
					},
				}, metav1.CreateOptions{})
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				_, err = client.CoreV1().Nodes().Create(ctx, &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node2",
						Labels: map[string]string{
							"topology.kubernetes.io/region": "eastus",
						},
						Annotations: map[string]string{
							"csi.volume.kubernetes.io/nodeid": `{"disk.csi.azure.com":"/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/myResourceGroup/providers/Microsoft.Compute/virtualMachines/myVM"}`,
						},
					},
					Spec: corev1.NodeSpec{
						ProviderID: "azure:///subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/myResourceGroup/providers/Microsoft.Compute/virtualMachines/myVM",
					},
				}, metav1.CreateOptions{})
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})

			ginkgo.It("should return an error", func() {
				_, err := cloudinfo.DetectNodeCloudInfo(ctx, client)
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.HavePrefix("multiple cloud providers found:"))
			})
		})

		ginkgo.Context("when multiple regions are found", func() {
			ginkgo.BeforeEach(func() {
				// Create nodes with different regions
				_, err := client.CoreV1().Nodes().Create(ctx, &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node1",
						Labels: map[string]string{
							"topology.kubernetes.io/region": "us-west-2",
						},
						Annotations: map[string]string{
							"csi.volume.kubernetes.io/nodeid": `{"ebs.csi.aws.com":"i-1234567890abcdef0"}`,
						},
					},
					Spec: corev1.NodeSpec{
						ProviderID: "aws:///us-west-2a/i-1234567890abcdef0",
					},
				}, metav1.CreateOptions{})
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				_, err = client.CoreV1().Nodes().Create(ctx, &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node2",
						Labels: map[string]string{
							"topology.kubernetes.io/region": "us-east-1",
						},
						Annotations: map[string]string{
							"csi.volume.kubernetes.io/nodeid": `{"ebs.csi.aws.com":"i-0987654321fedcba0"}`,
						},
					},
					Spec: corev1.NodeSpec{
						ProviderID: "aws:///us-east-1a/i-0987654321fedcba0",
					},
				}, metav1.CreateOptions{})
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})

			ginkgo.It("should return an error", func() {
				_, err := cloudinfo.DetectNodeCloudInfo(ctx, client)
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.Equal("multiple regions found: [us-west-2 us-east-1]"))
			})
		})

		ginkgo.Context("when provider ID format is unknown", func() {
			ginkgo.BeforeEach(func() {
				_, err := client.CoreV1().Nodes().Create(ctx, &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node1",
						Labels: map[string]string{
							"topology.kubernetes.io/region": "us-west-2",
						},
					},
					Spec: corev1.NodeSpec{
						ProviderID: "unknown://format",
					},
				}, metav1.CreateOptions{})
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})

			ginkgo.It("should return an error", func() {
				_, err := cloudinfo.DetectNodeCloudInfo(ctx, client)
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.Equal("provider ID format unknown: unknown://format"))
			})
		})

		ginkgo.Context("when provider and region are correctly set", func() {
			ginkgo.BeforeEach(func() {
				_, err := client.CoreV1().Nodes().Create(ctx, &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node1",
						Labels: map[string]string{
							"topology.kubernetes.io/region": "us-west-2",
						},
						Annotations: map[string]string{
							"csi.volume.kubernetes.io/nodeid": `{"ebs.csi.aws.com":"i-1234567890abcdef0"}`,
						},
					},
					Spec: corev1.NodeSpec{
						ProviderID: "aws:///us-west-2a/i-1234567890abcdef0",
					},
				}, metav1.CreateOptions{})
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})

			ginkgo.It("should return provider and region", func() {
				info, err := cloudinfo.DetectNodeCloudInfo(ctx, client)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(info.Provider).To(gomega.Equal("aws"))
				gomega.Expect(info.Region).To(gomega.Equal("us-west-2"))
				gomega.Expect(info.Source).To(gomega.Equal("node-labels"))
			})
		})
	})
})
