// Copyright 2024 Google LLC
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

package cmd

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/google/litmus/cli/utils"
)

// OpenRun opens the URL associated with a specific Litmus run ID in the browser.
func OpenRun(projectID, runID string) {
	serviceURL, err := utils.AccessSecret(projectID, "litmus-service-url")
	if err != nil {
		log.Fatalf("Error retrieving service URL from Secret Manager: %v", err)
	}

	runURL := fmt.Sprintf("%s/runs/%s", serviceURL, runID)

	if err := exec.Command("open", runURL).Start(); err != nil {
		log.Fatalf("Error opening URL: %v", err)
	}
}