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

	"github.com/google/litmus/cli/api"
	"github.com/google/litmus/cli/utils"
)

// ListRuns retrieves and displays a list of Litmus runs.
func ListRuns(projectID string) error {
	serviceURL, err := utils.AccessSecret(projectID, "litmus-service-url")
	if err != nil {
		log.Fatalf("Error retrieving service URL from Secret Manager: %v", err)
	}

	serviceURL = utils.RemoveAnsiEscapeSequences(serviceURL) 

	username, password, err := utils.GetAuthCredentials(projectID)
	if err != nil {
		return fmt.Errorf("error getting authentication credentials: %w", err)
	}

	// Create HTTP client
	client := &http.Client{}
	req, err := http.NewRequest("GET", serviceURL+"/runs", nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	// Set basic auth header
	req.SetBasicAuth(username, password)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Decode the response into a struct that matches the API response
	var response struct {
		Runs []api.RunInfo `json:"runs"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Fatalf("Error decoding response: %v", err)
	}

	runs := response.Runs // Access the runs slice from the decoded response

	if len(runs) == 0 {
		fmt.Println("No runs found.")
	} else {
		fmt.Println("Runs:")
		for _, run := range runs {
			fmt.Printf("Run ID: %s, Status: %s, Progress: %s, StartTime: %s, URL: %s/#/runs/%s\n", run.RunID, run.Status, run.Progress, run.StartTime, serviceURL, run.RunID)
		}
	}
	return nil
}