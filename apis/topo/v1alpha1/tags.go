/*
Copyright 2021 NDD.

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

package v1alpha1

const (
	NodePlatform = "platform" // ixdd2, sr1, sr1s
	NodePosition = "position"
	NodeIndex    = "index" // index for determinsitic allocations
	LinkLag      = "lag"   // true or false (default)
	LinkEPKind   = "kind"  // network(default), access(server, security, dcgw), loop
	//LinkEPGroup           = "endpoint-group"      // pod1, dcgw1
	LinkEPLagName = "lag-name" // flexible string
	//LinkEPMultiHoming     = "multihoming"         // true or false (default)
	//LinkEPMultiHomingName = "multihoming-name"    // flexible string
	//LinkEPBreakout        = "breakout"            // -> to be discussed
)
