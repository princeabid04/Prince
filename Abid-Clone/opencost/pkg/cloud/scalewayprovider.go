package cloud

import (
	"errors"
	"fmt"
	"github.com/opencost/opencost/pkg/kubecost"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/opencost/opencost/pkg/clustercache"
	"github.com/opencost/opencost/pkg/env"
	"github.com/opencost/opencost/pkg/util"
	"github.com/opencost/opencost/pkg/util/json"

	"github.com/opencost/opencost/pkg/log"
	v1 "k8s.io/api/core/v1"

	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	InstanceAPIPricing = "Instance API Pricing"
)

type ScalewayPricing struct {
	NodesInfos map[string]*instance.ServerType
	PVCost     float64
}

type Scaleway struct {
	Clientset               clustercache.ClusterCache
	Config                  *ProviderConfig
	Pricing                 map[string]*ScalewayPricing
	DownloadPricingDataLock sync.RWMutex
}

func (c *Scaleway) DownloadPricingData() error {
	c.DownloadPricingDataLock.Lock()
	defer c.DownloadPricingDataLock.Unlock()

	// TODO wait for an official Pricing API from Scaleway
	// Let's use a static map and an old API

	if len(c.Pricing) != 0 {
		// Already initialized
		return nil
	}

	// PV pricing per AZ
	pvPrice := map[string]float64{
		"fr-par-1": 0.00011,
		"fr-par-2": 0.00011,
		"fr-par-3": 0.00032,
		"nl-ams-1": 0.00008,
		"nl-ams-2": 0.00008,
		"pl-waw-1": 0.00011,
	}

	c.Pricing = make(map[string]*ScalewayPricing)

	// The endpoint we are trying to hit does not have authentication
	client, err := scw.NewClient(scw.WithoutAuth())
	if err != nil {
		return err
	}

	instanceAPI := instance.NewAPI(client)

	for _, zone := range scw.AllZones {
		resp, err := instanceAPI.ListServersTypes(&instance.ListServersTypesRequest{Zone: zone})
		if err != nil {
			log.Errorf("Could not get Scaleway pricing data from instance API in zone %s: %+v", zone, err)
			continue
		}
		c.Pricing[zone.String()] = &ScalewayPricing{
			PVCost:     pvPrice[zone.String()],
			NodesInfos: map[string]*instance.ServerType{},
		}

		for name, infos := range resp.Servers {
			c.Pricing[zone.String()].NodesInfos[name] = infos
		}
	}

	return nil
}

func (c *Scaleway) AllNodePricing() (interface{}, error) {
	c.DownloadPricingDataLock.RLock()
	defer c.DownloadPricingDataLock.RUnlock()
	return c.Pricing, nil
}

type scalewayKey struct {
	Labels map[string]string
}

func (k *scalewayKey) Features() string {
	instanceType, _ := util.GetInstanceType(k.Labels)
	zone, _ := util.GetZone(k.Labels)

	return zone + "," + instanceType
}

func (k *scalewayKey) GPUCount() int {
	return 0
}

func (k *scalewayKey) GPUType() string {
	instanceType, _ := util.GetInstanceType(k.Labels)
	if strings.HasPrefix(instanceType, "RENDER") || strings.HasPrefix(instanceType, "GPU") {
		return instanceType
	}
	return ""
}
func (k *scalewayKey) ID() string {
	return ""
}

func (c *Scaleway) NodePricing(key Key) (*Node, error) {
	c.DownloadPricingDataLock.RLock()
	defer c.DownloadPricingDataLock.RUnlock()

	// There is only the zone and the instance ID in the providerID, hence we must use the features
	split := strings.Split(key.Features(), ",")
	if pricing, ok := c.Pricing[split[0]]; ok {
		if info, ok := pricing.NodesInfos[split[1]]; ok {
			return &Node{
				Cost:        fmt.Sprintf("%f", info.HourlyPrice),
				PricingType: DefaultPrices,
				VCPU:        fmt.Sprintf("%d", info.Ncpus),
				RAM:         fmt.Sprintf("%d", info.RAM),
				// This is tricky, as instances can have local volumes or not
				Storage:      fmt.Sprintf("%d", info.PerVolumeConstraint.LSSD.MinSize),
				GPU:          fmt.Sprintf("%d", info.Gpu),
				InstanceType: split[1],
				Region:       split[0],
				GPUName:      key.GPUType(),
			}, nil

		}

	}
	return nil, fmt.Errorf("Unable to find node pricing matching thes features `%s`", key.Features())
}

func (c *Scaleway) LoadBalancerPricing() (*LoadBalancer, error) {
	// Different LB types, lets take the cheaper for now, we can't get the type
	// without a service specifying the type in the annotations
	return &LoadBalancer{
		Cost: 0.014,
	}, nil
}

func (c *Scaleway) NetworkPricing() (*Network, error) {
	// it's free baby!
	return &Network{
		ZoneNetworkEgressCost:     0,
		RegionNetworkEgressCost:   0,
		InternetNetworkEgressCost: 0,
	}, nil
}

func (c *Scaleway) GetKey(l map[string]string, n *v1.Node) Key {
	return &scalewayKey{
		Labels: l,
	}
}

type scalewayPVKey struct {
	Labels                 map[string]string
	StorageClassName       string
	StorageClassParameters map[string]string
	Name                   string
	Zone                   string
}

func (key *scalewayPVKey) ID() string {
	return ""
}

func (key *scalewayPVKey) GetStorageClass() string {
	return key.StorageClassName
}

func (key *scalewayPVKey) Features() string {
	// Only 1 type of PV for now
	return key.Zone
}

func (c *Scaleway) GetPVKey(pv *v1.PersistentVolume, parameters map[string]string, defaultRegion string) PVKey {
	// the csi volume handle is the form <az>/<volume-id>
	zone := strings.Split(pv.Spec.CSI.VolumeHandle, "/")[0]
	return &scalewayPVKey{
		Labels:                 pv.Labels,
		StorageClassName:       pv.Spec.StorageClassName,
		StorageClassParameters: parameters,
		Name:                   pv.Name,
		Zone:                   zone,
	}
}

func (c *Scaleway) PVPricing(pvk PVKey) (*PV, error) {
	c.DownloadPricingDataLock.RLock()
	defer c.DownloadPricingDataLock.RUnlock()

	pricing, ok := c.Pricing[pvk.Features()]
	if !ok {
		log.Infof("Persistent Volume pricing not found for %s: %s", pvk.GetStorageClass(), pvk.Features())
		return &PV{}, nil
	}
	return &PV{
		Cost:  fmt.Sprintf("%f", pricing.PVCost),
		Class: pvk.GetStorageClass(),
	}, nil
}

func (c *Scaleway) ServiceAccountStatus() *ServiceAccountStatus {
	return &ServiceAccountStatus{
		Checks: []*ServiceAccountCheck{},
	}
}

func (*Scaleway) ClusterManagementPricing() (string, float64, error) {
	return "", 0.0, nil
}

func (c *Scaleway) CombinedDiscountForNode(instanceType string, isPreemptible bool, defaultDiscount, negotiatedDiscount float64) float64 {
	return 1.0 - ((1.0 - defaultDiscount) * (1.0 - negotiatedDiscount))
}

func (c *Scaleway) Regions() []string {
	// These are zones but hey, its 2022
	zones := []string{}
	for _, zone := range scw.AllZones {
		zones = append(zones, zone.String())
	}
	return zones
}

func (*Scaleway) ApplyReservedInstancePricing(map[string]*Node) {}

func (*Scaleway) GetAddresses() ([]byte, error) {
	return nil, nil
}

func (*Scaleway) GetDisks() ([]byte, error) {
	return nil, nil
}

func (*Scaleway) GetOrphanedResources() ([]OrphanedResource, error) {
	return nil, errors.New("not implemented")
}

func (scw *Scaleway) ClusterInfo() (map[string]string, error) {
	remoteEnabled := env.IsRemoteEnabled()

	m := make(map[string]string)
	m["name"] = "Scaleway Cluster #1"
	c, err := scw.GetConfig()
	if err != nil {
		return nil, err
	}
	if c.ClusterName != "" {
		m["name"] = c.ClusterName
	}
	m["provider"] = kubecost.ScalewayProvider
	m["remoteReadEnabled"] = strconv.FormatBool(remoteEnabled)
	m["id"] = env.GetClusterID()
	return m, nil

}

func (c *Scaleway) UpdateConfigFromConfigMap(a map[string]string) (*CustomPricing, error) {
	return c.Config.UpdateFromMap(a)
}

func (c *Scaleway) UpdateConfig(r io.Reader, updateType string) (*CustomPricing, error) {
	defer c.DownloadPricingData()

	return c.Config.Update(func(c *CustomPricing) error {
		a := make(map[string]interface{})
		err := json.NewDecoder(r).Decode(&a)
		if err != nil {
			return err
		}
		for k, v := range a {
			kUpper := strings.Title(k) // Just so we consistently supply / receive the same values, uppercase the first letter.
			vstr, ok := v.(string)
			if ok {
				err := SetCustomPricingField(c, kUpper, vstr)
				if err != nil {
					return err
				}
			} else {
				return fmt.Errorf("type error while updating config for %s", kUpper)
			}
		}

		if env.IsRemoteEnabled() {
			err := UpdateClusterMeta(env.GetClusterID(), c.ClusterName)
			if err != nil {
				return err
			}
		}

		return nil
	})
}
func (scw *Scaleway) GetConfig() (*CustomPricing, error) {
	c, err := scw.Config.GetCustomPricingData()
	if err != nil {
		return nil, err
	}
	if c.Discount == "" {
		c.Discount = "0%"
	}
	if c.NegotiatedDiscount == "" {
		c.NegotiatedDiscount = "0%"
	}
	if c.CurrencyCode == "" {
		c.CurrencyCode = "EUR"
	}
	return c, nil
}

func (*Scaleway) GetLocalStorageQuery(window, offset time.Duration, rate bool, used bool) string {
	return ""
}

func (scw *Scaleway) GetManagementPlatform() (string, error) {
	nodes := scw.Clientset.GetAllNodes()

	if len(nodes) > 0 {
		n := nodes[0]
		if _, ok := n.Labels["k8s.scaleway.com/kapsule"]; ok {
			return "kapsule", nil
		}
		if _, ok := n.Labels["kops.k8s.io/instancegroup"]; ok {
			return "kops", nil
		}
	}
	return "", nil
}

func (c *Scaleway) PricingSourceStatus() map[string]*PricingSource {
	return map[string]*PricingSource{
		InstanceAPIPricing: &PricingSource{
			Name:      InstanceAPIPricing,
			Enabled:   true,
			Available: true,
		},
	}
}
