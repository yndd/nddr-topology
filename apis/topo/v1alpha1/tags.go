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
	MaxUint32 = ^uint32(0)
	MinUint32 = 0
	MaxInt    = int(MaxUint32 >> 1)
	MinInt    = -MaxInt - 1
)

const (
	KeyPurpose            = "purpose"  // used in ipam for loopback, isl
	KeyNode               = "node"     // used in allocation e.g. aspool
	KeyNodePlatform       = "platform" // ixdd2, sr1, sr1s
	KeyNodePosition       = "position"
	KeyNodeIndex          = "index"      // index for determinsitic allocations
	KeyLink               = "link"       // used in allocation
	KeyLinkLagMember      = "lag-member" // true or false (default) -> this is set in the config
	KeyLinkLag            = "lag"        // true or false (default) -> this is set on the logical link which is created by the topolink parser
	KeyLinkLacp           = "lacp"       // true (default) or false
	KeyLinkKind           = "kind"       // "infra" (default), "loop" -> used when both sides of the link are the same
	KeyLinkEPKind         = "kind"       // "infra" (default), "loop", "access", "oob"
	KeyLinkEPLacpFallback = "lacp-fallback"  // true or false (default)
	//keyLinkEPSRIOV   = "sriov"   // "true", "false" (default)
	//keyLinkEPIPVLAN   = "ipvlan"   // "true", "false" (default)
	KeyLinkEPGroup           = "endpoint-group"   //  server-pod1, dcgw1 -> default("")
	KeyLinkEPLagName         = "lag-name"         // flexible string
	KeyLinkEPMultiHoming     = "multihoming"      // true or false (default)
	KeyLinkEPMultiHomingName = "multihoming-name" // flexible string (group)
	//keyLinkEPBreakout        = "breakout"            // -> to be discussed (true, false;) -> with a real interface CR (single ended)
)

// p2p lag
// set lag-member to true on individual links
// -> the link reconciler creates a new logical link with name : <prefix:logical-link>-<node-name-epA>-<lag-name-epA>-<node-name-epB><lag-name-epB>)

// mh lag
// set lag-member to true on individual links
// set multihoming to true on individual links
// set multihoming-name to a global unique name
// -> the link reconciler creates a new logical link with name : <prefix:logical-mh-link>-<multihoming-name>-<node-name-epB><lag-name-epB>)

type LinkEPKind string

const (
	LinkEPKindInfra   LinkEPKind = "infra"
	LinkEPKindLoop    LinkEPKind = "loop"
	LinkEPKindAccess  LinkEPKind = "access"
	LinkEPKindOob     LinkEPKind = "oob"
	LinkEPKindUnknown LinkEPKind = "unknown"
)

func (s LinkEPKind) String() string {
	switch s {
	case LinkEPKindInfra:
		return "infra"
	case LinkEPKindLoop:
		return "loop"
	case LinkEPKindAccess:
		return "access"
	case LinkEPKindOob:
		return "oob"
	}
	return "unknown"
}
