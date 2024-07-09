// Copyright 2023 Harness, Inc.
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

package infraprovider

import (
	"context"
	"fmt"
	"strings"
	"time"

	apiauth "github.com/harness/gitness/app/api/auth"
	"github.com/harness/gitness/app/auth"
	"github.com/harness/gitness/infraprovider"
	infraproviderenum "github.com/harness/gitness/infraprovider/enum"
	"github.com/harness/gitness/types"
	"github.com/harness/gitness/types/check"
	"github.com/harness/gitness/types/enum"
)

type CreateInput struct {
	Identifier string                              `json:"identifier"`
	SpaceRef   string                              `json:"space_ref"` // Ref of the parent space
	Name       string                              `json:"name"`
	Type       infraproviderenum.InfraProviderType `json:"type"`
	Metadata   map[string]string                   `json:"metadata"`
	Resources  []ResourceInput                     `json:"resources"`
}

type ResourceInput struct {
	Identifier         string                              `json:"identifier"`
	Name               string                              `json:"name"`
	InfraProviderType  infraproviderenum.InfraProviderType `json:"infra_provider_type"`
	CPU                *string                             `json:"cpu"`
	Memory             *string                             `json:"memory"`
	Disk               *string                             `json:"disk"`
	Network            *string                             `json:"network"`
	Region             []string                            `json:"region"`
	Metadata           map[string]string                   `json:"metadata"`
	GatewayHost        *string                             `json:"gateway_host"`
	GatewayPort        *string                             `json:"gateway_port"`
	TemplateIdentifier *string                             `json:"template_identifier"`
}

// Create creates a new infraprovider config.
func (c *Controller) Create(
	ctx context.Context,
	session *auth.Session,
	in *CreateInput,
) (*types.InfraProviderConfig, error) {
	if err := c.sanitizeCreateInput(in); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	now := time.Now().UnixMilli()
	parentSpace, err := c.spaceStore.FindByRef(ctx, in.SpaceRef)
	if err != nil {
		return nil, fmt.Errorf("failed to find parent by ref: %w", err)
	}
	if err = apiauth.CheckInfraProvider(
		ctx,
		c.authorizer,
		session,
		parentSpace.Path,
		"",
		enum.PermissionInfraProviderEdit); err != nil {
		return nil, err
	}
	infraProviderConfig := &types.InfraProviderConfig{
		Identifier: in.Identifier,
		Name:       in.Name,
		SpaceID:    parentSpace.ID,
		Type:       in.Type,
		Created:    now,
		Updated:    now,
	}
	err = c.infraProviderConfigStore.Create(ctx, infraProviderConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create infraprovider config for : %q %w", infraProviderConfig.Identifier, err)
	}
	infraProviderConfiginDB, err := c.infraProviderConfigStore.FindByIdentifier(
		ctx,
		parentSpace.ID,
		infraProviderConfig.Identifier)
	if err != nil {
		return nil, err
	}
	infraProvider, err := c.infraProviderFactory.GetInfraProvider(infraProviderConfiginDB.Type)
	if err != nil {
		return nil, err
	}
	if len(infraProvider.TemplateParams()) > 0 {
		return nil, fmt.Errorf("failed to fetch templates") // TODO Implement
	}
	parameters := []infraprovider.Parameter{}
	// TODO logic to populate paramteters as per the provider type
	err = infraProvider.ValidateParams(parameters)
	if err != nil {
		return nil, err
	}
	for _, res := range in.Resources {
		entity := &types.InfraProviderResource{
			Identifier:            res.Identifier,
			InfraProviderConfigID: infraProviderConfiginDB.ID,
			InfraProviderType:     res.InfraProviderType,
			Name:                  res.Name,
			SpaceID:               parentSpace.ID,
			CPU:                   res.CPU,
			Memory:                res.Memory,
			Disk:                  res.Disk,
			Network:               res.Network,
			Region:                strings.Join(res.Region, " "), // TODO fix
			Metadata:              res.Metadata,
			GatewayHost:           res.GatewayHost,
			GatewayPort:           res.GatewayPort, // No template as of now
			Created:               now,
			Updated:               now,
		}
		err = c.infraProviderResourceStore.Create(ctx, infraProviderConfiginDB.ID, entity)
		if err != nil {
			return nil, fmt.Errorf("failed to create infraprovider resource for : %q %w", entity.Identifier, err)
		}
	}
	resources, err := c.infraProviderResourceStore.List(ctx, infraProviderConfiginDB.ID, types.ListQueryFilter{})
	infraProviderConfig.Resources = resources
	if err != nil {
		return nil, fmt.Errorf(
			"error creating infra provider resource for config : %q %w", infraProviderConfiginDB.Identifier, err)
	}
	return infraProviderConfig, nil
}

func (c *Controller) sanitizeCreateInput(in *CreateInput) error {
	if err := check.Identifier(in.Identifier); err != nil {
		return err
	}
	for _, resource := range in.Resources {
		if err := check.Identifier(resource.Identifier); err != nil {
			return err
		}
	}
	return nil
}
