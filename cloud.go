/*
Copyright 2014 Lee Boynton

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cloud

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/boynton/ell"
)

//
// A Cloud provider supports this abstraction. Such a "cloud" is in a single region.
// problem: the "default" network in GCP is cross-region. How to represent that?
// -> only show the subnets in the selected region
//
type Provider interface {
	Name() string
	Profile() string
	Identity() string
	Account() string
	Region() string

	Describe(nameOrTemplate *ell.Object) (*ell.Object, error)
	Plan(template *ell.Object) (*ell.Object, error)
	Apply(template *ell.Object) (*ell.Object, error)
	Destroy(nameOrtemplate *ell.Object) (*ell.Object, error)

	ListNetworks() (*ell.Object, error)
	CreateNetwork(name string, cidr string, zones []string) (*ell.Object, error)
	DestroyNetwork(name string) (*ell.Object, error)
	DescribeNetwork(name string) (*ell.Object, error)
	DescribeSubnet(name string) (*ell.Object, error)

//	ListSubnets(net string) (*ell.Object, error)
}

//-------------------------------------------------------------------------

var CloudErrorKey = ell.Intern("cloud-error:")

//types that are all templates
var CloudType = ell.Intern("<cloud>")
var NetworkType = ell.Intern("<network>")
var SubnetType = ell.Intern("<subnet>")

func ellCloud(argv []*ell.Object) (*ell.Object, error) {
	provider := ell.StringValue(argv[0])
	profile := ell.StringValue(argv[1])
	region := ell.StringValue(argv[2])
	describe := argv[3]
	plan := argv[4]
	apply := argv[5]
	destroy := argv[6]
	var prov Provider
	var err error
	switch provider {
	case "gcp":
		prov, err = gcpProvider(profile, region)
	case "aws":
		prov, err = awsProvider(profile, region)
	default:
		err = ell.Error(CloudErrorKey, "Unrecognized cloud provider: '", provider, "'")
	}
	if err != nil {
		return nil, ell.Error(ell.ArgumentErrorKey, fmt.Sprintf("cannot connect to provider '%s': %v  ", provider, err))
	}
	if describe != ell.Null {
		return prov.Describe(describe)
	} else if plan != ell.Null {
		return prov.Plan(plan)
	} else if apply != ell.Null {
		return prov.Apply(apply)
	} else if destroy != ell.Null {
		return prov.Destroy(destroy)
	}
	return ell.NewObject(CloudType, prov), nil //a <cloud> type can be either connect, like this, or as a simple template
}

func pretty(obj interface{}) string {
	b, _ := json.MarshalIndent(obj, "", "   ")
	return string(b)
}

//-------------------------------------------------------------------------

type Extension struct {
}

//(cloud deploy foobar.template
//   provider: "aws",
//   profile: "default"
//   region: "us-west1"
//   template: {}
//   action: 

func (*Extension) Init() error {
	cloudPath := os.Getenv("GOPATH") + "/src/github.com/boynton/ell-cloud"
	ell.AddEllDirectory(cloudPath)
	//ell.DefineFunctionKeyArgs("cloud", ellCloud, ell.StructType, //(cloud provider: "aws" profile: "dev")
	ell.DefineFunctionKeyArgs("cloud", ellCloud, CloudType, //(cloud provider: "aws" profile: "dev")
		[]*ell.Object{ //return types
			ell.StringType,
			ell.StringType,
			ell.StringType,
			ell.AnyType, //string or template type
			ell.AnyType, //any template type
			ell.AnyType, //any template type
			ell.AnyType, //string ot template type
		},
		[]*ell.Object{ //defaults
			ell.String("gcp"),
			ell.String(""),
			ell.String(""),
			ell.Null,
			ell.Null,
			ell.Null,
			ell.Null,
		},
		[]*ell.Object{ //keywords
			ell.Intern("provider:"),
			ell.Intern("profile:"),
			ell.Intern("region:"),
			ell.Intern("describe:"),
			ell.Intern("plan:"),
			ell.Intern("apply:"),
			ell.Intern("destroy:"),
		},
	)
//	ell.DefineFunction("networks", ellNetworks, ell.ListType, CloudType)
//	ell.DefineFunction("network", ellNetwork, ell.AnyType, CloudType, ell.StringType)
//	ell.DefineFunction("subnet", ellSubnet, ell.AnyType, CloudType, ell.StringType)
//	ell.DefineFunction("subnet", ellSubnet, ell.AnyType, CloudType, ell.StringType)
	return ell.Load("cloud")
}

func (*Extension) Cleanup() {
}

func (e *Extension) String() string {
	return "cloud"
}
