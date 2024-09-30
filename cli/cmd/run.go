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
	"io"
	"log"
	"net/http"

	"github.com/google/litmus/cli/api"
	"github.com/google/litmus/cli/utils"
)

// OpenRun opens the URL associated with a specific Litmus run ID in the browser.
func OpenRun(projectID, runID string) error {
	serviceURL, err := utils.AccessSecret(projectID, "litmus-service-url")
	if err != nil {
		log.Fatalf("Error retrieving service URL from Secret Manager: %v", err)
	}
	serviceURL = utils.RemoveAnsiEscapeSequences(serviceURL)

	username, password, err := utils.GetAuthCredentials(projectID)
	if err != nil {
		return fmt.Errorf("error getting authentication credentials: %w", err)
	}

	runURL := fmt.Sprintf("%s/runs/status/%s", serviceURL, runID)
	fmt.Println(runURL)

	// Create HTTP client
	client := &http.Client{}
	req, err := http.NewRequest("GET", runURL, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	// Set basic auth header
	req.SetBasicAuth(username, password)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close() // Close the body AFTER reading

	// Handle the response (check for success/errors)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %s, response: %s", resp.Status, string(body))
	}

	body, err := io.ReadAll(resp.Body) // Read the body here
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	// Unmarshal the JSON response
	var runDetails api.RunDetails
	if err := json.Unmarshal(body, &runDetails); err != nil {
		return fmt.Errorf("error unmarshalling JSON response: %w", err)
	}

	// Now you can access the data in a structured way:
	fmt.Println("Progress:", runDetails.Progress)
	fmt.Println("Status:", runDetails.Status)
	// ... access other fields ...

	for _, testCase := range runDetails.TestCases {
		fmt.Println("Test Case ID:", testCase.ID)
		fmt.Println("Status:", testCase.Response.Status)
		fmt.Println("Tracing ID:", testCase.TracingID)
		// ... access other fields within the test case ...
	}

	return nil

}