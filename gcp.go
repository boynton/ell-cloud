package cloud

import (
	//	"strings"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/boynton/ell"
	"github.com/go-ini/ini"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/storage/v1"
)

func gcpProvider(profile, region string) (Provider, error) {
	//precedence of arguments: parameter, then profile.metadata, then os.Getenv
	var err error
	gcp := &GCP{profile: profile}
	gcp.project = gcpMetadata("project/project-id")
	gcp.region = region
	if gcp.project != "" {
		//running in the cloud
		gcp.identity = gcpMetadata("instance/service-accounts/default/email")
		//userAttribute := gcpMetadata("instance/attributes/foobar")
		//zone := gcpMetadata("instance/zone") //i.e. "projects/1232132321312/zones/us-west1-a"
		//i := strings.LastIndex(z, "/")
		//if i < 0 {
		//	return fmt.Errorf("Cannot read zone/region from metadata")
		//}
	} else {
		//running locally
		confDir := os.Getenv("HOME") + "/.config/gcloud"
		if profile == "" {
			b, err := ioutil.ReadFile(confDir + "/active_config")
			if err == nil {
				profile = string(b)
			}
			if profile == "" {
				profile = os.Getenv("GCP_PROFILE")
				if profile == "" {
					profile = "default"
				}
			}
		}
		gcp.profile = profile
		confFile := confDir + "/configurations/config_" + gcp.profile
		if _, err := os.Stat(confFile); os.IsNotExist(err) {
			return nil, fmt.Errorf("No such profile '%s'\n", gcp.profile)
		}
		config, err := ini.Load(confFile)
		if err != nil {
			return nil, err
		}
		if gcp.project == "" {
			gcp.project = config.Section("core").Key("project").String()
		}
		if gcp.identity == "" {
			gcp.identity = config.Section("core").Key("account").String()
		}
		if gcp.region == "" {
			gcp.region = config.Section("compute").Key("region").String()
			if gcp.region == "" {
				gcp.region = os.Getenv("GCP_REGION")
			}
		}
	}
	ctx := context.Background()
	gcp.client, err = google.DefaultClient(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return nil, err
	}
	gcp.compute, err = compute.New(gcp.client)
	if err != nil {
		return nil, err
	}
	gcp.storage, err = storage.New(gcp.client)
	if err != nil {
		return nil, err
	}
	return gcp, nil
}

type GCP struct {
	profile  string
	identity string
	project  string
	region   string
	client   *http.Client
	compute  *compute.Service
	storage  *storage.Service
	network  *compute.Network
}

func (gcp *GCP) Name() string {
	return "gcp"
}

func (gcp *GCP) Identity() string {
	return gcp.identity
}

func (gcp *GCP) Account() string {
	return gcp.project
}

func (gcp *GCP) Profile() string {
	return gcp.profile
}

func (gcp *GCP) Region() string {
	return gcp.region
}

func (gcp *GCP) Describe(resource *ell.Object) (*ell.Object, error) {
	//note that the resource is a vacant DOM, so for example a zone could be given but not a name, etc. So this is a general query
	//for now, just use the name, err out on anything else
	switch resource.Type {
	case NetworkType:
		val, _ := ell.Get(resource, ell.Intern("name:"))
		if val != ell.Null {
			return gcp.DescribeNetwork(val.String())
		}
	case SubnetType:
		val, _ := ell.Get(resource, ell.Intern("name:"))
		if val != ell.Null {
			return gcp.DescribeSubnet(val.String())
		}
	case CloudType:
		nets, err := gcp.ListNetworks()
		if err != nil {
			return nil, err
		}
		repr := ell.MakeStruct(7)
		ell.Put(repr, ell.Intern("provider:"), ell.String("gcp"))
		if gcp.profile != "" {
			ell.Put(repr, ell.Intern("profile:"), ell.String(gcp.profile))
		}
		if gcp.identity != "" {
			ell.Put(repr, ell.Intern("identity:"), ell.String(gcp.identity))
		}
		if gcp.project != "" {
			ell.Put(repr, ell.Intern("account:"), ell.String(gcp.project))
		}
		if gcp.region != "" {
			ell.Put(repr, ell.Intern("region:"), ell.String(gcp.region))
		}
		ell.Put(repr, ell.Intern("networks:"), nets)
		return ell.Instance(CloudType, repr)
	}
	return nil, ell.Error(CloudErrorKey, "template does not provide enough info: ", resource)
}

func (gcp *GCP) Plan(resource *ell.Object) (*ell.Object, error) {
	return nil, ell.Error(CloudErrorKey, "gcp.Plan NYI")
}

func (gcp *GCP) Apply(resource *ell.Object) (*ell.Object, error) {
	return nil, ell.Error(CloudErrorKey, "gcp.Apply NYI")
}

func (gcp *GCP) Destroy(resource *ell.Object) (*ell.Object, error) {
	return nil, ell.Error(CloudErrorKey, "gcp.Destroy NYI")
}

func gcpMetadata(path string) string {
	base := "http://metadata.google.internal/computeMetadata/v1/"
	c := &http.Client{}
	c.Timeout = 2 * time.Second
	req, err := http.NewRequest("GET", base+path, nil)
	if err != nil {
		return ""
	}
	req.Header.Add("Metadata-Flavor", "Google")
	res, err := c.Do(req)
	if err != nil {
		return ""
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return ""
	}
	if res.StatusCode == 200 {
		return string(body)
	}
	return ""
}

func gcpInCloud() bool {
	if gcpMetadata("project/project-id") != "" {
		return true
	}
	return false
}

type GcpNetwork struct {
	name        string
	description string
}

func (net *GcpNetwork) Name() string {
	return net.name
}

func (net *GcpNetwork) Description() string {
	return net.description
}

//func (net *GcpNetwork) Subnets() ([]Subnet, error) {
//	return nil, nil
//}

func (gcp *GCP) netRepresentation(net *compute.Network) *ell.Object {
	repr := ell.MakeStruct(10)
	ell.Put(repr, ell.Intern("id:"), ell.String(fmt.Sprint(net.Id)))
	ell.Put(repr, ell.Intern("name:"), ell.String(net.Name))
	ell.Put(repr, ell.Intern("description:"), ell.String(net.Description))
	ell.Put(repr, ell.Intern("created:"), ell.String(net.CreationTimestamp))
	subnets := make([]*ell.Object, 0, len(net.Subnetworks))
	for _, subnetUrl := range net.Subnetworks {
		n := strings.LastIndex(subnetUrl, "/")
		subnetName := subnetUrl[n+1:]
		subnetRepr := ell.MakeStruct(10)
		ell.Put(subnetRepr, ell.Intern("name:"), ell.String(subnetName))
		subnet, _ := ell.Instance(SubnetType, subnetRepr)
		subnets = append(subnets, subnet)
	}
	ell.Put(repr, ell.Intern("subnets:"), ell.VectorFromElementsNoCopy(subnets))
	obj, _ := ell.Instance(NetworkType, repr)
	return obj
}

func (gcp *GCP) subnetRepresentation(subnet *compute.Subnetwork) *ell.Object {
	repr := ell.MakeStruct(10)
	ell.Put(repr, ell.Intern("name:"), ell.String(subnet.Name))
	ell.Put(repr, ell.Intern("cidr:"), ell.String(subnet.IpCidrRange))
	ell.Put(repr, ell.Intern("id:"), ell.String(fmt.Sprint(subnet.Id)))
	ell.Put(repr, ell.Intern("created:"), ell.String(subnet.CreationTimestamp))
	//ell.Put(repr, ell.Intern("network:"), ell.String(networkNameFromUrl(subnet.Network))) //?
	ell.Put(repr, ell.Intern("gateway:"), ell.String(subnet.GatewayAddress)) //?
	obj, _ := ell.Instance(SubnetType, repr)
	return obj
}

func (gcp *GCP) ListNetworks() (*ell.Object, error) {
	res, err := gcp.compute.Networks.List(gcp.project).Do()
	if err != nil {
		return nil, err
	}
	nets := make([]*ell.Object, 0, len(res.Items))
	for _, item := range res.Items {
		nets = append(nets, gcp.netRepresentation(item))
	}
	return ell.VectorFromElementsNoCopy(nets), nil
}

func (gcp *GCP) CreateNetwork(name string, cidr string, zones []string) (*ell.Object, error) {
	//for each zone in the region, create a subnet with the name "{network}-{zone}"
	return ell.Null, nil
}

func (gcp *GCP) DestroyNetwork(name string) (*ell.Object, error) {
	return ell.Null, nil
}

func (gcp *GCP) DescribeNetwork(name string) (*ell.Object, error) {
	net, err := gcp.compute.Networks.Get(gcp.project, name).Do()
	if err != nil {
		return nil, err
	}
	return gcp.netRepresentation(net), nil
	//	fmt.Println("network:", pretty(net))
	//	fmt.Println("subnets:", pretty(net.Subnetworks))
	//
	//	return ell.Null, nil
}

func (gcp *GCP) DescribeSubnet(name string) (*ell.Object, error) {
	subnet, err := gcp.compute.Subnetworks.Get(gcp.project, gcp.region, name).Do()
	if err != nil {
		return nil, err
	}
	return gcp.subnetRepresentation(subnet), nil
}

func (gcp *GCP) ListSubnets(netName string) (*ell.Object, error) {
	/*	net, err := gcp.GetNetwork(netName)
		//note: subnets must be enumerated in each region
		if err != nil {
			return nil, err
		}
		res, err := prov.compute.Subnetworks.List(prov.project, prov.region).Do()
		if err != nil {
			return nil, err
		}
	*/
	nets := make([]*ell.Object, 0)
	/*
		for _, item := range res.Items {
			repr := ell.MakeStruct(2)
			ell.Put(repr, ell.Intern("name:"), ell.String(item.Name))
			ell.Put(repr, ell.Intern("description:"), ell.String(item.Description))
			obj, _ := ell.Instance(NetworkType, repr)
			nets = append(nets, obj)
		}
	*/
	return ell.VectorFromElementsNoCopy(nets), nil
}

/*
	if err != nil {
		return nil, err
	}
	lst := make([]*ell.Object, 0, len(nets))
	for _, net := range nets {
		repr := ell.MakeStruct(2)
		ell.Put(repr, ell.Intern("name:"), ell.String(net.Name()))
		ell.Put(repr, ell.Intern("description:"), ell.String(net.Description()))
		obj := ell.NewObject(NetworkType, repr)
		ell.Println("obj: ", obj, ", type: ", obj.Type)
		lst = append(lst, obj)
	}
	ell.Println("lst:", lst)
	return ell.ListFromValues(lst), nil
*/

func (gcp *GCP) Repr() map[string]interface{} {
	repr := map[string]interface{}{
		"provider": "gcp",
	}
	if gcp.profile != "" {
		repr["profile"] = gcp.profile
	}
	if gcp.identity != "" {
		repr["identity"] = gcp.identity
	}
	if gcp.project != "" {
		repr["project"] = gcp.project
	}
	if gcp.region != "" {
		repr["region"] = gcp.region
	}
	return repr
}

//func (gcp *GCP) String() string {
//	return pretty(gcp.Repr())
//}

/*
func gcpNetworkList(project string) ([]*compute.Network, error) {
	res, err := cloud.compute.Networks.List(cloud.project).Do()
	if err != nil {
		return nil, err
	}
	return res.Items, nil
}
*/

//func (cloud *GcpCloud) gcpListNetworks() (*ell.Object, error) {
//}

/*
func (cloud *GcpCloud) ListSubnets() (*ell.Object, error) {
	res := ell.MakeList()
	n, err := cloud.gcpNetwork()
	if err == nil && n != nil {
		for _, sn := range n.Subnetworks {
			_, name := urlPathSplit(sn)
			if name != "" {
				gcp, err := cloud.gcpSubnetwork(name)
				if err != nil {
					return nil, err
				}
				res = append(res, &Subnet{GCP: gcp})
			}
		}
	}
	return res, err
}
*/
