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
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/litmus/cli/utils"
)

// ListRuns retrieves and displays a list of Litmus runs.
func ListRuns(projectID string) {
	serviceURL, err := utils.AccessSecret(projectID, "litmus-service-url")
	if err != nil {
		log.Fatalf("Error retrieving service URL from Secret Manager: %v", err)
	}

	resp, err := http.Get(serviceURL + "/runs")
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	var runs []utils.RunInfo
	if err := json.NewDecoder(resp.Body).Decode(&runs); err != nil {
		log.Fatalf("Error decoding response: %v", err)
	}

	if len(runs) == 0 {
		fmt.Println("No runs found.")
	} else {
		fmt.Println("Runs:")
		for _, run := range runs {
			fmt.Printf("Run ID: %s, URL: %s\n", run.RunID, run.URL)
		}
	}
}