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
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main() {
	// Get project ID from command-line arguments (or hardcode it)
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run deploy <project-id> [region] [env-var1=value1] [env-var2=value2] ...")
		return
	}
	projectID := os.Args[1]

	// Optional region argument (defaults to "us-central1")
	region := "us-central1"
	if len(os.Args) > 2 {
		region = os.Args[2]
	}

	// Extract environment variables from command-line arguments
	envVars := make(map[string]string)
	for i := 3; i < len(os.Args); i++ {
		parts := strings.Split(os.Args[i], "=")
		if len(parts) == 2 {
			envVars[parts[0]] = parts[1]
		}
	}

	// Enable required APIs
	apisToEnable := []string{
		"artifactregistry.googleapis.com",
		"cloudbuild.googleapis.com",
		"run.googleapis.com",
		"firestore.googleapis.com",
		"iam.googleapis.com",        // Add IAM API for service account management
		"aiplatform.googleapis.com", // Enable Vertex AI API
	}

	for _, api := range apisToEnable {
		if !isAPIEnabled(api, projectID) {
			fmt.Printf("Enabling API %s ", api)
			enableAPICmd := exec.Command("gcloud", "services", "enable", api, "--project", projectID)
			go showInProgress(enableAPICmd)
			if err := enableAPICmd.Run(); err != nil {
				log.Fatalf("Error enabling API %s: %v", api, err)
			}
			fmt.Println("Done!")
		} else {
			fmt.Printf("API %s is already enabled.\n", api)
		}
	}

	// Check if Firestore database exists
	fmt.Print("Checking if Firestore database exists... ")
	listFirestoreCmd := exec.Command("gcloud", "firestore", "databases", "list", "--project", projectID)
	output, err := listFirestoreCmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error listing Firestore databases: %v\nOutput: %s", err, output)
	}

	if strings.Contains(string(output), "(default)") { // Check for "(default)" in output
		fmt.Println("Firestore database already exists.")
	} else {
		// Create default Firestore database
		fmt.Print("Creating default Firestore database ")
		createFirestoreCmd := exec.Command(
			"gcloud", "firestore", "databases", "create",
			"--project", projectID,
			"--location", region,
		)
		go showInProgress(createFirestoreCmd)
		if err := createFirestoreCmd.Run(); err != nil {
			log.Fatalf("Error creating Firestore database: %v", err)
		}
		fmt.Println("Done!")
	}

	// --- Service Account for API ---
	apiServiceAccount := fmt.Sprintf("%s-api@%s.iam.gserviceaccount.com", projectID, projectID)
	fmt.Printf("Creating/Updating service account for API: %s ", apiServiceAccount)
	createServiceAccountCmd := exec.Command(
		"gcloud", "iam", "service-accounts", "create",
		fmt.Sprintf("%s-api", projectID), // Service account name (without @...)
		"--project", projectID,
		"--display-name", "Litmus API Service Account",
	)
	go showInProgress(createServiceAccountCmd)
	if err := createServiceAccountCmd.Run(); err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			log.Fatalf("Error creating service account: %v", err)
		}
	}
	fmt.Println("Done!")

	// --- Service Account for Worker ---
	workerServiceAccount := fmt.Sprintf("%s-worker@%s.iam.gserviceaccount.com", projectID, projectID)
	fmt.Printf("Creating/Updating service account for Worker: %s ", workerServiceAccount)
	createWorkerServiceAccountCmd := exec.Command(
		"gcloud", "iam", "service-accounts", "create",
		fmt.Sprintf("%s-worker", projectID), // Service account name (without @...)
		"--project", projectID,
		"--display-name", "Litmus Worker Service Account",
	)
	go showInProgress(createWorkerServiceAccountCmd)
	if err := createWorkerServiceAccountCmd.Run(); err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			log.Fatalf("Error creating service account: %v", err)
		}
	}
	fmt.Println("Done!")

	// --- Grant Vertex AI and Firestore permissions to API service account ---
	fmt.Print("Granting permissions to API service account... ")
	if err := grantPermissions(apiServiceAccount, projectID); err != nil {
		log.Fatalf("Error granting permissions to API service account: %v", err)
	}
	fmt.Println("Done!")

	// --- Grant Vertex AI and Firestore permissions to Worker service account ---
	fmt.Print("Granting permissions to Worker service account... ")
	if err := grantPermissions(workerServiceAccount, projectID); err != nil {
		log.Fatalf("Error granting permissions to Worker service account: %v", err)
	}
	fmt.Println("Done!")

	// --- Deploy Cloud Run service with service account ---
	fmt.Print("Deploying Cloud Run service 'litmus-api' ")
	deployServiceCmd := exec.Command(
		"gcloud", "run", "deploy", "litmus-api",
		"--project", projectID,
		"--region", region,
		"--allow-unauthenticated",
		"--image", "gcr.io/litmusai-dev/litmus-ai-api:latest",
		"--service-account", apiServiceAccount, // Use the created service account
		// Add other required/optional flags for your Cloud Run service
	)

	// Add environment variables to the command
	for name, value := range envVars {
		deployServiceCmd.Args = append(deployServiceCmd.Args, "--set-env-vars", fmt.Sprintf("%s=%s", name, value))
	}

	// Add Region
	deployServiceCmd.Args = append(deployServiceCmd.Args, "--set-env-vars", fmt.Sprintf("%s=%s", "GCP_REGION", region))
	// Add Project
	deployServiceCmd.Args = append(deployServiceCmd.Args, "--set-env-vars", fmt.Sprintf("%s=%s", "GCP_PROJECT", projectID))

	go showInProgress(deployServiceCmd)
	output2, err := deployServiceCmd.CombinedOutput() // Capture command output
	if err != nil {
		log.Fatalf("Error deploying Cloud Run service: %v\nOutput: %s", err, output2)
	}
	fmt.Println("Done!")

	// --- Deploy Cloud Run job with service account ---
	fmt.Print("Deploying Cloud Run job 'litmus-worker' ")
	deployJobCmd := exec.Command(
		"gcloud", "run", "jobs", "create", "litmus-worker",
		"--project", projectID,
		"--region", region,
		"--image", "gcr.io/litmusai-dev/litmus-ai-worker:latest",
		"--service-account", workerServiceAccount, // Use the created service account
		// Add other required/optional flags for your Cloud Run job
	)

	// Add environment variables to the command
	for name, value := range envVars {
		deployJobCmd.Args = append(deployJobCmd.Args, "--set-env-vars", fmt.Sprintf("%s=%s", name, value))
	}

	// Add Region
	deployServiceCmd.Args = append(deployServiceCmd.Args, "--set-env-vars", fmt.Sprintf("%s=%s", "GCP_REGION", region))
	// Add Project
	deployServiceCmd.Args = append(deployServiceCmd.Args, "--set-env-vars", fmt.Sprintf("%s=%s", "GCP_PROJECT", projectID))

	go showInProgress(deployJobCmd)
	if err := deployJobCmd.Run(); err != nil {
		log.Fatalf("Error deploying Cloud Run job: %v", err)
	}
	fmt.Println("Done!")

	// --- Grant API permission to invoke Worker ---
	fmt.Print("Granting API permission to invoke Worker... ")
	grantPermissionCmd := exec.Command(
		"gcloud", "run", "jobs", "add-iam-policy-binding", "litmus-worker", // Replace with your worker service name
		"--member", fmt.Sprintf("serviceAccount:%s", apiServiceAccount),
		"--role", "roles/run.invoker",
		"--project", projectID,
		"--region", region,
	)
	go showInProgress(grantPermissionCmd)
	if err := grantPermissionCmd.Run(); err != nil {
		log.Fatalf("Error granting permission: %v", err)
	}
	fmt.Println("Done!")

	fmt.Println("\nAll deployments completed!")

	// Extract and print the service URL
	serviceURL := extractServiceURL(string(output2))
	fmt.Println("Get started now by visiting: ", serviceURL)
}

// grantPermissions grants Vertex AI and Firestore permissions to the given service account
func grantPermissions(serviceAccount, projectID string) error {
	roles := []string{
		"roles/aiplatform.user",   // Vertex AI access
		"roles/datastore.user",    // Firestore access
		"roles/logging.logWriter", // Logging
		"roles/run.developer",     //Run Invoker
	}

	for _, role := range roles {
		cmd := exec.Command(
			"gcloud", "projects", "add-iam-policy-binding", projectID,
			"--member", fmt.Sprintf("serviceAccount:%s", serviceAccount),
			"--role", role,
		)
		go showInProgress(cmd)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error granting role '%s': %v", role, err)
		}
	}

	return nil
}

// extractServiceURL extracts the service URL from the gcloud command output
func extractServiceURL(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "URL:") {
			parts := strings.Split(line, ": ")
			if len(parts) > 1 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return "" // Return empty string if URL is not found
}

// isAPIEnabled checks if a given API is enabled for the project
func isAPIEnabled(api, projectID string) bool {
	checkCmd := exec.Command("gcloud", "services", "list", "--project", projectID, "--enabled")
	output, err := checkCmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error checking API status: %v\nOutput: %s", err, output)
	}
	return strings.Contains(string(output), api)
}

// showInProgress displays an in-progress animation until the command finishes
func showInProgress(cmd *exec.Cmd) {
	done := make(chan struct{})
	defer close(done)
	go func() {
		<-done
	}()

	for {
		select {
		case <-done:
			return
		case <-time.After(500 * time.Millisecond):
			fmt.Print(".")
		}
	}
}
