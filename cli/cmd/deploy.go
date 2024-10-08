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
	"github.com/google/litmus/cli/analytics"
	"github.com/google/litmus/cli/utils"
)

// DeployApplication deploys the Litmus application to Google Cloud.
func DeployApplication(projectID, region string, envVars map[string]string, env string, quiet bool) {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond) // Create a new spinner instance
	if !quiet {
		// --- Confirm deployment ---
		if !utils.ConfirmPrompt(fmt.Sprintf("\nThis will deploy Litmus resources in the project '%s'. Are you sure you want to continue?", projectID)) {
			fmt.Println("\nAborting deployment.")
			return
		}
	}

	// Enable required APIs
	apisToEnable := []string{
		"run.googleapis.com",
		"firestore.googleapis.com",
		"iam.googleapis.com",
		"aiplatform.googleapis.com",
		"secretmanager.googleapis.com",
		"cloudresourcemanager.googleapis.com",
		"storage.googleapis.com", // Add Storage API
	}
	for _, api := range apisToEnable {
		if !utils.IsAPIEnabled(api, projectID) {
			if !quiet {
				s.Suffix = fmt.Sprintf(" Enabling API %s... ", api)
				s.Start()
				defer s.Stop()
			}
			enableAPICmd := exec.Command("gcloud", "services", "enable", api, "--project", projectID)
			output, err := enableAPICmd.CombinedOutput()
			if err != nil {
				log.Fatalf("Error enabling API %s: %v\nOutput: %s", api, err, output) // Print gcloud output
			}
			if !quiet {
				fmt.Printf("\nDone! API %s enabled!", api)
			}
		} else if !quiet {
			fmt.Printf("\nAPI %s is already enabled.", api)
		}
	}

	// Check if Firestore database exists
	if !utils.FirestoreDatabaseExists(projectID) {
		if !quiet {
			// Create default Firestore database
			s.Suffix = " Creating default Firestore database... "
			s.Start()
			defer s.Stop()
		}
		createFirestoreCmd := exec.Command(
			"gcloud", "firestore", "databases", "create",
			"--project", projectID,
			"--location", region,
		)
		output, err := createFirestoreCmd.CombinedOutput() // Capture gcloud output
		if err != nil {
			log.Fatalf("\nError creating Firestore database: %v\nOutput: %s", err, output)
		}
		if !quiet {
			fmt.Println("\nDone! Firestore created!")
		}
	} else if !quiet {
		fmt.Println("\nFirestore database already exists.")
	}

	// --- Create Files Bucket ---
	bucketName := fmt.Sprintf("%s-litmus-files", projectID)
	if !quiet {
		s.Suffix = fmt.Sprintf(" Creating files bucket '%s'... ", bucketName)
		s.Start()
		defer s.Stop()
	}
	if err := createFilesBucket(bucketName, region, projectID, quiet); err != nil {
		log.Fatalf("Error creating files bucket: %v\n", err)
	}
	if !quiet {
		fmt.Printf("Done! Created files bucket: %s\n", bucketName)
	}

	// --- Service Account for API ---
	apiServiceAccount := fmt.Sprintf("%s-api@%s.iam.gserviceaccount.com", projectID, projectID)
	if !utils.ServiceAccountExists(projectID, apiServiceAccount) {
		if !quiet {
			s.Suffix = fmt.Sprintf(" Creating service account for API: %s... ", apiServiceAccount)
			s.Start()
			defer s.Stop()
		}
		createServiceAccountCmd := exec.Command(
			"gcloud", "iam", "service-accounts", "create",
			fmt.Sprintf("%s-api", projectID),
			"--project", projectID,
			"--display-name", "Litmus API Service Account",
		)
		output, err := createServiceAccountCmd.CombinedOutput()
		if err != nil {
			log.Fatalf("Error creating service account: %v\nOutput: %s", err, output)
		}
		if !quiet {
			fmt.Printf("Done! Service account for API created: %s\n", apiServiceAccount)
		}
	} else if !quiet {
		fmt.Printf("Service account for API already exists: %s (skipping)\n", apiServiceAccount)
	}

	// --- Service Account for Worker ---
	workerServiceAccount := fmt.Sprintf("%s-worker@%s.iam.gserviceaccount.com", projectID, projectID)
	if !utils.ServiceAccountExists(projectID, workerServiceAccount) {
		if !quiet {
			s.Suffix = fmt.Sprintf(" Creating service account for Worker: %s... ", workerServiceAccount)
			s.Start()
			defer s.Stop()
		}
		createWorkerServiceAccountCmd := exec.Command(
			"gcloud", "iam", "service-accounts", "create",
			fmt.Sprintf("%s-worker", projectID),
			"--project", projectID,
			"--display-name", "Litmus Worker Service Account",
		)
		output, err := createWorkerServiceAccountCmd.CombinedOutput()
		if err != nil {
			log.Fatalf("Error creating service account: %v\nOutput: %s", err, output)
		}
		if !quiet {
			fmt.Printf("Done! Service account for Worker created: %s\n", workerServiceAccount)
		}
	} else if !quiet {
		fmt.Printf("Service account for Worker already exists: %s (skipping)\n", workerServiceAccount)
	}

	// --- Grant Vertex AI, Firestore, and Storage permissions to API service account ---
	if !quiet {
		s.Suffix = " Granting permissions to API service account... "
		s.Start()
		defer s.Stop()
	}
	if err := grantPermissions(apiServiceAccount, projectID, quiet, bucketName); err != nil {
		log.Fatalf("Error granting permissions to API service account: %v \n", err)
	}
	if !quiet {
		fmt.Printf("Done! Granted permissions to API service account\n")
	}
	// --- Grant Vertex AI, Firestore, and Storage permissions to Worker service account ---
	if !quiet {
		s.Suffix = " Granting permissions to Worker service account... "
		s.Start()
		defer s.Stop()
	}
	if err := grantPermissions(workerServiceAccount, projectID, quiet, bucketName); err != nil {
		log.Fatalf("Error granting permissions to Worker service account: %v\n", err)
	}
	if !quiet {
		fmt.Printf("Done! Granted permissions to Worker service account\n")
	}

	// --- Password, URL with Secret Manager ---
	var password, serviceURL string
	if !quiet {
		s.Suffix = " Getting or creating passwords... "
		s.Start()
		defer s.Stop()
	}
	// Get or create password and store it in Secret Manager
	password, err := utils.AccessSecret(projectID, "litmus-password")
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			password = utils.GenerateRandomPassword(16)
			if err := utils.CreateOrUpdateSecret(projectID, "litmus-password", password, quiet); err != nil {
				log.Fatalf("Error storing password in Secret Manager: %v", err)
			}
		} else {
			log.Fatalf("Error accessing password in Secret Manager: %v", err)
		}
	}
	envVars["PASSWORD"] = password

	// --- Deploy Cloud Run service with service account ---
	if !quiet {
		s.Suffix = " Deploying Cloud Run service 'litmus-api'... "
		s.Start()
		defer s.Stop()
	}

	apiImage := fmt.Sprintf("europe-docker.pkg.dev/litmusai-%s/litmus/api:latest", env)
	deployServiceCmd := exec.Command(
		"gcloud", "run", "deploy", "litmus-api",
		"--project", projectID,
		"--region", region,
		"--allow-unauthenticated",
		"--image", apiImage,
		"--service-account", apiServiceAccount,
	)

	for name, value := range envVars {
		deployServiceCmd.Args = append(deployServiceCmd.Args, "--set-env-vars", fmt.Sprintf("%s=%s", name, value))
	}

	deployServiceCmd.Args = append(deployServiceCmd.Args, "--set-env-vars", fmt.Sprintf("GCP_REGION=%s", region))
	deployServiceCmd.Args = append(deployServiceCmd.Args, "--set-env-vars", fmt.Sprintf("GCP_PROJECT=%s", projectID))
	deployServiceCmd.Args = append(deployServiceCmd.Args, "--set-env-vars", fmt.Sprintf("FILES_BUCKET=%s", bucketName))

	if utils.ServiceExists(projectID, region, "litmus-api") {
		deployServiceCmd.Args = append(deployServiceCmd.Args, "--no-traffic")
	}

	output, err := deployServiceCmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error deploying Cloud Run service: %v\nOutput: %s\n", err, output)
	}
	if !quiet {
		fmt.Println("Done! Deployed API.")
	}

	if strings.Contains(string(output), "Routing traffic...") {
		if !quiet {
			s.Suffix = " Routing traffic to the latest revision... "
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
			log.Fatalf("Error routing traffic to the latest revision: %v", err)
		}
		if !quiet {
			fmt.Println("Done! Routed traffic to the latest revision.")
		}
	}

	// --- Extract Service URL and Store in Secret Manager ---
	serviceURL = utils.ExtractServiceURL(string(output))
	if !quiet {
		s.Suffix = " Storing service URL... "
		s.Start()
		defer s.Stop()
	}
	if err := utils.CreateOrUpdateSecret(projectID, "litmus-service-url", serviceURL, quiet); err != nil {
		log.Fatalf("Error storing service URL in Secret Manager: %v", err)
	}

	// --- Deploy Cloud Run job with service account ---
	if !quiet {
		s.Suffix = " Deploying Cloud Run job 'litmus-worker'... "
		s.Start()
		defer s.Stop()
	}
	workerImage := fmt.Sprintf("europe-docker.pkg.dev/litmusai-%s/litmus/worker:latest", env)
	deployJobCmd := exec.Command(
		"gcloud", "run", "jobs", "deploy", "litmus-worker",
		"--project", projectID,
		"--region", region,
		"--image", workerImage,
		"--service-account", workerServiceAccount,
	)

	for name, value := range envVars {
		deployJobCmd.Args = append(deployJobCmd.Args, "--set-env-vars", fmt.Sprintf("%s=%s", name, value))
	}

	deployJobCmd.Args = append(deployJobCmd.Args, "--set-env-vars", fmt.Sprintf("GCP_REGION=%s", region))
	deployJobCmd.Args = append(deployJobCmd.Args, "--set-env-vars", fmt.Sprintf("GCP_PROJECT=%s", projectID))
	deployJobCmd.Args = append(deployJobCmd.Args, "--set-env-vars", fmt.Sprintf("FILES_BUCKET=%s", bucketName)) // Pass bucket name to Worker

	if utils.JobExists(projectID, region, "litmus-worker") {
		deployJobCmd.Args[3] = "update"
	}

	output, err = deployJobCmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error deploying Cloud Run job: %v\nOutput: %s", err, output) // Print gcloud output
	}
	if !quiet {
		fmt.Println("Done! Deployed Worker")
	}

	// --- Grant API permission to invoke Worker ---
	if !utils.BindingExists(projectID, region, "litmus-worker", apiServiceAccount, "roles/run.invoker") {
		if !quiet {
			s.Suffix = " Granting API permission to invoke Worker... "
			s.Start()
			defer s.Stop()
		}
		grantPermissionCmd := exec.Command(
			"gcloud", "run", "jobs", "add-iam-policy-binding", "litmus-worker",
			"--member", fmt.Sprintf("serviceAccount:%s", apiServiceAccount),
			"--role", "roles/run.invoker",
			"--project", projectID,
			"--region", region,
		)
		if err := grantPermissionCmd.Run(); err != nil {
			log.Fatalf("Error granting permission: %v\n", err)
		}
		if !quiet {
			fmt.Println("Done! Granting API permission to invoke Worker.\n")
		}
	} else if !quiet {
		fmt.Println("API permission to invoke Worker already exists.\n")
	}

	if !quiet {
		s.Suffix = " Setting up analytics... "
		s.Start()
		defer s.Stop()
	}
	// Deploy Analytics
	if err := analytics.DeployAnalytics(projectID, region, true); err != nil {
		utils.HandleGcloudError(err)
	}

	if !quiet {
		fmt.Println("\nAll deployments completed \n")
		fmt.Println("Get started now by visiting: ", serviceURL)
		fmt.Println("User: admin")
		fmt.Println("Password: ", password)
	}
}

// grantPermissions grants Vertex AI, Firestore, and Storage permissions to the given service account.
func grantPermissions(serviceAccount, projectID string, quiet bool, bucketName string) error {

	roles := []string{
		"roles/aiplatform.user",
		"roles/datastore.user",
		"roles/logging.logWriter",
		"roles/run.developer",
		"roles/bigquery.dataViewer",
		"roles/bigquery.jobUser",
	}

	for _, role := range roles {
		if !utils.BindingExists(projectID, "", "", serviceAccount, role) {
			cmd := exec.Command(
				"gcloud", "projects", "add-iam-policy-binding", projectID,
				"--member", fmt.Sprintf("serviceAccount:%s", serviceAccount),
				"--role", role,
				"--condition=None",
			)
			output, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("error granting role '%s': %v\nOutput: %s", role, err, output)
			}
		} else if !quiet {
			fmt.Printf("Role '%s' already granted to service account.\n", role)
		}
	}

	// Grant Storage Object Admin role on the bucket
	if !utils.BindingExists(projectID, "", bucketName, serviceAccount, "roles/storage.objectAdmin") {
		cmd := exec.Command(
			"gcloud", "storage", "buckets",
			"add-iam-policy-binding", fmt.Sprintf("gs://%s", bucketName),
			"--member", fmt.Sprintf("serviceAccount:%s", serviceAccount),
			"--role", "roles/storage.objectAdmin",
		)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("error granting Storage Object Admin role: %w\nOutput: %s", err, output)
		}
	} else if !quiet {
		fmt.Printf("Storage Object Admin role already granted to service account on bucket '%s'.\n", bucketName)
	}

	return nil
}

func createFilesBucket(bucketName, region, projectID string, quiet bool) error {
	// Check if the bucket already exists using gcloud
	cmd := exec.Command(
		"gcloud", "storage", "buckets", "describe",
		fmt.Sprintf("gs://%s", bucketName),
		"--project", projectID,
	)
	output, err := cmd.CombinedOutput()

	if err != nil {
		// Check if the error is specifically because the bucket doesn't exist
		if strings.Contains(string(output), "not found") {
			// Bucket does not exist, create it
			cmd = exec.Command(
				"gcloud", "storage", "buckets", "create",
				fmt.Sprintf("gs://%s", bucketName),
				"--location", region,
				"--project", projectID,
			)
			output, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("error creating files bucket: %w\nOutput: %s", err, output)
			}
			if !quiet {
				fmt.Printf("Created files bucket: gs://%s\n", bucketName)
			}
		} else {
			return fmt.Errorf("error describing bucket (it might exist, but there could be other issues): %w\nOutput: %s", err, output)
		}
	} else if !quiet {
		fmt.Printf("Files bucket '%s' already exists, skipping creation.\n", bucketName)
	}

	return nil
}
