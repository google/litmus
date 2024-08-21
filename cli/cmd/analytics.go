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
	"os/exec"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/google/litmus/cli/utils"
)

// Analytics represents the configuration for Litmus analytics.
type Analytics struct {
	ProjectID   string
	Region      string
	BucketName  string
	DatasetName string
}

// DeployAnalytics deploys Litmus analytics resources.
func DeployAnalytics(projectID, region string, quiet bool) error {
	if projectID == "" {
		var err error
		projectID, err = utils.GetDefaultProjectID()
		if err != nil {
			utils.HandleGcloudError(err)
			return err
		}
	}

	if region == "" {
		region = "us-central1" // Default region
	}

	analytics := Analytics{
		ProjectID:   projectID,
		Region:      region,
		BucketName:  fmt.Sprintf("%s-litmus-analytics", projectID),
		DatasetName: "litmus_analytics",
	}

	if !quiet {
		// --- Confirm deployment ---
		if !utils.ConfirmPrompt(fmt.Sprintf("\nThis will deploy Litmus analytics resources in project '%s' and region '%s'. Are you sure you want to continue?", analytics.ProjectID, analytics.Region)) {
			fmt.Println("\nAborting deployment.")
			return nil
		}
	}

	if !quiet {
		fmt.Println("\nDeploying Litmus Analytics...")
		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Start()
		defer s.Stop()
	}
	// --- Create logging bucket --- DISABLED FOR NOW
	//if err := createLoggingBucket(analytics, quiet); err != nil {
	//	return fmt.Errorf("error creating logging bucket: %w", err)
	//}

	// --- Create BigQuery dataset ---
	if err := createBigQueryDataset(analytics, quiet); err != nil {
		return fmt.Errorf("error creating BigQuery dataset: %w", err)
	}

	// --- Wait for BigQuery dataset to be created ---
	if err := waitForBigQueryDataset(analytics, quiet); err != nil {
		return fmt.Errorf("error waiting for BigQuery dataset creation: %w", err)
	}

	// --- Create log sink for proxy ---
	if err := createLogSink(analytics, quiet, "litmus-proxy-sink", "litmus-proxy-log"); err != nil {
		return fmt.Errorf("error creating log sink: %w", err)
	}

	// --- Create log sink for api ---
	if err := createLogSink(analytics, quiet, "litmus-core-sink", "litmus-core-log"); err != nil {
		return fmt.Errorf("error creating log sink: %w", err)
	}

	if !quiet {
		fmt.Println("Done! Deployed Litmus Analytics.")
	}
	return nil
}

// DeleteAnalytics deletes Litmus analytics resources.
func DeleteAnalytics(projectID, region string, quiet bool) error {
	if projectID == "" {
		var err error
		projectID, err = utils.GetDefaultProjectID()
		if err != nil {
			utils.HandleGcloudError(err)
			return err
		}
	}

	if region == "" {
		region = "us-central1" // Default region
	}

	analytics := Analytics{
		ProjectID:   projectID,
		Region:      region,
		BucketName:  fmt.Sprintf("%s-litmus-analytics", projectID),
		DatasetName: "litmus_analytics",
	}

	// --- Confirm deletion ---
	if !quiet {
		if !utils.ConfirmPrompt(fmt.Sprintf("\nThis will delete Litmus analytics resources in project '%s' and region '%s'. Are you sure you want to continue?", analytics.ProjectID, analytics.Region)) {
			fmt.Println("\nAborting deletion.")
			return nil
		}
	}

	if !quiet {
		fmt.Println("\nDeleting Litmus Analytics...")
		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Start()
		defer s.Stop()
	}

	// --- Delete log sink ---
	if err := deleteLogSink(analytics, quiet); err != nil {
		// Don't return an error here, as we still want to attempt
		// to delete the bucket and dataset even if the sink deletion fails.
		if !quiet {
			fmt.Printf("Error deleting log sink: %v\n", err)
		}
	}

	// --- Delete BigQuery dataset ---
	if err := deleteBigQueryDataset(analytics, quiet); err != nil {
		// Same as above - don't fail fast
		if !quiet {
			fmt.Printf("Error deleting BigQuery dataset: %v\n", err)
		}
	}

	// --- Delete logging bucket --- DISABLED FOR NOW
	// if err := deleteLoggingBucket(analytics, quiet); err != nil {
	// 	if !quiet {
	// 		fmt.Printf("Error deleting logging bucket: %v\n", err)
	// 	}
	// }
	if !quiet {
		fmt.Println("Done! Deleted Litmus Analytics.")
	}
	return nil
}

// func createLoggingBucket(a Analytics, quiet bool) error {
// 	// Check if bucket already exists
// 	cmd := exec.Command(
// 		"gsutil", "ls",
// 		fmt.Sprintf("gs://%s", a.BucketName),
// 	)
// 	_, err := cmd.CombinedOutput()
// 	if err == nil {
// 		if !quiet {
// 			fmt.Printf("Logging bucket 'gs://%s' already exists, skipping creation.\n", a.BucketName)
// 		}
// 		return nil
// 	}

// 	// Bucket doesn't exist, proceed with creation
// 	cmd = exec.Command(
// 		"gsutil", "mb",
// 		"-l", a.Region,
// 		"-p", a.ProjectID,
// 		fmt.Sprintf("gs://%s", a.BucketName),
// 	)
// 	output, err := cmd.CombinedOutput()
// 	if err != nil {
// 		return fmt.Errorf("error creating logging bucket: %w\nOutput: %s", err, output)
// 	}

// 	if !quiet {
// 		fmt.Printf("Created logging bucket: gs://%s\n", a.BucketName)
// 	}
// 	return nil
// }

func createBigQueryDataset(a Analytics, quiet bool) error {
	// Check if dataset already exists
	cmd := exec.Command(
		"gcloud", "alpha", "bq", "datasets", "describe",
		fmt.Sprintf("%s", a.DatasetName),
		"--project", a.ProjectID,
	)
	_, err := cmd.CombinedOutput()
	if err == nil {
		if !quiet {
			fmt.Printf("BigQuery dataset '%s:%s' already exists, skipping creation.\n", a.ProjectID, a.DatasetName)
		}
		return nil
	}

	// Dataset doesn't exist, proceed with creation
	cmd = exec.Command(
		"gcloud", "alpha", "bq", "datasets", "create",
		fmt.Sprintf("%s", a.DatasetName),
		"--project", a.ProjectID,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error creating BigQuery dataset: %w\nOutput: %s", err, output)
	}

	if !quiet {
		fmt.Printf("Created BigQuery dataset: %s:%s\n", a.ProjectID, a.DatasetName)
	}
	return nil
}

func waitForBigQueryDataset(a Analytics, quiet bool) error {
	if quiet {
		// If quiet mode, don't display the spinner
		return waitForBigQueryDatasetQuiet(a)
	}

	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Waiting for BigQuery dataset creation..."
	s.Start()
	defer s.Stop()

	return waitForBigQueryDatasetQuiet(a)
}

func waitForBigQueryDatasetQuiet(a Analytics) error {
	timeout := time.After(5 * time.Minute) // Set a timeout for dataset creation
	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for BigQuery dataset '%s' to be created", a.DatasetName)
		case <-ticker.C:
			cmd := exec.Command(
				"bq", "--project_id", a.ProjectID,
				"show",
				fmt.Sprintf("%s:%s", a.ProjectID, a.DatasetName),
			)
			_, err := cmd.CombinedOutput()
			if err == nil {
				return nil // Dataset exists
			}
		}
	}
}

func createLogSink(a Analytics, quiet bool, name string, filter string) error {
	// Check if log sink exists
	checkCmd := exec.Command( // Use a different variable name here
		"gcloud", "logging", "sinks", "describe", name,
		"--project", a.ProjectID,
	)
	_, err := checkCmd.CombinedOutput()

	// --- Create/Update Log Sink ---
	var cmd *exec.Cmd
	if err == nil {
		// Log sink exists, update it
		if !quiet {
			fmt.Println("Log sink 'litmus-proxy-sink' already exists, updating...")
		}

		cmd = exec.Command(
			"gcloud", "logging", "sinks", "update", name,
			fmt.Sprintf("bigquery.googleapis.com/projects/%s/datasets/%s", a.ProjectID, a.DatasetName),
			"--project", a.ProjectID,
			"--log-filter", "logName=projects/"+a.ProjectID+"/logs/"+filter,
		)

	} else {
		// Log sink doesn't exist, create it
		cmd = exec.Command(
			"gcloud", "logging", "sinks", "create", name,
			fmt.Sprintf("bigquery.googleapis.com/projects/%s/datasets/%s", a.ProjectID, a.DatasetName),
			"--project", a.ProjectID,
			"--log-filter", "logName=projects/"+a.ProjectID+"/logs/"+filter,
		)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error creating/updating log sink: %w\nOutput: %s", err, output)
	}

	// --- Grant BigQuery Data Editor Role ---
	if !quiet {
		fmt.Println("Granting BigQuery Data Editor role to logging service account...")
	}

	// Extract service account email from output
	serviceAccountEmail := extractServiceAccountEmail(string(output))
	if serviceAccountEmail == "" {
		return fmt.Errorf("unable to extract service account email from output: %s", output)
	}

	grantBigQueryDataEditorRole := exec.Command(
		"gcloud", "projects", "add-iam-policy-binding", a.ProjectID,
		"--member", fmt.Sprintf("serviceAccount:%s", serviceAccountEmail),
		"--role", "roles/bigquery.dataEditor",
	)

	if err := grantBigQueryDataEditorRole.Run(); err != nil {
		return fmt.Errorf("error granting BigQuery Data Editor role: %w", err)
	}
	if !quiet {
		fmt.Println("Created/Updated log sink: " + name)
	}
	return nil
}

// func deleteLoggingBucket(a Analytics, quiet bool) error {
// 	cmd := exec.Command(
// 		"gsutil", "-m", "rm", "-r",
// 		fmt.Sprintf("gs://%s", a.BucketName),
// 	)
// 	output, err := cmd.CombinedOutput()
// 	if err != nil && !strings.Contains(string(output), "BucketNotFoundException") {
// 		return fmt.Errorf("error deleting logging bucket: %w\nOutput: %s", err, output)
// 	}

// 	if !quiet {
// 		fmt.Printf("Deleted logging bucket: gs://%s\n", a.BucketName)
// 	}
// 	return nil
// }

func deleteBigQueryDataset(a Analytics, quiet bool) error {
	cmd := exec.Command(
		"gcloud", "alpha", "bq", "datasets", "delete",
		fmt.Sprintf("%s", a.DatasetName),
		"--project", a.ProjectID,
		"--recursive", // Use --recursive for recursive delete
		"--force",     // Use --force to force deletion
	)
	output, err := cmd.CombinedOutput()
	if err != nil && !strings.Contains(string(output), "NOT_FOUND") {
		return fmt.Errorf("error deleting BigQuery dataset: %w\nOutput: %s", err, output)
	}

	if !quiet {
		fmt.Printf("Deleted BigQuery dataset: %s:%s\n", a.ProjectID, a.DatasetName)
	}
	return nil
}

func deleteLogSink(a Analytics, quiet bool) error {
	cmd := exec.Command(
		"gcloud", "logging", "sinks", "delete", "litmus-proxy-sink",
		"--project", a.ProjectID,
		"--quiet", // Assume quiet for deletion unless specified otherwise
	)
	output, err := cmd.CombinedOutput()
	if err != nil && !strings.Contains(string(output), "NOT_FOUND") {
		return fmt.Errorf("error deleting log sink: %w\nOutput: %s", err, output)
	}

	if !quiet {
		fmt.Println("Deleted log sink: litmus-proxy-sink")
	}
	return nil
}

// Extracts the service account email from the gcloud output
func extractServiceAccountEmail(output string) string {
	start := strings.Index(output, "serviceAccount:")
	if start == -1 {
		return ""
	}
	start += len("serviceAccount:")
	end := strings.Index(output[start:], "`")
	if end == -1 {
		return ""
	}
	return output[start : start+end]
}
