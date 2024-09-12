package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/litmus/cli/utils"
)

// SubmitRun submits a Litmus run.
func SubmitRun(templateID, runID, projectID, authToken string) error {
	serviceURL, err := utils.AccessSecret(projectID, "litmus-service-url")
	if err != nil {
		log.Fatalf("Error retrieving service URL from Secret Manager: %v", err)
	}
	
	serviceURL = utils.RemoveAnsiEscapeSequences(serviceURL) 

	username, password, err := utils.GetAuthCredentials(projectID)
	if err != nil {
		return fmt.Errorf("error getting authentication credentials: %w", err)
	}

	url := fmt.Sprintf("%s/submit_run_simple", serviceURL)
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

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadJSON))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Set basic auth header
	req.SetBasicAuth(username, password)

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