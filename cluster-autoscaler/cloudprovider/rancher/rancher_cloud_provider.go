package rancher

import (
	"io"
	"os"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog"
)

const (
	// ProviderName is the cloud provider name for rancher
	ProviderName = "rancher"
)

type RancherCloudProvider struct{
	rancherManager *RancherManager
	resourceLimiter *cloudprovider.ResourceLimiter
}




func BuildRancher(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
	var cfg io.ReadCloser
	if opts.CloudConfig != "" {
		var err error
		cfg, err = os.Open(opts.CloudConfig)
		if err != nil {
			klog.Fatalf("Couldn't open cloud provider configuration %s: %#v", opts.CloudConfig, err)
		}
		defer cfg.Close()
	}

	manager, err := CreateRancherManager(cfg)
	if err != nil {
		klog.Fatalf("Failed to create Rancher Manager: %v", err)
	}

	provider, err := BuildRancherCloudProvider(manager, do, rl)
	if err != nil {
		klog.Fatalf("Failed to create Rancher cloud provider: %v", err)
	}
	return provider
}


// BuildRancherCloudProvider builds CloudProvider implementation for Rancher.
func BuildRancherCloudProvider(rancherManager *RancherManager,discoveryOpts cloudprovider.NodeGroupDiscoveryOptions, resourceLimiter *cloudprovider.ResourceLimiter) (cloudprovider.CloudProvider, error) {
	rancher := &RancherCloudProvider{
		rancherManager: rancherManager,
		resourceLimiter: resourceLimiter,
	}

	if err := rancher.rancherManager.fetchExplicitNodePools(discoveryOpts.NodeGroupSpecs); err != nil {
		return nil, err
	}
	return rancher, nil
}

func (rancher *RancherCloudProvider) Name() string { return ProviderName }

func (rancher *RancherCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	nodePools := rancher.rancherManager.getNodePools()
	ngs := make([]cloudprovider.NodeGroup, len(nodePools))
	for i, nodePool := range nodePools {
		ngs[i] = nodePool
	}
	return ngs
}

func (rancher *RancherCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	klog.V(6).Infof("Searching for node group for the node: %s", node.Name)
	ref := &RancherRef{
		Name: node.Name,
	}

	return rancher.rancherManager.GetNodePoolForInstance(ref)
}

func (rancher *RancherCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

func (rancher *RancherCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return []string{}, cloudprovider.ErrNotImplemented
}

func (rancher *RancherCloudProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string,
	taints []apiv1.Taint, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns struct containing limits (max, min) for resources (cores, memory etc.).
func (rancher *RancherCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return rancher.resourceLimiter, nil
}

// Cleanup cleans up all resources before the cloud provider is removed
func (rancher *RancherCloudProvider) Cleanup() error {
	rancher.rancherManager.Cleanup()
	return nil
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
// In particular the list of node groups returned by NodeGroups can change as a result of CloudProvider.Refresh().
func (rancher *RancherCloudProvider) Refresh() error {
	return rancher.rancherManager.Refresh()
}
