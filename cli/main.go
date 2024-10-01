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

	"github.com/google/litmus/cli/analytics"
	"github.com/google/litmus/cli/cmd"
	"github.com/google/litmus/cli/utils"
	"github.com/google/uuid"
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
	quiet := false           // Check for --quiet flag
	preserveData := false // Flag to preserve data

	// Parse command-line arguments
	args := os.Args[2:] // Skip program name and command
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--project":
			if i+1 < len(args) {
				projectID = args[i+1]
				i++ // Skip the next argument (project ID)
			} else {
				fmt.Println("Error: --project flag requires an argument")
				return
			}
		case "--region":
			if i+1 < len(args) {
				region = args[i+1]
				i++ // Skip the next argument (region)
			} else {
				fmt.Println("Error: --region flag requires an argument")
				return
			}
		case "--quiet":
			quiet = true
		case "--preserve-data":
			preserveData = true
		case "open": // Assuming "open" might also need a runID
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				runID = args[i+1]
				i++
			}
			// No error here, as "open" without runID might be valid
		case "run":
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				runID = args[i+1]
				i++
			} else {
				fmt.Println("Error: 'run' command requires a runID argument")
				return
			}
		}
	}

	// Extract environment variables from command-line arguments
	envVars := make(map[string]string)
	for _, arg := range args {
		// Skip flags and commands
		if strings.HasPrefix(arg, "-") || arg == command {
			continue
		}
		parts := strings.Split(arg, "=")
		if len(parts) == 2 {
			envVars[parts[0]] = parts[1]
		}
	}

	switch command {
	case "deploy":
		env := "prod"
		if len(args) > 0 && !strings.HasPrefix(args[0], "-") { // Check if a service name is provided
			env = args[0]
		}
		cmd.DeployApplication(projectID, region, envVars, env, quiet)
	case "destroy":
		cmd.DestroyResources(projectID, region, preserveData, quiet)
	case "update":
		env := "prod"
		if len(args) > 0 && !strings.HasPrefix(args[0], "-") { // Check if a service name is provided
			env = args[0]
		}
		cmd.UpdateApplication(projectID, region, env, quiet)
	case "execute":
		if len(args) < 1 {
			fmt.Println("Usage: litmus execute <payload>")
			return
		}
		payload := args[0]
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
	case "start":
		// 1. Handle TEMPLATE_ID
		if len(args) < 1 {
			fmt.Println("Error: 'start' command requires a TEMPLATE_ID argument")
			return
		}
		templateID := args[0]

		// 2. Handle RUN_ID (generate if not provided)
		runID := ""
		if len(args) >= 2 { // Check if runID is provided
			runID = args[1]
		} else {
			runID = uuid.New().String() // Generate a random UUID
			fmt.Printf("Generated Run ID: %s\n", runID)
		}

		// 3. Get AUTH_TOKEN (optional)
		authToken := os.Getenv("AUTH_TOKEN")

		// Example: Assuming cmd.SubmitRun takes templateID, runID, and optionally authToken
		err := cmd.SubmitRun(templateID, runID, projectID, authToken)
		if err != nil {
			fmt.Printf("Error submitting run: %v\n", err)
			return
		}

		fmt.Println("Run submitted successfully.")
	case "status":
		cmd.ShowStatus(projectID)
	case "version":
		utils.DisplayVersion()
	case "analytics":
		if len(args) < 1 {
			fmt.Println("Invalid analytics subcommand.")
			fmt.Println("Usage: litmus analytics [deploy | destroy]")
			return
		}

		subcommand := args[0]
		switch subcommand {
		case "deploy":
			err := analytics.DeployAnalytics(projectID, region, quiet)
			if err != nil {
				utils.HandleGcloudError(err)
			}
		case "destroy":
			err := analytics.DestroyAnalytics(projectID, region, quiet)
			if err != nil {
				utils.HandleGcloudError(err)
			}
		default:
			fmt.Println("Invalid analytics subcommand:", subcommand)
			fmt.Println("Usage: litmus analytics [deploy | destroy]")
		}
	case "proxy":
		if len(args) < 1 {
			fmt.Println("Invalid proxy subcommand.")
			fmt.Println("Usage: litmus proxy [deploy --upstreamURL <upstreamURL> | list | destroy <service_name> | destroy-all]")
			return
		}

		subcommand := args[0]
		switch subcommand {
		case "deploy":
			var upstreamURL string
			if len(args) >= 3 && args[1] == "--upstreamURL" {
				upstreamURL = args[2]
			}
			err := cmd.DeployProxy(projectID, region, upstreamURL, quiet)
			if err != nil {
				utils.HandleGcloudError(err)
			}
		case "list":
			_, err := cmd.ListProxyServices(projectID, quiet)
			if err != nil {
				utils.HandleGcloudError(err)
			}
		case "destroy":
			var serviceName string
			if len(args) >= 2 { // Check if a service name is provided
				serviceName = args[1]
			}
			err := cmd.DestroyProxyService(projectID, serviceName, region, quiet)
			if err != nil {
				utils.HandleGcloudError(err)
			}
		case "destroy-all":
			err := cmd.DestroyAllProxyServices(projectID, region, quiet)
			if err != nil {
				utils.HandleGcloudError(err)
			}
		default:
			fmt.Println("Invalid proxy subcommand:", subcommand)
			fmt.Println("Usage: litmus proxy [deploy --upstreamURL <upstreamURL> | list | destroy <service_name> | destroy-all]")
		}
	default:
		fmt.Println("Invalid command:", command)
		utils.PrintUsage()
	}
}