// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Load Balancing API
//
// API for the Load Balancing service. Use this API to manage load balancers, backend sets, and related items. For more
// information, see Overview of Load Balancing (https://docs.us-phoenix-1.oraclecloud.com/iaas/Content/Balance/Concepts/balanceoverview.htm).
//

package loadbalancer

import (
	"github.com/oracle/oci-go-sdk/common"
)

// Hostname A hostname resource associated with a load balancer for use by one or more listeners.
// **Warning:** Oracle recommends that you avoid using any confidential information when you supply string values using the API.
type Hostname struct {

	// A friendly name for the hostname resource. It must be unique and it cannot be changed. Avoid entering confidential
	// information.
	// Example: `example_hostname_001`
	Name *string `mandatory:"true" json:"name"`

	// A virtual hostname. For more information about virtual hostname string construction, see
	// Managing Request Routing (https://docs.us-phoenix-1.oraclecloud.com/Content/Balance/Tasks/managingrequest.htm#routing).
	// Example: `app.example.com`
	Hostname *string `mandatory:"true" json:"hostname"`
}

func (m Hostname) String() string {
	return common.PointerString(m)
}
