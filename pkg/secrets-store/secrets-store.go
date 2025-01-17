/*
Copyright 2018 The Kubernetes Authors.

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

package secretsstore

import (
	"github.com/container-storage-interface/spec/lib/go/csi"

	csicommon "github.com/deislabs/secrets-store-csi-driver/pkg/csi-common"

	log "github.com/sirupsen/logrus"
)

type SecretsStore struct {
	driver *csicommon.CSIDriver
	ns     *nodeServer
	cs     *controllerServer
}

var (
	vendorVersion = "0.0.3"
)

func GetDriver() *SecretsStore {
	return &SecretsStore{}
}

func newNodeServer(d *csicommon.CSIDriver, providerVolumePath string) *nodeServer {
	return &nodeServer{
		DefaultNodeServer:  csicommon.NewDefaultNodeServer(d),
		providerVolumePath: providerVolumePath,
	}
}

func newControllerServer(d *csicommon.CSIDriver) *controllerServer {
	return &controllerServer{
		DefaultControllerServer: csicommon.NewDefaultControllerServer(d),
		vols:                    make(map[string]csi.Volume),
	}
}

func (s *SecretsStore) Run(driverName, nodeID, endpoint, providerVolumePath string) {
	log.Infof("Driver: %v ", driverName)
	log.Infof("Version: %s", vendorVersion)
	log.Infof("Provider Volume Path: %s", providerVolumePath)

	// Initialize default library driver
	s.driver = csicommon.NewCSIDriver(driverName, vendorVersion, nodeID)
	if s.driver == nil {
		log.Fatal("Failed to initialize SecretsStore CSI Driver.")
	}
	s.driver.AddControllerServiceCapabilities(
		[]csi.ControllerServiceCapability_RPC_Type{
			csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
		})
	s.driver.AddVolumeCapabilityAccessModes([]csi.VolumeCapability_AccessMode_Mode{
		csi.VolumeCapability_AccessMode_SINGLE_NODE_READER_ONLY,
		csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY,
	})

	s.ns = newNodeServer(s.driver, providerVolumePath)
	s.cs = newControllerServer(s.driver)

	server := csicommon.NewNonBlockingGRPCServer()
	server.Start(endpoint, csicommon.NewDefaultIdentityServer(s.driver), s.cs, s.ns)
	server.Wait()
}
