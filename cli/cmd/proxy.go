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
	"os/exec"
	"time"

	"github.com/briandowns/spinner"
	"github.com/google/litmus/cli/utils"
)

// ProxyService represents a deployed Litmus proxy Cloud Run service.
type ProxyService struct {
	Name        string
	ProjectID   string
	Region      string
	UpstreamURL string
}

// DeployProxy deploys a Litmus proxy to Google Cloud Run.
func DeployProxy(projectID, region, upstreamURL string, quiet bool) error {
	if projectID == "" {
		var err error
		projectID, err = utils.GetDefaultProjectID()
		if err != nil {
			utils.HandleGcloudError(err)
			return err
		}
	}

	if region == "" {
		region = "us-central1" // Default region
	}

	if upstreamURL == "" {
		upstreamURL, err := utils.SelectUpstreamURL()
		if err != nil {
			return err
		}
		if upstreamURL == "" {
			return fmt.Errorf("no upstream URL selected")
		}
	}

	// Generate a unique service name
	serviceName := fmt.Sprintf("litmus-proxy-%d", time.Now().UnixNano())

	if !quiet {
		// --- Confirm deployment ---
		if !utils.ConfirmPrompt(fmt.Sprintf("\nThis will deploy the Litmus proxy '%s' in the project '%s' and region '%s'. Are you sure you want to continue?", serviceName, projectID, region)) {
			fmt.Println("\nAborting deployment.")
			return nil
		}
	}

	if !quiet {
		// --- Deploy Cloud Run service ---
		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond) // Create a new spinner instance
		s.Suffix = fmt.Sprintf(" Deploying Cloud Run service '%s'...", serviceName)
		s.Start()
		defer s.Stop()
	}

	// Construct the deploy command
	deployCmd := exec.Command(
		"gcloud", "run", "deploy", serviceName,
		"--image", "europe-docker.pkg.dev/litmusai-dev/litmus/proxy:latest",
		"--project", projectID,
		"--region", region,
		"--allow-unauthenticated",
		"--set-env-vars", fmt.Sprintf("PROJECT_ID=%s,UPSTREAM_URL=%s", projectID, upstreamURL),
	)

	output, err := deployCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error deploying Cloud Run service: %v\nOutput: %s", err, output)
	}

	if !quiet {
		fmt.Println("Done! Deployed Proxy.")
	}

	// Extract and print the service URL
	serviceURL := utils.ExtractServiceURL(string(output))
	if !quiet {
		fmt.Println("\nAll deployments completed \n")
		fmt.Printf("Proxy URL for '%s': %s\n", serviceName, serviceURL)
	}

	return nil
}

// ListProxyServices lists all deployed Litmus proxy Cloud Run services.
func ListProxyServices(projectID string, quiet bool) ([]ProxyService, error) {
	if projectID == "" {
		var err error
		projectID, err = utils.GetDefaultProjectID()
		if err != nil {
			utils.HandleGcloudError(err)
			return nil, err
		}
	}

	cmd := exec.Command(
		"gcloud", "run", "services", "list",
		"--project", projectID,
		"--filter", "name~litmus-proxy", // Filter by services starting with "litmus-proxy"
		"--format=json",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error listing Cloud Run services: %v\nOutput: %s", err, output)
	}

	var services []map[string]interface{}
	if err := json.Unmarshal(output, &services); err != nil {
		return nil, fmt.Errorf("error parsing JSON output: %v", err)
	}

	var proxyServices []ProxyService
	for _, service := range services {
		metadata := service["metadata"].(map[string]interface{})
		spec := service["spec"].(map[string]interface{})
		template := spec["template"].(map[string]interface{})
		metadataAnnotations := template["metadata"].(map[string]interface{})["annotations"].(map[string]interface{})

		proxyServices = append(proxyServices, ProxyService{
			Name:        metadata["name"].(string),
			ProjectID:   projectID,
			Region:      metadataAnnotations["run.googleapis.com/region"].(string),
			UpstreamURL: metadataAnnotations["run.googleapis.com/ingress-settings"].(string),
		})
	}

	if !quiet {
		if len(proxyServices) > 0 {
			fmt.Println("Deployed Litmus Proxy services:")
			for _, s := range proxyServices {
				fmt.Printf("- Name: %s\n", s.Name)
				fmt.Printf("  Project ID: %s\n", s.ProjectID)
				fmt.Printf("  Region: %s\n", s.Region)
				fmt.Printf("  Upstream URL: %s\n", s.UpstreamURL)
				fmt.Println("--------------------")
			}
		} else {
			fmt.Println("No Litmus Proxy services found.")
		}
	}

	return proxyServices, nil
}

// DeleteProxyService deletes a deployed Litmus proxy Cloud Run service.
func DeleteProxyService(projectID, serviceName string, quiet bool) error {
	if projectID == "" {
		var err error
		projectID, err = utils.GetDefaultProjectID()
		if err != nil {
			utils.HandleGcloudError(err)
			return err
		}
	}

	if !quiet {
		// --- Confirm deletion ---
		if !utils.ConfirmPrompt(fmt.Sprintf("\nThis will delete the Litmus proxy service '%s' in the project '%s'. Are you sure you want to continue?", serviceName, projectID)) {
			fmt.Println("\nAborting deletion.")
			return nil
		}
	}

	// Construct the delete command
	deleteCmd := exec.Command(
		"gcloud", "run", "services", "delete", serviceName,
		"--project", projectID,
		"--quiet", // Assume quiet for deletion unless specified otherwise
	)

	output, err := deleteCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error deleting Cloud Run service: %v\nOutput: %s", err, output)
	}

	if !quiet {
		fmt.Printf("Successfully deleted service '%s'\n", serviceName)
	}

	return nil
}

// DeleteAllProxyServices deletes all deployed Litmus proxy Cloud Run services.
func DeleteAllProxyServices(projectID string, quiet bool) error {
	if projectID == "" {
		var err error
		projectID, err = utils.GetDefaultProjectID()
		if err != nil {
			utils.HandleGcloudError(err)
			return err
		}
	}

	services, err := ListProxyServices(projectID, quiet)
	if err != nil {
		return err
	}

	if len(services) == 0 {
		if !quiet {
			fmt.Println("No Litmus Proxy services found.")
		}
		return nil
	}

	if !quiet {
		// --- Confirm deletion ---
		if !utils.ConfirmPrompt(fmt.Sprintf("\nThis will delete ALL Litmus proxy services in the project '%s'. Are you sure you want to continue?", projectID)) {
			fmt.Println("\nAborting deletion.")
			return nil
		}
	}

	for _, s := range services {
		err := DeleteProxyService(projectID, s.Name, quiet)
		if err != nil {
			return err
		}
	}

	if !quiet {
		fmt.Println("All Litmus Proxy services deleted.")
	}

	return nil
}
