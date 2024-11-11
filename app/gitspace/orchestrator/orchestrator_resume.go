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

package orchestrator

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/harness/gitness/app/gitspace/orchestrator/container"
	"github.com/harness/gitness/app/gitspace/secret"
	secretenum "github.com/harness/gitness/app/gitspace/secret/enum"
	"github.com/harness/gitness/app/paths"
	"github.com/harness/gitness/types"
	"github.com/harness/gitness/types/enum"

	"github.com/rs/zerolog/log"
)

func (o orchestrator) ResumeStartGitspace(
	ctx context.Context,
	gitspaceConfig types.GitspaceConfig,
	provisionedInfra types.Infrastructure,
) (types.GitspaceInstance, error) {
	gitspaceInstance := gitspaceConfig.GitspaceInstance
	gitspaceInstance.State = enum.GitspaceInstanceStateError

	secretResolver, err := o.getSecretResolver(gitspaceInstance.AccessType)
	if err != nil {
		log.Err(err).Msgf("could not find secret resolver for type: %s", gitspaceInstance.AccessType)
		return *gitspaceInstance, err
	}
	rootSpaceID, _, err := paths.DisectRoot(gitspaceConfig.SpacePath)
	if err != nil {
		log.Err(err).Msgf("unable to find root space id from space path: %s", gitspaceConfig.SpacePath)
		return *gitspaceInstance, err
	}
	resolvedSecret, err := secretResolver.Resolve(ctx, secret.ResolutionContext{
		UserIdentifier:     gitspaceConfig.GitspaceUser.Identifier,
		GitspaceIdentifier: gitspaceConfig.Identifier,
		SecretRef:          *gitspaceInstance.AccessKeyRef,
		SpaceIdentifier:    rootSpaceID,
	})
	if err != nil {
		log.Err(err).Msgf("could not resolve secret type: %s, ref: %s",
			gitspaceInstance.AccessType, *gitspaceInstance.AccessKeyRef)
		return *gitspaceInstance, err
	}
	gitspaceInstance.AccessKey = &resolvedSecret.SecretValue

	ideSvc, err := o.getIDEService(gitspaceConfig)
	if err != nil {
		return *gitspaceInstance, err
	}

	idePort := ideSvc.Port()

	err = o.infraProvisioner.ResumeProvision(ctx, gitspaceConfig, provisionedInfra)
	if err != nil {
		o.emitGitspaceEvent(ctx, gitspaceConfig, enum.GitspaceEventTypeInfraProvisioningFailed)

		return *gitspaceInstance, fmt.Errorf(
			"cannot provision infrastructure for ID %s: %w", gitspaceConfig.InfraProviderResource.UID, err)
	}

	if provisionedInfra.Status != enum.InfraStatusProvisioned {
		o.emitGitspaceEvent(ctx, gitspaceConfig, enum.GitspaceEventTypeInfraProvisioningFailed)

		return *gitspaceInstance, fmt.Errorf(
			"infra state is %v, should be %v for gitspace instance identifier %s",
			provisionedInfra.Status,
			enum.InfraStatusProvisioned,
			gitspaceConfig.GitspaceInstance.Identifier,
		)
	}

	o.emitGitspaceEvent(ctx, gitspaceConfig, enum.GitspaceEventTypeInfraProvisioningCompleted)

	scmResolvedDetails, err := o.scm.GetSCMRepoDetails(ctx, gitspaceConfig)
	if err != nil {
		return *gitspaceInstance, fmt.Errorf(
			"failed to fetch code repo details for gitspace config ID %w %d", err, gitspaceConfig.ID)
	}
	o.emitGitspaceEvent(ctx, gitspaceConfig, enum.GitspaceEventTypeAgentConnectStart)

	err = o.containerOrchestrator.Status(ctx, provisionedInfra)
	if err != nil {
		o.emitGitspaceEvent(ctx, gitspaceConfig, enum.GitspaceEventTypeAgentConnectFailed)

		return *gitspaceInstance, fmt.Errorf("couldn't call the agent health API: %w", err)
	}

	o.emitGitspaceEvent(ctx, gitspaceConfig, enum.GitspaceEventTypeAgentConnectCompleted)

	o.emitGitspaceEvent(ctx, gitspaceConfig, enum.GitspaceEventTypeAgentGitspaceCreationStart)

	// NOTE: Currently we use a static identifier as the Gitspace user.
	gitspaceConfig.GitspaceUser.Identifier = harnessUser

	startResponse, err := o.containerOrchestrator.CreateAndStartGitspace(
		ctx, gitspaceConfig, provisionedInfra, *scmResolvedDetails, o.config.DefaultBaseImage, ideSvc)
	if err != nil {
		o.emitGitspaceEvent(ctx, gitspaceConfig, enum.GitspaceEventTypeAgentGitspaceCreationFailed)

		return *gitspaceInstance, fmt.Errorf("couldn't call the agent start API: %w", err)
	}

	o.emitGitspaceEvent(ctx, gitspaceConfig, enum.GitspaceEventTypeAgentGitspaceCreationCompleted)

	ideURLString := generateIDEURL(provisionedInfra, idePort, startResponse, gitspaceConfig)
	gitspaceInstance.URL = &ideURLString

	now := time.Now().UnixMilli()
	gitspaceInstance.LastUsed = &now
	gitspaceInstance.ActiveTimeStarted = &now
	gitspaceInstance.State = enum.GitspaceInstanceStateRunning

	o.emitGitspaceEvent(ctx, gitspaceConfig, enum.GitspaceEventTypeGitspaceActionStartCompleted)

	return *gitspaceInstance, nil
}

func generateIDEURL(
	provisionedInfra types.Infrastructure,
	idePort *types.GitspacePort,
	startResponse *container.StartResponse,
	gitspaceConfig types.GitspaceConfig,
) string {
	var ideURL url.URL

	var forwardedPort string

	if provisionedInfra.GitspacePortMappings[idePort.Port].PublishedPort == 0 {
		forwardedPort = startResponse.PublishedPorts[idePort.Port]
	} else {
		forwardedPort = strconv.Itoa(provisionedInfra.GitspacePortMappings[idePort.Port].ForwardedPort)
	}

	scheme := provisionedInfra.GitspaceScheme
	host := provisionedInfra.GitspaceHost
	if provisionedInfra.ProxyGitspaceHost != "" {
		host = provisionedInfra.ProxyGitspaceHost
	}

	relativeRepoPath := strings.TrimPrefix(startResponse.AbsoluteRepoPath, "/")

	if gitspaceConfig.IDE == enum.IDETypeVSCodeWeb {
		ideURL = url.URL{
			Scheme:   scheme,
			Host:     host + ":" + forwardedPort,
			RawQuery: filepath.Join("folder=", relativeRepoPath),
		}
	} else if gitspaceConfig.IDE == enum.IDETypeVSCode {
		// TODO: the following userID is hard coded and should be changed.
		ideURL = url.URL{
			Scheme: "vscode-remote",
			Host:   "", // Empty since we include the host and port in the path
			Path: fmt.Sprintf(
				"ssh-remote+%s@%s:%s",
				gitspaceConfig.GitspaceUser.Identifier,
				host,
				filepath.Join(forwardedPort, relativeRepoPath),
			),
		}
	}
	ideURLString := ideURL.String()
	return ideURLString
}

func (o orchestrator) getSecretResolver(accessType enum.GitspaceAccessType) (secret.Resolver, error) {
	secretType := secretenum.PasswordSecretType
	switch accessType {
	case enum.GitspaceAccessTypeUserCredentials:
		secretType = secretenum.PasswordSecretType
	case enum.GitspaceAccessTypeJWTToken:
		secretType = secretenum.JWTSecretType
	case enum.GitspaceAccessTypeSSHKey:
		secretType = secretenum.SSHSecretType
	}
	return o.secretResolverFactory.GetSecretResolver(secretType)
}

func (o orchestrator) ResumeStopGitspace(
	ctx context.Context,
	gitspaceConfig types.GitspaceConfig,
	stoppedInfra types.Infrastructure,
) (enum.GitspaceInstanceStateType, error) {
	instanceState := enum.GitspaceInstanceStateError

	err := o.infraProvisioner.ResumeStop(ctx, gitspaceConfig, stoppedInfra)
	if err != nil {
		o.emitGitspaceEvent(ctx, gitspaceConfig, enum.GitspaceEventTypeInfraStopFailed)

		return instanceState, fmt.Errorf(
			"cannot stop provisioned infrastructure with ID %s: %w", gitspaceConfig.InfraProviderResource.UID, err)
	}

	if stoppedInfra.Status != enum.InfraStatusDestroyed &&
		stoppedInfra.Status != enum.InfraStatusStopped {
		o.emitGitspaceEvent(ctx, gitspaceConfig, enum.GitspaceEventTypeInfraStopFailed)

		return instanceState, fmt.Errorf(
			"infra state is %v, should be %v for gitspace instance identifier %s",
			stoppedInfra.Status,
			enum.InfraStatusDestroyed,
			gitspaceConfig.GitspaceInstance.Identifier)
	}

	o.emitGitspaceEvent(ctx, gitspaceConfig, enum.GitspaceEventTypeInfraStopCompleted)

	instanceState = enum.GitspaceInstanceStateDeleted

	o.emitGitspaceEvent(ctx, gitspaceConfig, enum.GitspaceEventTypeGitspaceActionStopCompleted)

	return instanceState, nil
}

func (o orchestrator) ResumeDeleteGitspace(
	ctx context.Context,
	gitspaceConfig types.GitspaceConfig,
	deprovisionedInfra types.Infrastructure,
) (enum.GitspaceInstanceStateType, error) {
	instanceState := enum.GitspaceInstanceStateError

	err := o.infraProvisioner.ResumeDeprovision(ctx, gitspaceConfig, deprovisionedInfra)
	if err != nil {
		o.emitGitspaceEvent(ctx, gitspaceConfig, enum.GitspaceEventTypeInfraDeprovisioningFailed)
		return instanceState, fmt.Errorf(
			"cannot deprovision infrastructure with ID %s: %w", gitspaceConfig.InfraProviderResource.UID, err)
	}

	if deprovisionedInfra.Status != enum.InfraStatusDestroyed {
		o.emitGitspaceEvent(ctx, gitspaceConfig, enum.GitspaceEventTypeInfraDeprovisioningFailed)

		return instanceState, fmt.Errorf(
			"infra state is %v, should be %v for gitspace instance identifier %s",
			deprovisionedInfra.Status,
			enum.InfraStatusDestroyed,
			gitspaceConfig.GitspaceInstance.Identifier)
	}

	o.emitGitspaceEvent(ctx, gitspaceConfig, enum.GitspaceEventTypeInfraDeprovisioningCompleted)

	instanceState = enum.GitspaceInstanceStateDeleted

	o.emitGitspaceEvent(ctx, gitspaceConfig, enum.GitspaceEventTypeGitspaceActionStopCompleted)
	return instanceState, nil
}

func (o orchestrator) ResumeCleanupInstanceResources(
	ctx context.Context,
	gitspaceConfig types.GitspaceConfig,
	cleanedUpInfra types.Infrastructure,
) (enum.GitspaceInstanceStateType, error) {
	instanceState := enum.GitspaceInstanceStateError

	err := o.infraProvisioner.ResumeCleanupInstance(ctx, gitspaceConfig, cleanedUpInfra)
	if err != nil {
		o.emitGitspaceEvent(ctx, gitspaceConfig, enum.GitspaceEventTypeInfraCleanupFailed)

		return instanceState, fmt.Errorf(
			"cannot cleanup provisioned infrastructure with ID %s: %w",
			gitspaceConfig.InfraProviderResource.UID,
			err,
		)
	}

	if cleanedUpInfra.Status != enum.InfraStatusDestroyed && cleanedUpInfra.Status != enum.InfraStatusStopped {
		o.emitGitspaceEvent(ctx, gitspaceConfig, enum.GitspaceEventTypeInfraCleanupFailed)

		return instanceState, fmt.Errorf(
			"infra state is %v, should be %v for gitspace instance identifier %s",
			cleanedUpInfra.Status,
			[]enum.InfraStatus{enum.InfraStatusDestroyed, enum.InfraStatusStopped},
			gitspaceConfig.GitspaceInstance.Identifier)
	}

	o.emitGitspaceEvent(ctx, gitspaceConfig, enum.GitspaceEventTypeInfraCleanupCompleted)

	instanceState = enum.GitspaceInstanceStateCleaned

	return instanceState, nil
}
