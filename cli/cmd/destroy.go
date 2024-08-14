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

// DestroyResources removes all resources created by the Litmus application.
func DestroyResources(projectID, region string) {
	fmt.Println("Destroying resources in project:", projectID)

	// --- Delete Cloud Run service ---
	fmt.Print("\nDeleting Cloud Run service 'litmus-api'... ")
	deleteServiceCmd := exec.Command("gcloud", "run", "services", "delete", "litmus-api",
		"--project", projectID,
		"--region", region,
		"--quiet",
	)
	if err := deleteServiceCmd.Run(); err != nil {
		log.Printf("Error deleting Cloud Run service: %v. You might need to delete it manually.\n", err)
	} else {
		fmt.Println("Done!")
	}

	// --- Delete Cloud Run job ---
	fmt.Print("\nDeleting Cloud Run job 'litmus-worker'... ")
	deleteJobCmd := exec.Command("gcloud", "run", "jobs", "delete", "litmus-worker",
		"--project", projectID,
		"--region", region,
		"--quiet",
	)
	if err := deleteJobCmd.Run(); err != nil {
		log.Printf("\nError deleting Cloud Run job: %v. You might need to delete it manually.\n", err)
	} else {
		fmt.Println("Done!")
	}

	// --- Delete Secrets from Secret Manager ---
	secretsToDelete := []string{"litmus-password", "litmus-service-url"}
	for _, secretID := range secretsToDelete {
		fmt.Printf("Deleting Secret '%s'... ", secretID)
		deleteSecretCmd := exec.Command("gcloud", "secrets", "delete", secretID,
			"--project", projectID,
			"--quiet",
		)
		if err := deleteSecretCmd.Run(); err != nil {
			log.Printf("\nError deleting Secret: %v. You might need to delete it manually.\n", err)
		} else {
			fmt.Println("Done!")
		}
	}

	// --- Delete Service Accounts ---
	serviceAccountsToDelete := []string{
		fmt.Sprintf("%s-api@%s.iam.gserviceaccount.com", projectID, projectID),
		fmt.Sprintf("%s-worker@%s.iam.gserviceaccount.com", projectID, projectID),
	}
	for _, sa := range serviceAccountsToDelete {
		fmt.Printf("\nDeleting Service Account '%s'... ", sa)
		deleteSaCmd := exec.Command("gcloud", "iam", "service-accounts", "delete", sa,
			"--project", projectID,
			"--quiet",
		)
		if err := deleteSaCmd.Run(); err != nil {
			log.Printf("\nError deleting Service Account: %v. You might need to delete it manually.\n", err)
		} else {
			fmt.Println("Done!")
		}
	}

	fmt.Println("\nResource destruction complete.")
}