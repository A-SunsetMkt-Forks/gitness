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

package types

import (
	"github.com/harness/gitness/app/api/controller/aiagent/types"
)

type PipelineStageGenerateRequest struct {
	Prompt       string
	Metadata     map[string]string    `json:"metadata"`
	Conversation []types.Conversation `json:"conversation"`
}

type PipelineStageGenerateResponse struct {
	YAML  string
	Error string
}

type PipelineStageUpdateRequest struct {
	Prompt       string
	Stage        string
	Metadata     map[string]string    `json:"metadata"`
	Conversation []types.Conversation `json:"conversation"`
}

type PipelineStageUpdateResponse struct {
	YAML string
}
