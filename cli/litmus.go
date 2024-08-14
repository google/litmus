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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	secretmanagerpb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
)

// Struct to hold run information
type RunInfo struct {
	RunID string `json:"runID"`
	URL   string `json:"url"`
}

func main() {
	// Get default project ID
	projectID, err := getDefaultProjectID()
	if err != nil {
		handleGcloudError(err)
		return
	}

	// Get command and parameters
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run litmus.go <command> [options] [flags]")
		fmt.Println("Commands:")
		fmt.Println("  open: Open the Web application")
		fmt.Println("  deploy: Deploy the application")
		fmt.Println("  destroy: Remove the application")
		fmt.Println("  status: Show the status of the Litmus deployment")
		fmt.Println("  version: Display the version of the Litmus CLI")
		fmt.Println("  execute: Execute a payload to the deployed endpoint")
		fmt.Println("  ls: List all runs")
		fmt.Println("  run <runID>: Open the URL for a certain runID")
		fmt.Println("Options:")
		fmt.Println("  --project <project-id>: Specify the project ID (overrides default)")
		fmt.Println("  --region <region>: Specify the region (defaults to 'us-central1')")
		return
	}

	command := os.Args[1]
	region := "us-central1" // Default region
	var runID string

	// Parse command-line arguments
	for i := 2; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--project":
			if i+1 < len(os.Args) {
				projectID = os.Args[i+1]
				i++
			} else {
				fmt.Println("Error: --project flag requires an argument")
				return
			}
		case "--region":
			if i+1 < len(os.Args) {
				region = os.Args[i+1]
				i++
			} else {
				fmt.Println("Error: --region flag requires an argument")
				return
			}
		case "open":
			if i+1 < len(os.Args) {
				runID = os.Args[i+1]
				i++
			} else {
				fmt.Println("Error: 'open' command requires a runID argument")
				return
			}
		}
	}

	// Extract environment variables from command-line arguments
	envVars := make(map[string]string)
	for i := 3; i < len(os.Args); i++ {
		parts := strings.Split(os.Args[i], "=")
		if len(parts) == 2 {
			envVars[parts[0]] = parts[1]
		}
	}

	switch command {
	case "deploy":
		deployApplication(projectID, region, envVars)
	case "destroy":
		destroyResources(projectID, region)
	case "execute":
		if len(os.Args) < 4 {
			fmt.Println("Usage: go run main.go execute <payload>")
			return
		}
		payload := os.Args[3]
		executePayload(projectID, payload)
	case "ls":
		listRuns(projectID)
	case "open":
		openLitmus(projectID)
	case "run":
		if runID == "" {
			fmt.Println("Error: 'run' command requires a runID argument")
			return
		}
		openRun(projectID, runID)
	case "status":
		showStatus(projectID)
	case "version":
		displayVersion()
	default:
		fmt.Println("Invalid command:", command)
	}
}

// --- Authentication and Error Handling ---

// handleGcloudError provides user-friendly messages for gcloud errors
func handleGcloudError(err error) {
	if strings.Contains(err.Error(), "executable file not found") ||
		strings.Contains(err.Error(), "Credential file cannot be found") {
		fmt.Println("Error using gcloud. Please make sure you have the Google Cloud SDK installed and authenticated.")
		fmt.Println("Run 'gcloud --version' to check if the SDK is installed.")
		fmt.Println("Run 'gcloud auth login' to authenticate.")
	} else {
		log.Fatalf("Error: %v", err)
	}
}

// getDefaultProjectID retrieves the default project ID from gcloud
func getDefaultProjectID() (string, error) {
	cmd := exec.Command("gcloud", "config", "get-value", "core/project")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	projectID := strings.TrimSpace(string(output))
	return projectID, nil
}

// --- API Enabling and Service Account Management ---

func deployApplication(projectID, region string, envVars map[string]string) {

	fmt.Println("Deploying to project:", projectID) // Print the project ID
	// Enable required APIs
	apisToEnable := []string{
		"run.googleapis.com",
		"firestore.googleapis.com",
		"iam.googleapis.com",           // Add IAM API for service account management
		"aiplatform.googleapis.com",    // Enable Vertex AI API
		"secretmanager.googleapis.com", // Enable Secret Manager
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
		log.Fatalf("\nError listing Firestore databases: %v\nOutput: %s", err, output)
	}

	if strings.Contains(string(output), "(default)") { // Check for "(default)" in output
		fmt.Println("\nFirestore database already exists.")
	} else {
		// Create default Firestore database
		fmt.Print("\nCreating default Firestore database ")
		createFirestoreCmd := exec.Command(
			"gcloud", "firestore", "databases", "create",
			"--project", projectID,
			"--location", region,
		)
		go showInProgress(createFirestoreCmd)
		if err := createFirestoreCmd.Run(); err != nil {
			log.Fatalf("\nError creating Firestore database: %v", err)
		}
		fmt.Println("Done!")
	}

	// --- Service Account for API ---
	apiServiceAccount := fmt.Sprintf("%s-api@%s.iam.gserviceaccount.com", projectID, projectID)
	if !serviceAccountExists(projectID, apiServiceAccount) {
		fmt.Printf("Creating service account for API: %s ", apiServiceAccount)
		createServiceAccountCmd := exec.Command(
			"gcloud", "iam", "service-accounts", "create",
			fmt.Sprintf("%s-api", projectID), // Service account name (without @...)
			"--project", projectID,
			"--display-name", "Litmus API Service Account",
		)
		go showInProgress(createServiceAccountCmd)
		if err := createServiceAccountCmd.Run(); err != nil {
			log.Fatalf("\nError creating service account: %v", err)
		}
		fmt.Println("Done!")
	} else {
		fmt.Printf("\nService account for API already exists: %s\n", apiServiceAccount)
	}

	// --- Service Account for Worker ---
	workerServiceAccount := fmt.Sprintf("%s-worker@%s.iam.gserviceaccount.com", projectID, projectID)
	if !serviceAccountExists(projectID, workerServiceAccount) {
		fmt.Printf("\nCreating service account for Worker: %s ", workerServiceAccount)
		createWorkerServiceAccountCmd := exec.Command(
			"gcloud", "iam", "service-accounts", "create",
			fmt.Sprintf("%s-worker", projectID), // Service account name (without @...)
			"--project", projectID,
			"--display-name", "Litmus Worker Service Account",
		)
		go showInProgress(createWorkerServiceAccountCmd)
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
	password, err := accessSecret(projectID, "litmus-password")
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			// Generate and store password if it doesn't exist
			password = generateRandomPassword(16)
			fmt.Printf("Generated random password: %s\n", password)
			if err := createOrUpdateSecret(projectID, "litmus-password", password); err != nil {
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
	if serviceExists(projectID, region, "litmus-api") {
		deployServiceCmd.Args = append(deployServiceCmd.Args, "--no-traffic")
	}

	go showInProgress(deployServiceCmd)
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
		go showInProgress(routeTrafficCmd)
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
	if jobExists(projectID, region, "litmus-worker") {
		deployJobCmd.Args[3] = "update" // Change "create" to "update"
	}

	go showInProgress(deployJobCmd)
	if err := deployJobCmd.Run(); err != nil {
		log.Fatalf("\nError deploying Cloud Run job: %v", err)
	}
	fmt.Println("Done!")

	// --- Grant API permission to invoke Worker ---
	if !bindingExists(projectID, region, "litmus-worker", apiServiceAccount, "roles/run.invoker") {
		fmt.Print("\nGranting API permission to invoke Worker... ")
		grantPermissionCmd := exec.Command(
			"gcloud", "run", "jobs", "add-iam-policy-binding", "litmus-worker",
			"--member", fmt.Sprintf("serviceAccount:%s", apiServiceAccount),
			"--role", "roles/run.invoker",
			"--project", projectID,
			"--region", region,
		)
		go showInProgress(grantPermissionCmd)
		if err := grantPermissionCmd.Run(); err != nil {
			log.Fatalf("Error granting permission: %v", err)
		}
		fmt.Println("\nDone!")
	} else {
		fmt.Println("\nAPI permission to invoke Worker already exists.")
	}

	// Extract and print the service URL
	serviceURL := extractServiceURL(string(output2))

	// Store the service URL in Secret Manager
	if err := createOrUpdateSecret(projectID, "litmus-service-url", serviceURL); err != nil {
		log.Fatalf("\nError storing service URL in Secret Manager: %v", err)
	}

	fmt.Println("\nAll deployments completed \n")

	fmt.Println("Get started now by visiting: ", serviceURL)
	fmt.Println("User: admin")
	fmt.Println("Password: ", password)
}

// destroyResources removes all resources created by this script
func destroyResources(projectID, region string) {

	fmt.Println("Destroying resources in project:", projectID) // Print the project ID

	// --- Delete Cloud Run service ---
	fmt.Print("\nDeleting Cloud Run service 'litmus-api'... ")
	deleteServiceCmd := exec.Command("gcloud", "run", "services", "delete", "litmus-api",
		"--project", projectID,
		"--region", region,
		"--quiet", // Use --quiet to suppress confirmation prompt
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
		"--quiet", // Use --quiet to suppress confirmation prompt
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

// showStatus displays the status of the Litmus deployment
func showStatus(projectID string) {
	serviceURL, err := accessSecret(projectID, "litmus-service-url")
	if err != nil {
		fmt.Println("Litmus is not deployed or there was an error retrieving the status.")
		return
	}

	password, err := accessSecret(projectID, "litmus-password")
	if err != nil {
		fmt.Println("Error retrieving password from Secret Manager:", err)
		return
	}

	fmt.Println("Litmus Deployment Status:")
	fmt.Println("URL:", serviceURL)
	fmt.Println("User: admin")
	fmt.Println("Password:", password)
}

func openbrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}

}

func removeAnsiEscapeSequences(text string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return re.ReplaceAllString(text, "")
}

// openLitmus opens the URL associated with a specific runID in the browser
// and includes the username and password in the URL.
func openLitmus(projectID string) {
	showStatus(projectID) // First, show the status so the user knows the credentials

	serviceURL, _ := accessSecret(projectID, "litmus-service-url")
	username := "admin"
	password, _ := accessSecret(projectID, "litmus-password")

	noAserviceURL := removeAnsiEscapeSequences(serviceURL)
	// Parse the URL
	parsedURL, err := url.Parse(noAserviceURL)
	if err != nil {
		panic(err)
	}

	// Set the username and password in the URL
	parsedURL.User = url.UserPassword(username, password)

	// Construct the final URL with credentials
	finalURL := parsedURL.String()
	openbrowser(finalURL)
}

// displayVersion prints the version of the Litmus CLI
func displayVersion() {
	fmt.Println("Litmus CLI version:", "1.0.0") // Update with your actual version
}

func executePayload(projectID, payload string) {
	// Get the service URL from the deployed service
	getServiceURLCmd := exec.Command("gcloud", "run", "services", "describe", "litmus-api", "--project", projectID, "--region=us-central1", "--format=value(status.url)")
	var out bytes.Buffer
	getServiceURLCmd.Stdout = &out
	if err := getServiceURLCmd.Run(); err != nil {
		log.Fatalf("Error getting service URL: %v", err)
	}
	serviceURL := strings.TrimSpace(out.String())

	// Send the payload to the endpoint
	requestBody, err := json.Marshal(map[string]string{
		"message": payload,
	})
	if err != nil {
		log.Fatalf("Error marshaling JSON: %v", err)
	}

	resp, err := http.Post(serviceURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Print the response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}
	fmt.Println("Response:", string(responseBody))
}

// grantPermissions grants Vertex AI and Firestore permissions to the given service account
func grantPermissions(serviceAccount, projectID string) error {
	roles := []string{
		"roles/aiplatform.user",   // Vertex AI access
		"roles/datastore.user",    // Firestore access
		"roles/logging.logWriter", // Logging
		"roles/run.developer",     // Run Invoker
	}

	for _, role := range roles {
		if !bindingExists(projectID, "", "", serviceAccount, role) { // No region needed for project-level bindings
			cmd := exec.Command(
				"gcloud", "projects", "add-iam-policy-binding", projectID,
				"--member", fmt.Sprintf("serviceAccount:%s", serviceAccount),
				"--role", role,
			)
			go showInProgress(cmd)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("error granting role '%s': %v", role, err)
			}
		} else {
			fmt.Printf("Role '%s' already granted to service account.\n", role)
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

// generateRandomPassword generates a random password of the given length
func generateRandomPassword(length int) string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()")
	var password []rune
	for i := 0; i < length; i++ {
		password = append(password, chars[rand.Intn(len(chars))])
	}
	return string(password)
}

// --- Secret Manager Functions ---

// accessSecret retrieves a secret from Secret Manager
func accessSecret(projectID, secretID string) (string, error) {
	// Create a client.
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create secretmanager client: %v", err)
	}
	defer client.Close()

	// Build the resource name of the secret.
	name := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", projectID, secretID)

	// Access the secret.
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: name,
	}
	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to access secret: %v", err)
	}

	return string(result.Payload.Data), nil
}

// createOrUpdateSecret creates or updates a secret in Secret Manager
func createOrUpdateSecret(projectID, secretID, secretValue string) error {
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create secretmanager client: %v", err)
	}
	defer client.Close()

	// Check if the secret already exists
	secretName := fmt.Sprintf("projects/%s/secrets/%s", projectID, secretID)
	_, err = client.GetSecret(ctx, &secretmanagerpb.GetSecretRequest{
		Name: secretName,
	})

	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			// Create a new secret
			fmt.Printf("Creating secret %s\n", secretID)
			createSecretReq := &secretmanagerpb.CreateSecretRequest{
				Parent:   fmt.Sprintf("projects/%s", projectID),
				SecretId: secretID,
				Secret: &secretmanagerpb.Secret{
					Replication: &secretmanagerpb.Replication{
						Replication: &secretmanagerpb.Replication_Automatic_{
							Automatic: &secretmanagerpb.Replication_Automatic{},
						},
					},
				},
			}
			_, err = client.CreateSecret(ctx, createSecretReq)
			if err != nil {
				return fmt.Errorf("failed to create secret: %v", err)
			}
		} else {
			return fmt.Errorf("failed to get secret: %v", err)
		}
	}

	// Add a new version to the secret
	addSecretVersionReq := &secretmanagerpb.AddSecretVersionRequest{
		Parent: secretName,
		Payload: &secretmanagerpb.SecretPayload{
			Data: []byte(secretValue),
		},
	}
	_, err = client.AddSecretVersion(ctx, addSecretVersionReq)
	if err != nil {
		return fmt.Errorf("failed to add secret version: %v", err)
	}

	return nil
}

// --- Run Management Functions ---

// listRuns retrieves and displays a list of runs from the service
func listRuns(projectID string) {
	// Get the service URL from Secret Manager
	serviceURL, err := accessSecret(projectID, "litmus-service-url")
	if err != nil {
		log.Fatalf("Error retrieving service URL from Secret Manager: %v", err)
	}

	// Send request to list runs
	resp, err := http.Get(serviceURL + "/runs")
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Decode the response
	var runs []RunInfo
	if err := json.NewDecoder(resp.Body).Decode(&runs); err != nil {
		log.Fatalf("Error decoding response: %v", err)
	}

	// Display the runs
	if len(runs) == 0 {
		fmt.Println("No runs found.")
	} else {
		fmt.Println("Runs:")
		for _, run := range runs {
			fmt.Printf("Run ID: %s, URL: %s\n", run.RunID, run.URL)
		}
	}
}

// openRun opens the URL associated with a specific runID in the browser
func openRun(projectID, runID string) {
	// Get the service URL from Secret Manager
	serviceURL, err := accessSecret(projectID, "litmus-service-url")
	if err != nil {
		log.Fatalf("Error retrieving service URL from Secret Manager: %v", err)
	}

	// Construct the run URL
	runURL := fmt.Sprintf("%s/runs/%s", serviceURL, runID)

	// Open the URL in the default browser
	if err := exec.Command("open", runURL).Start(); err != nil {
		log.Fatalf("Error opening URL: %v", err)
	}
}

// serviceAccountExists checks if a service account already exists
func serviceAccountExists(projectID, serviceAccount string) bool {
	cmd := exec.Command("gcloud", "iam", "service-accounts", "list",
		"--project", projectID,
		"--filter", fmt.Sprintf("email=%s", serviceAccount),
		"--format=value(email)")
	output, _ := cmd.CombinedOutput() // Ignore errors here, as we're just checking existence
	return strings.TrimSpace(string(output)) == serviceAccount
}

// serviceExists checks if a Cloud Run service already exists
func serviceExists(projectID, region, serviceName string) bool {
	cmd := exec.Command("gcloud", "run", "services", "list",
		"--project", projectID,
		"--region", region,
		"--filter", fmt.Sprintf("name=%s", serviceName),
		"--format=value(name)")
	output, _ := cmd.CombinedOutput()
	return strings.TrimSpace(string(output)) == serviceName
}

// jobExists checks if a Cloud Run job already exists
func jobExists(projectID, region, jobName string) bool {
	cmd := exec.Command("gcloud", "run", "jobs", "list",
		"--project", projectID,
		"--region", region,
		"--filter", fmt.Sprintf("name=%s", jobName),
		"--format=value(name)")
	output, _ := cmd.CombinedOutput()
	return strings.TrimSpace(string(output)) == jobName
}

// bindingExists checks if a specific IAM binding already exists
func bindingExists(projectID, region, resourceName, serviceAccount, role string) bool {
	var cmd *exec.Cmd
	if resourceName != "" { // If resourceName is provided, it's a resource-specific binding
		if region != "" { // Region is required for Cloud Run resources
			cmd = exec.Command("gcloud", "run", "jobs", "describe", resourceName,
				"--project", projectID,
				"--region", region,
				"--format=json",
			)
		} else {
			// If no region, assume project-level binding
			cmd = exec.Command("gcloud", "projects", "get-iam-policy", projectID, "--format=json")
		}
	} else {
		return false // If no resourceName, we can't check
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error checking IAM bindings: %v\nOutput: %s", err, output)
		return false // Assume binding doesn't exist on error
	}

	var data map[string]interface{}
	if err := json.Unmarshal(output, &data); err != nil {
		log.Printf("Error parsing JSON output: %v", err)
		return false
	}

	bindings, ok := data["bindings"].([]interface{})
	if !ok {
		return false
	}

	for _, b := range bindings {
		binding, ok := b.(map[string]interface{})
		if !ok {
			continue
		}

		if binding["role"] == role {
			members, ok := binding["members"].([]interface{})
			if !ok {
				continue
			}

			for _, m := range members {
				member, ok := m.(string)
				if ok && member == fmt.Sprintf("serviceAccount:%s", serviceAccount) {
					return true
				}
			}
		}
	}

	return false
}
