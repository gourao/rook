/*
Copyright 2016 The Rook Authors. All rights reserved.

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
package k8s

import (
	"fmt"

	"github.com/rook/rook/pkg/cephmgr/client"
	"github.com/rook/rook/pkg/cephmgr/mon"
	"github.com/rook/rook/pkg/cephmgr/rgw"
	"github.com/rook/rook/pkg/clusterd"
	"github.com/rook/rook/pkg/model"
	"github.com/rook/rook/pkg/operator/k8sutil"
	k8smds "github.com/rook/rook/pkg/operator/mds"
	k8srgw "github.com/rook/rook/pkg/operator/rgw"
	"k8s.io/client-go/kubernetes"
)

type clusterHandler struct {
	clientset   *kubernetes.Clientset
	context     *clusterd.DaemonContext
	clusterInfo *mon.ClusterInfo
	factory     client.ConnectionFactory
	version     string
}

func New(clientset *kubernetes.Clientset, context *clusterd.DaemonContext, clusterInfo *mon.ClusterInfo, factory client.ConnectionFactory, containerVersion string) *clusterHandler {
	return &clusterHandler{clientset: clientset, context: context, clusterInfo: clusterInfo, factory: factory, version: containerVersion}
}

func (s *clusterHandler) GetClusterInfo() (*mon.ClusterInfo, error) {
	return s.clusterInfo, nil
}

func (s *clusterHandler) EnableObjectStore() error {
	logger.Infof("Starting the Object store")
	r := k8srgw.New(k8sutil.Namespace, s.version, s.factory)
	err := r.Start(s.clientset, s.clusterInfo)
	if err != nil {
		return fmt.Errorf("failed to start rgw. %+v", err)
	}
	return nil
}

func (s *clusterHandler) RemoveObjectStore() error {
	logger.Infof("TODO: Remove the object store")
	return nil
}

func (s *clusterHandler) GetObjectStoreConnectionInfo() (*model.ObjectStoreConnectInfo, bool, error) {
	logger.Infof("Getting the object store connection info")
	service, err := s.clientset.Services(k8sutil.Namespace).Get("rgw")
	if err != nil {
		return nil, false, fmt.Errorf("failed to get rgw service. %+v", err)
	}

	info := &model.ObjectStoreConnectInfo{
		Host:       "rook-rgw",
		IPEndpoint: rgw.GetRGWEndpoint(service.Spec.ClusterIP),
	}
	logger.Infof("Object store connection: %+v", info)
	return info, true, nil
}

func (s *clusterHandler) StartFileSystem(fs *model.FilesystemRequest) error {
	logger.Infof("Starting the MDS")
	c := k8smds.New(k8sutil.Namespace, s.version, s.factory)
	return c.Start(s.clientset, s.clusterInfo)
}

func (s *clusterHandler) RemoveFileSystem(fs *model.FilesystemRequest) error {
	logger.Infof("TODO: Remove file system")
	return nil
}

func (s *clusterHandler) GetMonitors() (map[string]*mon.CephMonitorConfig, error) {
	logger.Infof("TODO: Get monitors")
	mons := map[string]*mon.CephMonitorConfig{}

	return mons, nil
}

func (s *clusterHandler) GetNodes() ([]model.Node, error) {
	logger.Infof("Getting nodes")
	return getNodes(s.clientset)
}
