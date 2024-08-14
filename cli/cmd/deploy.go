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

	"github.com/google/litmus/cli/utils"
)

// DeployApplication deploys the Litmus application to Google Cloud.
func DeployApplication(projectID, region string, envVars map[string]string) {
	// --- Confirm deployment ---
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("\nThis will deploy Litmus resources to project '%s'. Are you sure you want to continue? (y/N): ", projectID)
	confirmation, _ := reader.ReadString('\n')
	confirmation = strings.TrimSpace(confirmation) // Remove leading/trailing whitespace
	if strings.ToLower(confirmation) != "y" {
		fmt.Println("Aborting deployment.")
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
			fmt.Printf("Enabling API %s ", api)
			enableAPICmd := exec.Command("gcloud", "services", "enable", api, "--project", projectID)
			utils.ShowInProgress(enableAPICmd)
			if err := enableAPICmd.Run(); err != nil {
				log.Fatalf("Error enabling API %s: %v", api, err)
			}
			fmt.Println("Done!")
		} else {
			fmt.Printf("API %s is already enabled.\n", api)
		}
	}

	// Check if Firestore database exists
	if !utils.FirestoreDatabaseExists(projectID) {
		// Create default Firestore database
		fmt.Print("\nCreating default Firestore database ")
		createFirestoreCmd := exec.Command(
			"gcloud", "firestore", "databases", "create",
			"--project", projectID,
			"--location", region,
		)
		utils.ShowInProgress(createFirestoreCmd)
		if err := createFirestoreCmd.Run(); err != nil {
			log.Fatalf("\nError creating Firestore database: %v", err)
		}
		fmt.Println("Done!")
	} else {
		fmt.Println("\nFirestore database already exists.")
	}

	// --- Service Account for API ---
	apiServiceAccount := fmt.Sprintf("%s-api@%s.iam.gserviceaccount.com", projectID, projectID)
	if !utils.ServiceAccountExists(projectID, apiServiceAccount) {
		fmt.Printf("Creating service account for API: %s ", apiServiceAccount)
		createServiceAccountCmd := exec.Command(
			"gcloud", "iam", "service-accounts", "create",
			fmt.Sprintf("%s-api", projectID),
			"--project", projectID,
			"--display-name", "Litmus API Service Account",
		)
		utils.ShowInProgress(createServiceAccountCmd)
		if err := createServiceAccountCmd.Run(); err != nil {
			log.Fatalf("\nError creating service account: %v", err)
		}
		fmt.Println("Done!")
	} else {
		fmt.Printf("\nService account for API already exists: %s\n", apiServiceAccount)
	}

	// --- Service Account for Worker ---
	workerServiceAccount := fmt.Sprintf("%s-worker@%s.iam.gserviceaccount.com", projectID, projectID)
	if !utils.ServiceAccountExists(projectID, workerServiceAccount) {
		fmt.Printf("\nCreating service account for Worker: %s ", workerServiceAccount)
		createWorkerServiceAccountCmd := exec.Command(
			"gcloud", "iam", "service-accounts", "create",
			fmt.Sprintf("%s-worker", projectID),
			"--project", projectID,
			"--display-name", "Litmus Worker Service Account",
		)
		utils.ShowInProgress(createWorkerServiceAccountCmd)
		if err := createWorkerServiceAccountCmd.Run(); err != nil {
			log.Fatalf("\nError creating service account: %v", err)
		}
		fmt.Println("Done!")
	} else {
		fmt.Printf("\nService account for Worker already exists: %s\n", workerServiceAccount)
	}

	// --- Grant Vertex AI and Firestore permissions to API service account ---
	fmt.Print("\nGranting permissions to API service account... ")
	if err := grantPermissions(apiServiceAccount, projectID); err != nil {
		log.Fatalf("\nError granting permissions to API service account: %v", err)
	}
	fmt.Println("Done!")

	// --- Grant Vertex AI and Firestore permissions to Worker service account ---
	fmt.Print("\nGranting permissions to Worker service account... ")
	if err := grantPermissions(workerServiceAccount, projectID); err != nil {
		log.Fatalf("\nError granting permissions to Worker service account: %v", err)
	}
	fmt.Println("Done!")

	// --- Password and URL Management with Secret Manager ---

	// Get or create passwords and store them in Secret Manager
	password, err := utils.AccessSecret(projectID, "litmus-password")
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			// Generate and store password if it doesn't exist
			password = utils.GenerateRandomPassword(16)
			fmt.Printf("Generated random password: %s\n", password)
			if err := utils.CreateOrUpdateSecret(projectID, "litmus-password", password); err != nil {
				log.Fatalf("Error storing password in Secret Manager: %v", err)
			}
		} else {
			log.Fatalf("Error accessing password in Secret Manager: %v", err)
		}
	} else {
		fmt.Println("\nUsing existing password from Secret Manager.")
	}
	envVars["PASSWORD"] = password

	// --- Deploy Cloud Run service with service account ---
	fmt.Print("\nDeploying Cloud Run service 'litmus-api' ")

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

	utils.ShowInProgress(deployServiceCmd)
	output2, err := deployServiceCmd.CombinedOutput()
	if err != nil {
		log.Fatalf("\nError deploying Cloud Run service: %v\nOutput: %s", err, output2)
	}
	fmt.Println("Done!")

	// If the service was updated, route traffic back to the latest revision
	if strings.Contains(string(output2), "Routing traffic...") {
		fmt.Print("Routing traffic to the latest revision... ")
		routeTrafficCmd := exec.Command(
			"gcloud", "run", "services", "update-traffic", "litmus-api",
			"--project", projectID,
			"--region", region,
			"--to-latest",
		)
		utils.ShowInProgress(routeTrafficCmd)
		if err := routeTrafficCmd.Run(); err != nil {
			log.Fatalf("\nError routing traffic to the latest revision: %v", err)
		}
		fmt.Println("Done!")
	}

	// --- Deploy Cloud Run job with service account ---
	fmt.Print("\nDeploying Cloud Run job 'litmus-worker' ")

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
		deployJobCmd.Args[3] = "update" // Change "create" to "update"
	}

	utils.ShowInProgress(deployJobCmd)
	if err := deployJobCmd.Run(); err != nil {
		log.Fatalf("\nError deploying Cloud Run job: %v", err)
	}
	fmt.Println("Done!")

	// --- Grant API permission to invoke Worker ---
	if !utils.BindingExists(projectID, region, "litmus-worker", apiServiceAccount, "roles/run.invoker") {
		fmt.Print("\nGranting API permission to invoke Worker... ")
		grantPermissionCmd := exec.Command(
			"gcloud", "run", "jobs", "add-iam-policy-binding", "litmus-worker",
			"--member", fmt.Sprintf("serviceAccount:%s", apiServiceAccount),
			"--role", "roles/run.invoker",
			"--project", projectID,
			"--region", region,
		)
		utils.ShowInProgress(grantPermissionCmd)
		if err := grantPermissionCmd.Run(); err != nil {
			log.Fatalf("Error granting permission: %v", err)
		}
		fmt.Println("\nDone!")
	} else {
		fmt.Println("\nAPI permission to invoke Worker already exists.")
	}

	// Extract and print the service URL
	serviceURL := utils.ExtractServiceURL(string(output2))

	// Store the service URL in Secret Manager
	if err := utils.CreateOrUpdateSecret(projectID, "litmus-service-url", serviceURL); err != nil {
		log.Fatalf("\nError storing service URL in Secret Manager: %v", err)
	}

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
			utils.ShowInProgress(cmd)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("error granting role '%s': %v", role, err)
			}
		} else {
			fmt.Printf("Role '%s' already granted to service account.\n", role)
		}
	}

	return nil
}
