// +build k8s

/*
 * Copyright (C) 2018 IBM, Inc.
 *
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 *
 */

package tests

import (
	"fmt"
	"os"
	"testing"

	"github.com/skydive-project/skydive/gremlin"
	"github.com/skydive-project/skydive/tests/helper"
)

func k8sConfigFile(name string) string {
	return "./k8s/" + name + ".yaml"
}

const (
	manager    = "k8s"
	objectName = "skydive-test"
)

var (
	networkPolicyConfig = k8sConfigFile("networkpolicy")
	namespaceConfig     = k8sConfigFile("namespace")
)

var (
	nodeName, _       = os.Hostname()
	podName           = objectName
	containerName     = objectName
	networkPolicyName = objectName
	namespaceName     = objectName
)

var (
	setupPod = []helper.Cmd{
		{"kubectl run " + podName +
			"  --image=gcr.io/google_containers/echoserver:1.4" +
			"  --port=8080", true},
	}
	tearDownPod = []helper.Cmd{
		{"kubectl delete deployment " + podName, false},
	}
	setupNetworkPolicy = []helper.Cmd{
		{"kubectl create -f " + networkPolicyConfig, true},
	}
	tearDownNetworkPolicy = []helper.Cmd{
		{"kubectl delete -f " + networkPolicyConfig, false},
	}
	setupNamespace = []helper.Cmd{
		{"kubectl create -f " + namespaceConfig, true},
	}
	tearDownNamespace = []helper.Cmd{
		{"kubectl delete -f " + namespaceConfig, false},
	}
)

func testNodeCreation(t *testing.T, setupCmds, tearDownCmds []helper.Cmd, typ, name *gremlin.ValueString) {
	test := &Test{
		mode:         OneShot,
		setupCmds:    setupCmds,
		tearDownCmds: tearDownCmds,
		checks: []CheckFunction{func(c *CheckContext) error {
			g := gremlin.NewQueryString()
			g.G().V().HasNode(gremlin.NewValueString("k8s").Quote(), typ, name)
			fmt.Printf("Gremlin: %s\n", g.String())

			nodes, err := c.gh.GetNodes(g.String())
			if err != nil {
				return err
			}

			if len(nodes) != 1 {
				return fmt.Errorf("Ran \"%+v\", expected 1 node, got %+v", g, nodes)
			}

			return nil
		}},
	}
	RunTest(t, test)
}

func TestK8sPodNode(t *testing.T) {
	testNodeCreation(t, setupPod, tearDownPod, gremlin.NewValueString("pod").Quote(), gremlin.NewValueString(podName).StartsWith())
}

func TestK8sContainerNode(t *testing.T) {
	testNodeCreation(t, setupPod, tearDownPod, gremlin.NewValueString("container").Quote(), gremlin.NewValueString(containerName).Quote())
}

func TestK8sNetworkPolicyNode(t *testing.T) {
	testNodeCreation(t, setupNetworkPolicy, tearDownNetworkPolicy, gremlin.NewValueString("networkpolicy").Quote(), gremlin.NewValueString(networkPolicyName).Quote())
}

func TestK8sNodeNode(t *testing.T) {
	testNodeCreation(t, nil, nil, gremlin.NewValueString("node").Quote(), gremlin.NewValueString(nodeName).Quote())
}

func TestK8sNamespaceNode(t *testing.T) {
	testNodeCreation(t, setupNamespace, tearDownNamespace, gremlin.NewValueString("namespace").Quote(), gremlin.NewValueString(namespaceName).Quote())
}