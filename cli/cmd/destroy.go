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
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/briandowns/spinner"
)

// DestroyResources removes all resources created by the Litmus application.
func DestroyResources(projectID, region string) {
	// --- Confirm deletion ---
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("\nThis will delete all Litmus resources in project '%s'. Are you sure you want to continue? (y/N): ", projectID)
	confirmation, _ := reader.ReadString('\n')
	confirmation = strings.TrimSpace(confirmation) // Remove leading/trailing whitespace
	if strings.ToLower(confirmation) != "y" {
		fmt.Println("Aborting destruction.")
		return
	}

	// --- Delete Cloud Run service ---
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Deleting Cloud Run service 'litmus-api'... "
	s.Start()
	deleteServiceCmd := exec.Command("gcloud", "run", "services", "delete", "litmus-api",
		"--project", projectID,
		"--region", region,
		"--quiet",
	)
	if err := deleteServiceCmd.Run(); err != nil {
		s.Stop()
		log.Printf("Error deleting Cloud Run service: %v. You might need to delete it manually.\n", err)
	} else {
		s.Stop()
		fmt.Println("Done! Deleted Cloud Run service 'litmus-api'.")
	}

	// --- Delete Cloud Run job ---
	s = spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Deleting Cloud Run job 'litmus-worker'... "
	s.Start()
	deleteJobCmd := exec.Command("gcloud", "run", "jobs", "delete", "litmus-worker",
		"--project", projectID,
		"--region", region,
		"--quiet",
	)
	if err := deleteJobCmd.Run(); err != nil {
		s.Stop()
		log.Printf("\nError deleting Cloud Run job: %v. You might need to delete it manually.\n", err)
	} else {
		s.Stop()
		fmt.Println("Done! Deleted Cloud Run job 'litmus-worker'.")
	}

	// --- Delete Secrets from Secret Manager ---
	secretsToDelete := []string{"litmus-password", "litmus-service-url"}
	for _, secretID := range secretsToDelete {
		s = spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = fmt.Sprintf(" Deleting Secret '%s'... ", secretID)
		s.Start()
		deleteSecretCmd := exec.Command("gcloud", "secrets", "delete", secretID,
			"--project", projectID,
			"--quiet",
		)
		if err := deleteSecretCmd.Run(); err != nil {
			s.Stop()
			log.Printf("\nError deleting Secret: %v. You might need to delete it manually.\n", err)
		} else {
			s.Stop()
			fmt.Println("Done! Deleted Secret '%s'.", secretID)
		}
	}

	// --- Delete Service Accounts ---
	serviceAccountsToDelete := []string{
		fmt.Sprintf("%s-api@%s.iam.gserviceaccount.com", projectID, projectID),
		fmt.Sprintf("%s-worker@%s.iam.gserviceaccount.com", projectID, projectID),
	}
	for _, sa := range serviceAccountsToDelete {
		s = spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = fmt.Sprintf(" Deleting Service Account '%s'... ", sa)
		s.Start()
		deleteSaCmd := exec.Command("gcloud", "iam", "service-accounts", "delete", sa,
			"--project", projectID,
			"--quiet",
		)
		if err := deleteSaCmd.Run(); err != nil {
			s.Stop()
			log.Printf("\nError deleting Service Account: %v. You might need to delete it manually.\n", err)
		} else {
			s.Stop()
			fmt.Println("Done! Deleted Service Account '%s'.", sa)
		}
	}

	fmt.Println("\nResource destruction complete.")
}
