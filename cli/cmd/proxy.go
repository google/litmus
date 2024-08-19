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

	"time"

	"github.com/briandowns/spinner"
	"github.com/google/litmus/cli/utils"
)

// DeployProxy deploys the Litmus proxy to Google Cloud Run.
func DeployProxy(projectID, region, upstreamURL string, quiet bool) {
	if !quiet {
		// --- Confirm deployment ---
		if !utils.ConfirmPrompt(fmt.Sprintf("\nThis will deploy the Litmus proxy in the project '%s'. Are you sure you want to continue?", projectID)) {
			fmt.Println("\nAborting deployment.")
			return
		}
	}

	if !quiet {
		// --- Deploy Cloud Run service ---
		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond) // Create a new spinner instance
		s.Suffix = " Deploying Cloud Run service 'litmus-proxy'... "
		s.Start()
		defer s.Stop()
	}

	// Construct the deploy command
	deployCmd := exec.Command(
		"gcloud", "run", "deploy", "litmus-proxy",
		"--image", "europe-docker.pkg.dev/litmusai-dev/litmus/proxy:latest",
		"--region", region,
		"--allow-unauthenticated",
		"--set-env-vars", fmt.Sprintf("PROJECT_ID=%s,UPSTREAM_URL=%s", projectID, upstreamURL),
	)

	output, err := deployCmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error deploying Cloud Run service: %v\nOutput: %s", err, output)
	}

	if !quiet {
		fmt.Println("Done! Deployed Proxy.")
	}

	// Extract and print the service URL
	serviceURL := utils.ExtractServiceURL(string(output))
	if !quiet {
		fmt.Println("\nAll deployments completed \n")
		fmt.Println("Proxy URL: ", serviceURL)
	}
}

// PrintUpstreamURLOptions prints a list of available upstream URLs.
func PrintUpstreamURLOptions() {
	upstreamURLs := []string{
		"asia-east1-aiplatform.googleapis.com",
		"asia-east2-aiplatform.googleapis.com",
		"asia-northeast1-aiplatform.googleapis.com",
		"asia-northeast2-aiplatform.googleapis.com",
		"asia-northeast3-aiplatform.googleapis.com",
		"asia-south1-aiplatform.googleapis.com",
		"asia-southeast1-aiplatform.googleapis.com",
		"asia-southeast2-aiplatform.googleapis.com",
		"australia-southeast1-aiplatform.googleapis.com",
		"australia-southeast2-aiplatform.googleapis.com",
		"europe-central2-aiplatform.googleapis.com",
		"europe-north1-aiplatform.googleapis.com",
		"europe-southwest1-aiplatform.googleapis.com",
		"europe-west1-aiplatform.googleapis.com",
		"europe-west2-aiplatform.googleapis.com",
		"europe-west3-aiplatform.googleapis.com",
		"europe-west4-aiplatform.googleapis.com",
		"europe-west6-aiplatform.googleapis.com",
		"europe-west8-aiplatform.googleapis.com",
		"europe-west9-aiplatform.googleapis.com",
		"me-west1-aiplatform.googleapis.com",
		"northamerica-northeast1-aiplatform.googleapis.com",
		"northamerica-northeast2-aiplatform.googleapis.com",
		"southamerica-east1-aiplatform.googleapis.com",
		"southamerica-west1-aiplatform.googleapis.com",
		"us-central1-aiplatform.googleapis.com",
		"us-east1-aiplatform.googleapis.com",
		"us-east4-aiplatform.googleapis.com",
		"us-south1-aiplatform.googleapis.com",
		"us-west1-aiplatform.googleapis.com",
		"us-west2-aiplatform.googleapis.com",
		"us-west3-aiplatform.googleapis.com",
		"us-west4-aiplatform.googleapis.com",
	}

	fmt.Println("Available upstream URLs:")
	for _, url := range upstreamURLs {
		fmt.Println(url)
	}
}
