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

package pullreq

import (
	"net/http"

	"github.com/harness/gitness/app/api/controller/pullreq"
	"github.com/harness/gitness/app/api/render"
	"github.com/harness/gitness/app/api/request"
)

// HandlePRBranchCandidates returns an http.HandlerFunc that processes a request to find
// branches recently updated by the current user that don't have PRs.
func HandlePRBranchCandidates(pullreqCtrl *pullreq.Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		session, _ := request.AuthSessionFrom(ctx)
		repoRef, err := request.GetRepoRefFromPath(r)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		limit := request.ParseLimitOrDefaultWithMax(r, pullreq.PRBannerDefaultLimitForPage,
			pullreq.PRBannerMaxLimitForPage)
		branches, err := pullreqCtrl.PRBranchCandidates(ctx, session, repoRef, limit)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		render.JSON(w, http.StatusOK, branches)
	}
}
