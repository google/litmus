package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/google/litmus/cli/utils"
)

// SubmitRun submits a Litmus run.
func SubmitRun(templateID, runID, projectID, authToken string) error {
	serviceURL, err := utils.AccessSecret(projectID, "litmus-service-url")
	if err != nil {
		log.Fatalf("Error retrieving service URL from Secret Manager: %v", err)
	}
	
	// Remove ANSI escape codes using a regular expression
	re := regexp.MustCompile(`\x1b\[[0-9;]*[mG]`)
	serviceURL = re.ReplaceAllString(serviceURL, "") 
    if err != nil {
        log.Fatalf("Error retrieving service URL from Secret Manager: %v", err)
    }
    // Construct the URL for your submit_run_simple endpoint.
    url := fmt.Sprintf("%s/submit_run_simple", serviceURL)

    // Create the JSON payload
    payload := map[string]interface{}{
        "run_id":      runID,
        "template_id": templateID,
    }

    // Add authToken to payload only if it's set
    if authToken != "" {
        payload["auth_token"] = authToken 
    }

    payloadJSON, err := json.Marshal(payload)
    if err != nil {
        return fmt.Errorf("error marshaling JSON payload: %w", err)
    }

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second, // Set a timeout for the request
	}

	// Make the HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadJSON))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	// Handle the response (check for success/errors)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %s, response: %s", resp.Status, string(body))
	}

	// Handle successful response (You might want to process the response here)
	//fmt.Println("Run submitted successfully.")

	return nil
}