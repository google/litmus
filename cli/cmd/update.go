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
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/google/litmus/cli/utils"
)

// UpdateApplication updates the Litmus application to the latest version.
func UpdateApplication(projectID, region string, env string, quiet bool) {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)

	if !quiet {
		if !utils.ConfirmPrompt(fmt.Sprintf("\nThis will update Litmus resources in the project '%s'. Are you sure you want to continue?", projectID)) {
			fmt.Println("\nAborting update.")
			return
		}
	}

    // --- Update Cloud Run service ---
	if !quiet {
		s.Suffix = " Updating Cloud Run service 'litmus-api'... "
		s.Start()
		defer s.Stop()
	}

	apiImage := fmt.Sprintf("europe-docker.pkg.dev/litmusai-%s/litmus/api:latest",env)

	updateServiceCmd := exec.Command(
		"gcloud", "run", "deploy", "litmus-api",
		"--project", projectID,
		"--region", region,
		"--image", apiImage, 
		"--no-traffic", // Stop traffic during the update
	)
	output, err := updateServiceCmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error updating Cloud Run service: %v\nOutput: %s", err, output)
	}

	if !quiet {
		fmt.Println("Done! Updated API.\n")
	}
	// Route traffic back to the updated service
	if !quiet {
		s.Suffix = " Routing traffic to the updated service... "
		s.Start()
		defer s.Stop()
	}

	routeTrafficCmd := exec.Command(
		"gcloud", "run", "services", "update-traffic", "litmus-api",
		"--project", projectID,
		"--region", region,
		"--to-latest",
	)
	if err := routeTrafficCmd.Run(); err != nil {
		log.Fatalf("Error routing traffic to the updated service: %v", err)
	}

	if !quiet {
		fmt.Println("Done! Routed traffic to the updated service.")
	}

    // --- Update Cloud Run job ---
	if !quiet {
		s.Suffix = " Updating Cloud Run job 'litmus-worker'... "
		s.Start()
		defer s.Stop()
	}

	workerImage := fmt.Sprintf("europe-docker.pkg.dev/litmusai-%s/litmus/worker:latest",env)

	updateJobCmd := exec.Command(
		"gcloud", "run", "jobs", "update", "litmus-worker", 
		"--project", projectID,
		"--region", region,
		"--image", workerImage, 
	)
	output, err = updateJobCmd.CombinedOutput()
	if err != nil {
		if !strings.Contains(string(output), "already exists with the same image") {
			log.Fatalf("Error updating Cloud Run job: %v\nOutput: %s", err, output)
		} else if !quiet { // If the job exists with the same image, inform the user
			fmt.Println("Cloud Run job already up-to-date.")
		}
	} else if !quiet {
		fmt.Println("Done! Updated Worker.")
	}

	if !quiet {
		fmt.Println("\nLitmus application updated successfully!")
	}
}