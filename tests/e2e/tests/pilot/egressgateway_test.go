// Copyright 2017 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pilot

import (
	"fmt"
	"testing"

	"istio.io/istio/pkg/log"
)

// To route all external traffic via Istio Egress gateway
// 1. Add external services
// 2. Add egress gateway
// 3. Add virtual service for each external service such that
//    3.a. Traffic from all sidecars (i.e. mesh gateway) goes to egress gateway svc
//    3.b. Traffic from egress gateway goes to actual destination (in our case, its t)
// The tests will only check for requests from a->t with host matching ext service
func TestEgressGateway(t *testing.T) {
	if !tc.V1alpha3 {
		t.Skipf("Skipping %s: v1alpha3=false", t.Name())
	}

	cfgs := &deployableConfig{
		Namespace: tc.Kube.Namespace,
		YamlFiles: []string{
			"testdata/v1alpha3/egressgateway.yaml",
			"testdata/v1alpha3/external-service.yaml",
			"testdata/v1alpha3/rule-route-via-egressgateway.yaml"},
	}
	if err := cfgs.Setup(); err != nil {
		t.Fatal(err)
	}
	defer cfgs.Teardown()

	runRetriableTest(t, "RouteViaEgressGateway", defaultRetryBudget, func() error {
		// We use an arbitrary IP to ensure that the test fails if networking logic is implemented incorrectly
		reqURL := fmt.Sprintf("http://1.1.1.1/bookinfo")
		resp := ClientRequest("a", reqURL, 100, "-key Host -val eu.bookinfo.com")
		count := make(map[string]int)
		for _, elt := range resp.Host {
			count[elt] = count[elt] + 1
		}
		log.Infof("request counts %v", count)
		if count["eu.bookinfo.com"] >= 95 {
			return nil
		}
		return errAgain
	})
}
