//  Copyright 2023 Harness, Inc.
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

package rpm

import (
	urlprovider "github.com/harness/gitness/app/url"
	"github.com/harness/gitness/registry/app/pkg/base"
	"github.com/harness/gitness/registry/app/pkg/filemanager"
	"github.com/harness/gitness/registry/app/store"
	"github.com/harness/gitness/registry/services/index"
	"github.com/harness/gitness/store/database/dbtx"

	"github.com/google/wire"
)

func LocalRegistryProvider(
	localBase base.LocalBase,
	fileManager filemanager.FileManager,
	proxyStore store.UpstreamProxyConfigRepository,
	tx dbtx.Transactor,
	registryDao store.RegistryRepository,
	imageDao store.ImageRepository,
	artifactDao store.ArtifactRepository,
	urlProvider urlprovider.Provider,
	registryIndexService index.Service,
) LocalRegistry {
	registry := NewLocalRegistry(localBase, fileManager, proxyStore, tx, registryDao, imageDao, artifactDao,
		urlProvider, registryIndexService)
	base.Register(registry)
	return registry
}

func ProxyProvider(
	proxyStore store.UpstreamProxyConfigRepository,
	registryDao store.RegistryRepository,
	imageDao store.ImageRepository,
	artifactDao store.ArtifactRepository,
	fileManager filemanager.FileManager,
	tx dbtx.Transactor,
	urlProvider urlprovider.Provider,
) Proxy {
	proxy := NewProxy(fileManager, proxyStore, tx, registryDao, imageDao, artifactDao, urlProvider)
	base.Register(proxy)
	return proxy
}

var WireSet = wire.NewSet(LocalRegistryProvider, ProxyProvider)
