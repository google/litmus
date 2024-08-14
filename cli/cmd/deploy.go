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
	"github.com/google/litmus/cli/utils"
)

// DeployApplication deploys the Litmus application to Google Cloud.
func DeployApplication(projectID, region string, envVars map[string]string) {
	// --- Confirm deployment ---
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("\nThis will deploy Litmus resources in the project '%s'. Are you sure you want to continue? (y/N): ", projectID)
	confirmation, _ := reader.ReadString('\n')
	confirmation = strings.TrimSpace(confirmation) // Remove leading/trailing whitespace
	if strings.ToLower(confirmation) != "y" {
		fmt.Println("\nAborting deployment.")
		return
	}

	// Enable required APIs
	apisToEnable := []string{
		"run.googleapis.com",
		"firestore.googleapis.com",
		"iam.googleapis.com",
		"aiplatform.googleapis.com",
		"secretmanager.googleapis.com",
	}

	for _, api := range apisToEnable {
		if !utils.IsAPIEnabled(api, projectID) {
			s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
			s.Suffix = fmt.Sprintf(" Enabling API %s... ", api)
			s.Start()
			enableAPICmd := exec.Command("gcloud", "services", "enable", api, "--project", projectID)
			if err := enableAPICmd.Run(); err != nil {
				s.Stop() // Stop the spinner in case of error
				log.Fatalf("Error enabling API %s: %v", api, err)
			}
			s.Stop()
			fmt.Printf("\nDone! API %s enabled!", api)
		} else {
			fmt.Printf("\nAPI %s is already enabled.", api)
		}
	}

	// Check if Firestore database exists
	if !utils.FirestoreDatabaseExists(projectID) {
		// Create default Firestore database
		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = " Creating default Firestore database... "
		s.Start()
		createFirestoreCmd := exec.Command(
			"gcloud", "firestore", "databases", "create",
			"--project", projectID,
			"--location", region,
		)
		if err := createFirestoreCmd.Run(); err != nil {
			s.Stop() // Stop the spinner in case of error
			log.Fatalf("\nError creating Firestore database: %v", err)
		}
		s.Stop()
		fmt.Println("\nDone! Firestore created!")
	} else {
		fmt.Println("\nFirestore database already exists.")
	}

	// --- Service Account for API ---
	apiServiceAccount := fmt.Sprintf("%s-api@%s.iam.gserviceaccount.com", projectID, projectID)
	if !utils.ServiceAccountExists(projectID, apiServiceAccount) {
		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = fmt.Sprintf(" Creating service account for API: %s... ", apiServiceAccount)
		s.Start()
		createServiceAccountCmd := exec.Command(
			"gcloud", "iam", "service-accounts", "create",
			fmt.Sprintf("%s-api", projectID),
			"--project", projectID,
			"--display-name", "Litmus API Service Account",
		)
		if err := createServiceAccountCmd.Run(); err != nil {
			s.Stop() // Stop the spinner in case of error
			log.Fatalf("Error creating service account: %v\n", err)
		}
		s.Stop()
		fmt.Printf("Done! Service account for API created: %s\n", apiServiceAccount)
	} else {
		fmt.Printf("Service account for API already exists: %s (skipping)\n", apiServiceAccount)
	}

	// --- Service Account for Worker ---
	workerServiceAccount := fmt.Sprintf("%s-worker@%s.iam.gserviceaccount.com", projectID, projectID)
	if !utils.ServiceAccountExists(projectID, workerServiceAccount) {
		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = fmt.Sprintf(" Creating service account for Worker: %s... ", workerServiceAccount)
		s.Start()
		createWorkerServiceAccountCmd := exec.Command(
			"gcloud", "iam", "service-accounts", "create",
			fmt.Sprintf("%s-worker", projectID),
			"--project", projectID,
			"--display-name", "Litmus Worker Service Account",
		)
		if err := createWorkerServiceAccountCmd.Run(); err != nil {
			s.Stop() // Stop the spinner in case of error
			log.Fatalf("Error creating service account: %v\n", err)
		}
		s.Stop()
		fmt.Printf("Done! Service account for Worker created: %s\n", workerServiceAccount)
	} else {
		fmt.Printf("Service account for Worker already exists: %s (skipping)\n", workerServiceAccount)
	}

	// --- Grant Vertex AI and Firestore permissions to API service account ---
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Granting permissions to API service account... "
	s.Start()
	if err := grantPermissions(apiServiceAccount, projectID); err != nil {
		s.Stop() // Stop the spinner in case of error
		log.Fatalf("Error granting permissions to API service account: %v \n", err)
	}
	s.Stop()
	fmt.Printf("Done! Granted permissions to API service account\n")

	// --- Grant Vertex AI and Firestore permissions to Worker service account ---
	s = spinner.New(spinner.CharSets[14], 100*time.Millisecond) // Create a new spinner instance
	s.Suffix = " Granting permissions to Worker service account... "
	s.Start()
	if err := grantPermissions(workerServiceAccount, projectID); err != nil {
		s.Stop() // Stop the spinner in case of error
		log.Fatalf("Error granting permissions to Worker service account: %v\n", err)
	}
	s.Stop()
	fmt.Printf("Done! Granted permissions to Worker service account\n")

	// --- Password and URL Management with Secret Manager ---
	s = spinner.New(spinner.CharSets[14], 100*time.Millisecond) // Create a new spinner instance
	s.Suffix = " Getting or creating passwords... "
	s.Start()

	// Get or create passwords and store them in Secret Manager
	password, err := utils.AccessSecret(projectID, "litmus-password")
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			// Generate and store password if it doesn't exist
			password = utils.GenerateRandomPassword(16)
			if err := utils.CreateOrUpdateSecret(projectID, "litmus-password", password); err != nil {
				s.Stop() // Stop the spinner in case of error
				log.Fatalf("Error storing password in Secret Manager: %v", err)
			}
		} else {
			s.Stop() // Stop the spinner in case of error
			log.Fatalf("Error accessing password in Secret Manager: %v", err)
		}
	}
	envVars["PASSWORD"] = password
	s.Stop()

	// --- Deploy Cloud Run service with service account ---
	s = spinner.New(spinner.CharSets[14], 100*time.Millisecond) // Create a new spinner instance
	s.Suffix = " Deploying Cloud Run service 'litmus-api'... "
	s.Start()

	// Construct the deploy command with --no-traffic flag for updates
	deployServiceCmd := exec.Command(
		"gcloud", "run", "deploy", "litmus-api",
		"--project", projectID,
		"--region", region,
		"--allow-unauthenticated",
		"--image", "europe-docker.pkg.dev/litmusai-prod/litmus/api:latest",
		"--service-account", apiServiceAccount,
		// Add other required/optional flags for your Cloud Run service
	)

	// Add environment variables to the command
	for name, value := range envVars {
		deployServiceCmd.Args = append(deployServiceCmd.Args, "--set-env-vars", fmt.Sprintf("%s=%s", name, value))
	}

	// Add Region
	deployServiceCmd.Args = append(deployServiceCmd.Args, "--set-env-vars", fmt.Sprintf("GCP_REGION=%s", region))
	// Add Project
	deployServiceCmd.Args = append(deployServiceCmd.Args, "--set-env-vars", fmt.Sprintf("GCP_PROJECT=%s", projectID))

	// Check if service already exists and add --no-traffic flag
	if utils.ServiceExists(projectID, region, "litmus-api") {
		deployServiceCmd.Args = append(deployServiceCmd.Args, "--no-traffic")
	}

	output2, err := deployServiceCmd.CombinedOutput()
	if err != nil {
		s.Stop() // Stop the spinner in case of error
		log.Fatalf("Error deploying Cloud Run service: %v\nOutput: %s\n", err, output2)
	}
	s.Stop()
	fmt.Println("Done! Deployed API.")

	// If the service was updated, route traffic back to the latest revision
	if strings.Contains(string(output2), "Routing traffic...") {
		s = spinner.New(spinner.CharSets[14], 100*time.Millisecond) // Create a new spinner instance
		s.Suffix = " Routing traffic to the latest revision... "
		s.Start()
		routeTrafficCmd := exec.Command(
			"gcloud", "run", "services", "update-traffic", "litmus-api",
			"--project", projectID,
			"--region", region,
			"--to-latest",
		)
		if err := routeTrafficCmd.Run(); err != nil {
			s.Stop() // Stop the spinner in case of error
			log.Fatalf("Error routing traffic to the latest revision: %v", err)
		}
		s.Stop()
		fmt.Println("Done! Routed traffic to the latest revision.")
	}

	// --- Deploy Cloud Run job with service account ---
	s = spinner.New(spinner.CharSets[14], 100*time.Millisecond) // Create a new spinner instance
	s.Suffix = " Deploying Cloud Run job 'litmus-worker'... "
	s.Start()

	// Construct the deploy command (always create new)
	deployJobCmd := exec.Command(
		"gcloud", "run", "jobs", "deploy", "litmus-worker", // Always use "create"
		"--project", projectID,
		"--region", region,
		"--image", "europe-docker.pkg.dev/litmusai-prod/litmus/worker:latest",
		"--service-account", workerServiceAccount,
		// Add other required/optional flags for your Cloud Run job
	)

	// Add environment variables to the command
	for name, value := range envVars {
		deployJobCmd.Args = append(deployJobCmd.Args, "--set-env-vars", fmt.Sprintf("%s=%s", name, value))
	}

	// Add Region
	deployJobCmd.Args = append(deployJobCmd.Args, "--set-env-vars", fmt.Sprintf("GCP_REGION=%s", region))
	// Add Project
	deployJobCmd.Args = append(deployJobCmd.Args, "--set-env-vars", fmt.Sprintf("GCP_PROJECT=%s", projectID))

	// Check if job already exists and change command to use --update-job flag
	if utils.JobExists(projectID, region, "litmus-worker") {
		deployJobCmd.Args[3] = "update"
	}

	if err := deployJobCmd.Run(); err != nil {
		s.Stop() // Stop the spinner in case of error
		log.Fatalf("Error deploying Cloud Run job: %v\n", err)
	}
	s.Stop()
	fmt.Println("Done! Deployed Worker")

	// --- Grant API permission to invoke Worker ---
	if !utils.BindingExists(projectID, region, "litmus-worker", apiServiceAccount, "roles/run.invoker") {
		s = spinner.New(spinner.CharSets[14], 100*time.Millisecond) // Create a new spinner instance
		s.Suffix = " Granting API permission to invoke Worker... "
		s.Start()
		grantPermissionCmd := exec.Command(
			"gcloud", "run", "jobs", "add-iam-policy-binding", "litmus-worker",
			"--member", fmt.Sprintf("serviceAccount:%s", apiServiceAccount),
			"--role", "roles/run.invoker",
			"--project", projectID,
			"--region", region,
		)
		if err := grantPermissionCmd.Run(); err != nil {
			s.Stop() // Stop the spinner in case of error
			log.Fatalf("Error granting permission: %v\n", err)
		}
		s.Stop()
		fmt.Println("Done! Granting API permission to invoke Worker.\n")
	} else {
		fmt.Println("API permission to invoke Worker already exists.\n")
	}

	// Extract and print the service URL
	serviceURL := utils.ExtractServiceURL(string(output2))

	// Store the service URL in Secret Manager
	s = spinner.New(spinner.CharSets[14], 100*time.Millisecond) // Create a new spinner instance
	s.Suffix = " Storing service URL... "
	s.Start()
	if err := utils.CreateOrUpdateSecret(projectID, "litmus-service-url", serviceURL); err != nil {
		s.Stop() // Stop the spinner in case of error
		log.Fatalf("Error storing service URL in Secret Manager: %v", err)
	}
	s.Stop()

	fmt.Println("\nAll deployments completed \n")

	fmt.Println("Get started now by visiting: ", serviceURL)
	fmt.Println("User: admin")
	fmt.Println("Password: ", password)
}

// grantPermissions grants Vertex AI and Firestore permissions to the given service account.
func grantPermissions(serviceAccount, projectID string) error {
	roles := []string{
		"roles/aiplatform.user",
		"roles/datastore.user",
		"roles/logging.logWriter",
		"roles/run.developer",
	}

	for _, role := range roles {
		if !utils.BindingExists(projectID, "", "", serviceAccount, role) { // No region needed for project-level bindings
			cmd := exec.Command(
				"gcloud", "projects", "add-iam-policy-binding", projectID,
				"--member", fmt.Sprintf("serviceAccount:%s", serviceAccount),
				"--role", role,
			)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("error granting role '%s': %v", role, err)
			}
		} else {
			fmt.Printf("Role '%s' already granted to service account.\n", role)
		}
	}

	return nil
}
