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

package ide

import (
	"context"
	"strings"

	"github.com/harness/gitness/app/gitspace/orchestrator/devcontainer"
	gitspaceTypes "github.com/harness/gitness/app/gitspace/types"
	"github.com/harness/gitness/types"
	"github.com/harness/gitness/types/enum"
)

const (
	templateSetupSSHServer string = "setup_ssh_server.sh"
	templateRunSSHServer   string = "run_ssh_server.sh"
)

type IDE interface {
	// Setup is responsible for doing all the operations for setting up the IDE in the container e.g. installation,
	// copying settings and configurations.
	Setup(
		ctx context.Context,
		exec *devcontainer.Exec,
		args map[gitspaceTypes.IDEArg]interface{},
		gitspaceLogger gitspaceTypes.GitspaceLogger,
	) error

	// Run runs the IDE and supporting services.
	Run(
		ctx context.Context,
		exec *devcontainer.Exec,
		args map[gitspaceTypes.IDEArg]interface{},
		gitspaceLogger gitspaceTypes.GitspaceLogger,
	) error

	// Port provides the port which will be used by this IDE.
	Port() *types.GitspacePort

	// Type provides the IDE type to which the service belongs.
	Type() enum.IDEType

	// GenerateURL returns the url to redirect user to ide from gitspace
	GenerateURL(absoluteRepoPath, host, port, user string) string
}

func getHomePath(absoluteRepoPath string) string {
	pathList := strings.Split(absoluteRepoPath, "/")
	return strings.Join(pathList[:len(pathList)-1], "/")
}
