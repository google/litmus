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

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/google/litmus/cli/cmd"
	"github.com/google/litmus/cli/utils"
)

func main() {
	// Get default project ID
	projectID, err := utils.GetDefaultProjectID()
	if err != nil {
		utils.HandleGcloudError(err)
		return
	}

	// Get command and parameters
	if len(os.Args) < 2 {
		utils.PrintUsage()
		return
	}

	command := os.Args[1]
	region := "us-central1" // Default region
	var runID string

	// Parse command-line arguments
	for i := 2; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--project":
			if i+1 < len(os.Args) {
				projectID = os.Args[i+1]
				i++
			} else {
				fmt.Println("Error: --project flag requires an argument")
				return
			}
		case "--region":
			if i+1 < len(os.Args) {
				region = os.Args[i+1]
				i++
			} else {
				fmt.Println("Error: --region flag requires an argument")
				return
			}
		case "open": // Assuming "open" might also need a runID
			if i+1 < len(os.Args) {
				runID = os.Args[i+1]
				i++
			}
			// No error here, as "open" without runID might be valid
		case "run":
			if i+1 < len(os.Args) {
				runID = os.Args[i+1]
				i++
			} else {
				fmt.Println("Error: 'run' command requires a runID argument")
				return
			}
		}
	}

	// Extract environment variables from command-line arguments
	envVars := make(map[string]string)
	for i := 2; i < len(os.Args); i++ { // Start from index 2 to skip command and runID
		parts := strings.Split(os.Args[i], "=")
		if len(parts) == 2 {
			envVars[parts[0]] = parts[1]
		}
	}

	switch command {
	case "deploy":
		cmd.DeployApplication(projectID, region, envVars)
	case "destroy":
		cmd.DestroyResources(projectID, region)
	case "execute":
		if len(os.Args) < 4 {
			fmt.Println("Usage: go run main.go execute <payload>")
			return
		}
		payload := os.Args[3]
		cmd.ExecutePayload(projectID, payload)
	case "ls":
		cmd.ListRuns(projectID)
	case "open":
		if runID != "" {
			cmd.OpenRun(projectID, runID) // Open specific run
		} else {
			cmd.OpenLitmus(projectID) // Open Litmus dashboard
		}
	case "run":
		if runID == "" {
			fmt.Println("Error: 'run' command requires a runID argument")
			return
		}
		cmd.OpenRun(projectID, runID)
	case "status":
		cmd.ShowStatus(projectID)
	case "version":
		utils.DisplayVersion()
	default:
		fmt.Println("Invalid command:", command)
		utils.PrintUsage()
	}
}
