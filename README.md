# ell-cloud
The Ell language interface to AWS and GCP public clouds.

The general idea is to have a simple describe/plan/apply/destroy set up functionality based on a a cloud description data
structure that is the same for both declaring intent and reading current live status.

## Usage

This assumes you have set up gcloud related tooling, with some configurations, as it reuses the config conventions, called
"profile" here.

    $ go get github.com/boynton/ell-cloud/...
    $ cell
    ell v0.2 (with cloud bindings)
    ? (cloud)
    = #[cloud provider: "gcp" profile: "dev" region: "" identity: "somebody@somwhere.com" account: "dev-one"]
    ? (cloud profile: "dev2")
    = #[cloud provider: "gcp" profile: "dev2" region: "us-west1" identity: "somebody@somewhere.com" account: "dev-two"]
    ? (cloud describe: #<cloud>{})
    = (#<network>{id: "2014652406467471095" name: "default" description: "Default network for the project" created: "2016-09-04T14:54:32.364-07:00" subnets: ("default" "default" "default" "default" "default")} #<network>{created: "2016-09-04T15:12:36.915-07:00" subnets: ("mysubnet" "mynet-us-west1-a") id: "188895417056961211" name: "mynet" description: "My test network"})
    ? (cloud describe: #<network>{name: "mynet"})
    = #<network>{description: "My test network" created: "2016-09-04T15:12:36.915-07:00" id: "188895417056961211" name: "mynet" subnets: ("mysubnet" "mynet-us-west1-a")}
    ? (subnets: (cloud describe: #<network>{name: "mynet"}))
    = #<subnet>{name: "mysubnet"}
    ? (cloud describe: (car (subnets: (cloud describe: #<network>{name: "mynet"}))))
    = #<subnet>{cidr: "10.0.1.0/24" id: "6077311531381453533" created: "2016-09-05T12:49:06.336-07:00" gateway: "10.0.1.1" name: "mysubnet"}

Note also that "account" is used as a more generic term for a gcp project.

## License

Copyright 2016 Lee Boynton

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
                                                                                                                                   
  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
