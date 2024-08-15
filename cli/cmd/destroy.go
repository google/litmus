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

// DestroyResources removes all resources created by the Litmus application.
func DestroyResources(projectID, region string, quiet bool) {
	if !quiet {
		if !utils.ConfirmPrompt(fmt.Sprintf("\nThis will delete all Litmus resources in the project '%s'. Are you sure you want to continue?", projectID)) {
			fmt.Println("Aborting destruction.")
			return
		}
	}

	deleteResource := func(resourceType, resourceName string) {
		var cmd *exec.Cmd
		if resourceType == "service" {
			cmd = exec.Command("gcloud", "run", "services", "delete", resourceName,
				"--project", projectID,
				"--region", region,
				"--quiet",
			)
		} else if resourceType == "job" {
			cmd = exec.Command("gcloud", "run", "jobs", "delete", resourceName,
				"--project", projectID,
				"--region", region,
				"--quiet",
			)
		} else if resourceType == "secret" {
			cmd = exec.Command("gcloud", "secrets", "delete", resourceName,
				"--project", projectID,
				"--quiet",
			)
		} else if resourceType == "serviceAccount" {
			cmd = exec.Command("gcloud", "iam", "service-accounts", "delete", resourceName,
				"--project", projectID,
				"--quiet",
			)
		} else {
			log.Fatalf("Invalid resource type: %s", resourceType)
		}

		if !quiet {
			s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
			s.Suffix = fmt.Sprintf(" Deleting %s '%s'... ", resourceType, resourceName)
			s.Start()
			defer s.Stop()
		}

		if err := cmd.Run(); err != nil {
			if !quiet {
				log.Printf("Error deleting %s: %v. You might need to delete it manually.\n", resourceType, err)
			}
		} else if !quiet {
			fmt.Printf("Done! Deleted %s '%s'.\n", resourceType, resourceName)
		}
	}

	// --- Delete Cloud Run service ---
	deleteResource("service", "litmus-api")

	// --- Delete Cloud Run job ---
	deleteResource("job", "litmus-worker")

	// --- Delete Secrets from Secret Manager ---
	secretsToDelete := []string{"litmus-password", "litmus-service-url"}
	for _, secretID := range secretsToDelete {
		deleteResource("secret", secretID)
	}

	// --- Delete Service Accounts ---
	serviceAccountsToDelete := []string{
		fmt.Sprintf("%s-api@%s.iam.gserviceaccount.com", projectID, projectID),
		fmt.Sprintf("%s-worker@%s.iam.gserviceaccount.com", projectID, projectID),
	}
	for _, sa := range serviceAccountsToDelete {
		deleteResource("serviceAccount", sa)
	}

	if !quiet {
		fmt.Println("\nResource destruction complete.")
	}
}
