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
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os/exec"
	"regexp"
	"strings"
	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

// ShowInProgress displays an in-progress animation until the command finishes.
func ShowInProgress(cmd *exec.Cmd) {
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
func CreateOrUpdateSecret(projectID, secretID, secretValue string) error {
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
	projectID := strings.TrimSpace(string(output))
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

// PrintUsage displays the help message for the Litmus CLI.
func PrintUsage() {
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
}

// DisplayVersion prints the version of the Litmus CLI.
func DisplayVersion() {
	fmt.Println("Litmus CLI version:", "1.0.0") // Update with your actual version
}
