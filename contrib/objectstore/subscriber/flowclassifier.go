/*
 * Copyright (C) 2019 IBM, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy ofthe License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specificlanguage governing permissions and
 * limitations under the License.
 *
 */

package subscriber

import (
	"errors"
	"fmt"
	"net"

	"github.com/skydive-project/skydive/flow"
	"github.com/skydive-project/skydive/logging"
)

// FlowClassifier classifies flows to different tags (strings)
type FlowClassifier interface {
	// GetFlowTag returns the tag of the given flow
	GetFlowTag(fl *flow.Flow) Tag
}

// Tag represents the flow classification
type Tag string

const (
	tagOther    Tag = "other"
	tagEgress   Tag = "egress"
	tagIngress  Tag = "ingress"
	tagInternal Tag = "internal"
)

// FlowDirectionClassifier classifies flows by their direction (ingress, egress, etc)
type FlowDirectionClassifier struct {
	clusterNetMasks []*net.IPNet
}

// GetFlowTag tag flows based on src and dst IP ranges
func (fc *FlowDirectionClassifier) GetFlowTag(fl *flow.Flow) Tag {
	if fl == nil || fl.Network == nil {
		return tagOther
	}
	isSrcInCluster, err := fc.isClusterIP(fl.Network.A)
	if err != nil {
		return tagOther
	}
	isDstInCluster, err := fc.isClusterIP(fl.Network.B)
	if err != nil {
		return tagOther
	}

	if isSrcInCluster {
		if isDstInCluster {
			return tagInternal
		}
		return tagEgress
	}

	if isDstInCluster {
		return tagIngress
	}
	return tagOther
}

// isClusterIP check if IP is in defined subnet
func (fc *FlowDirectionClassifier) isClusterIP(ip string) (bool, error) {
	var err error
	clusterIP := false
	netIP := net.ParseIP(ip)
	if netIP == nil {
		err = errors.New("Cannot parse IP " + ip)
		logging.GetLogger().Warn(err.Error())
		return false, err
	}

	for _, mask := range fc.clusterNetMasks {
		clusterIP = clusterIP || mask.Contains(netIP)
		if clusterIP {
			return true, nil
		}
	}
	return false, nil
}

// NewFlowClassifier returns a new FlowDirectionClassifier, based on the given cluster net masks
func NewFlowClassifier(clusterNetMasks []string) (*FlowDirectionClassifier, error) {
	parsedNetMasks := make([]*net.IPNet, 0, len(clusterNetMasks))
	for _, netMask := range clusterNetMasks {
		_, sa, err := net.ParseCIDR(netMask)
		if err != nil {
			return nil, fmt.Errorf("Cannot parse netmask '%s': %s", netMask, err.Error())
		}
		parsedNetMasks = append(parsedNetMasks, sa)
	}
	return &FlowDirectionClassifier{clusterNetMasks: parsedNetMasks}, nil
}
