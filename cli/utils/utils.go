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

package utils

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

// GenerateRandomPassword generates a random password of the given length.
func GenerateRandomPassword(length int) string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()")
	var password []rune
	for i := 0; i < length; i++ {
		password = append(password, chars[rand.Intn(len(chars))])
	}
	return string(password)
}

// AccessSecret retrieves a secret from Secret Manager.
func AccessSecret(projectID, secretID string) (string, error) {
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create secretmanager client: %v", err)
	}
	defer client.Close()

	name := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", projectID, secretID)

	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: name,
	}
	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to access secret: %v", err)
	}

	return string(result.Payload.Data), nil
}

// CreateOrUpdateSecret creates or updates a secret in Secret Manager.
func CreateOrUpdateSecret(projectID, secretID, secretValue string, quiet bool) error {
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create secretmanager client: %v", err)
	}
	defer client.Close()

	secretName := fmt.Sprintf("projects/%s/secrets/%s", projectID, secretID)
	_, err = client.GetSecret(ctx, &secretmanagerpb.GetSecretRequest{
		Name: secretName,
	})

	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			if !quiet {
				fmt.Printf("Creating secret %s", secretID)
			}
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

// IsAPIEnabled checks if a given API is enabled for the project.
func IsAPIEnabled(api, projectID string) bool {
	checkCmd := exec.Command("gcloud", "services", "list", "--project", projectID, "--enabled")
	output, err := checkCmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error checking API status: %v\nOutput: %s", err, output)
	}
	return strings.Contains(string(output), api)
}

// FirestoreDatabaseExists checks if the default Firestore database exists for the project.
func FirestoreDatabaseExists(projectID string) bool {
	listFirestoreCmd := exec.Command("gcloud", "firestore", "databases", "list", "--project", projectID)
	output, err := listFirestoreCmd.CombinedOutput()
	if err != nil {
		log.Fatalf("\nError listing Firestore databases: %v\nOutput: %s", err, output)
	}

	return strings.Contains(string(output), "(default)")
}

// RemoveAnsiEscapeSequences removes ANSI escape sequences from a string.
func RemoveAnsiEscapeSequences(text string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return re.ReplaceAllString(text, "")
}

// ServiceAccountExists checks if a service account already exists.
func ServiceAccountExists(projectID, serviceAccount string) bool {
	cmd := exec.Command("gcloud", "iam", "service-accounts", "list",
		"--project", projectID,
		"--filter", fmt.Sprintf("email=%s", serviceAccount),
		"--format=value(email)")
	output, _ := cmd.CombinedOutput() // Ignore errors here, as we're just checking existence
	return strings.TrimSpace(string(output)) == serviceAccount
}

// ServiceExists checks if a Cloud Run service already exists.
func ServiceExists(projectID, region, serviceName string) bool {
	cmd := exec.Command("gcloud", "run", "services", "list",
		"--project", projectID,
		"--region", region,
		"--filter", fmt.Sprintf("name=%s", serviceName),
		"--format=value(name)")
	output, _ := cmd.CombinedOutput()
	return strings.TrimSpace(string(output)) == serviceName
}

// JobExists checks if a Cloud Run job already exists.
func JobExists(projectID, region, jobName string) bool {
	cmd := exec.Command("gcloud", "run", "jobs", "list",
		"--project", projectID,
		"--region", region,
		"--filter", fmt.Sprintf("name=%s", jobName),
		"--format=value(name)")
	output, _ := cmd.CombinedOutput()
	return strings.TrimSpace(string(output)) == jobName
}

// BindingExists checks if a specific IAM binding already exists.
func BindingExists(projectID, region, resourceName, serviceAccount, role string) bool {
	var cmd *exec.Cmd
	if resourceName != "" {
		if region != "" {
			cmd = exec.Command("gcloud", "run", "jobs", "describe", resourceName,
				"--project", projectID,
				"--region", region,
				"--format=json",
			)
		} else {
			cmd = exec.Command("gcloud", "projects", "get-iam-policy", projectID, "--format=json")
		}
	} else {
		return false
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error checking IAM bindings: %v\nOutput: %s", err, output)
		return false
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

// ExtractServiceURL extracts the service URL from the gcloud command output.
func ExtractServiceURL(output string) string {
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

// GetDefaultProjectID retrieves the default project ID from gcloud.
func GetDefaultProjectID() (string, error) {
	cmd := exec.Command("gcloud", "config", "get-value", "core/project")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	// Trim any extra output before the project ID
	projectID := strings.TrimSpace(string(output))
	lines := strings.Split(projectID, "\n")
	if len(lines) > 1 {
		projectID = lines[len(lines)-1] // Take the last line
	}

	projectID = strings.TrimSpace(projectID) // Trim whitespace again

	return projectID, nil
}

// HandleGcloudError provides user-friendly messages for gcloud errors.
func HandleGcloudError(err error) {
	if strings.Contains(err.Error(), "executable file not found") ||
		strings.Contains(err.Error(), "Credential file cannot be found") {
		fmt.Println("Error using gcloud. Please make sure you have the Google Cloud SDK installed and authenticated.")
		fmt.Println("Run 'gcloud --version' to check if the SDK is installed.")
		fmt.Println("Run 'gcloud auth login' to authenticate.")
	} else {
		log.Fatalf("Error: %v", err)
	}
}

// Updated PrintUsage function
func PrintUsage() {
	fmt.Println("Usage: litmus <command> [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  deploy      Deploy the Litmus application")
	fmt.Println("  destroy     Destroy Litmus resources")
	fmt.Println("  tunnel      Create a tunnel to the Litmus UI")
	fmt.Println("  execute     Execute a payload against the Litmus application")
	fmt.Println("  ls          List Litmus runs")
	fmt.Println("  open        Open the Litmus dashboard")
	fmt.Println("  run         Open a specific Litmus run")
	fmt.Println("  start       Starts a new Litmus run")
	fmt.Println("  status      Show the status of the Litmus application")
	fmt.Println("  update      Update the Litmus application")
	fmt.Println("  version     Display the Litmus CLI version")
	fmt.Println("  analytics   Manage Litmus analytics (deploy or destroy)")
	fmt.Println("  proxy       Manage Litmus proxy (deploy, list, destroy, destroy-all)")
	fmt.Println("\nOptions:")
	fmt.Println("  --project <project_id>  Specify the Google Cloud project ID")
	fmt.Println("  --region <region>      Specify the Google Cloud region (default: us-central1)")
	fmt.Println("  --quiet                Suppress verbose output")
	fmt.Println("  --preserve-data        Preserve data in Cloud Storage, Firestore, and BigQuery")
	fmt.Println("\nExamples:")
	fmt.Println("  litmus deploy")
	fmt.Println("  litmus deploy --project my-project --region us-east1")
	fmt.Println("  litmus destroy --project my-project")
	fmt.Println("  litmus tunnel")
	fmt.Println("  litmus execute my-payload.json")
	fmt.Println("  litmus start my-template my-run")
	fmt.Println("  litmus ls")
	fmt.Println("  litmus open")
	fmt.Println("  litmus status")
	fmt.Println("  litmus analytics deploy")
	fmt.Println("  litmus proxy deploy --upstreamURL us-central1-aiplatform.googleapis.com")
	fmt.Println("  litmus proxy list")
	fmt.Println("  litmus proxy destroy us-west3-aiplatform-litmus-abcd")
	fmt.Println("  litmus proxy destroy-all")
}

// DisplayVersion prints the version of the Litmus CLI.
func DisplayVersion() {
	fmt.Println("Litmus CLI version:", "1.0.0") // Update with your actual version
}

// ConfirmPrompt asks the user for confirmation with a yes/no question.
func ConfirmPrompt(message string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s (y/N): ", message)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(response) // Remove leading/trailing whitespace
	return strings.ToLower(response) == "y"
}

// SelectUpstreamURL presents a list of upstream URLs to the user and lets them choose one.
func SelectUpstreamURL() (string, error) {
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
	for i, url := range upstreamURLs {
		fmt.Printf("%d. %s\n", i+1, url)
	}

	var choice int
	for {
		fmt.Print("Enter the number of your choice: ")
		_, err := fmt.Scanln(&choice)
		if err != nil {
			return "", fmt.Errorf("invalid input: %v", err)
		}

		if choice > 0 && choice <= len(upstreamURLs) {
			break
		}

		fmt.Println("Invalid choice. Please enter a number from the list.")
	}

	return upstreamURLs[choice-1], nil
}

// getAuthCredentials retrieves the basic authentication username and password from Secret Manager.
func GetAuthCredentials(projectID string) (string, string, error) {
	//username, err := AccessSecret(projectID, "litmus-username") // Replace with your secret name
	//if err != nil {
	//	return "", "", fmt.Errorf("error retrieving username from Secret Manager: %w", err)
	//}
	username := "admin"

	password, err := AccessSecret(projectID, "litmus-password") // Replace with your secret name
	if err != nil {
		return "", "", fmt.Errorf("error retrieving password from Secret Manager: %w", err)
	}

	return username, password, nil
}