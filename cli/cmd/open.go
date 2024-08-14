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
	"net/url"
	"os/exec"
	"runtime"

	"github.com/google/litmus/cli/utils"
)

// OpenLitmus opens the Litmus application in a browser,
// including the username and password in the URL.
func OpenLitmus(projectID string) {
	ShowStatus(projectID) // First, show the status so the user knows the credentials

	serviceURL, _ := utils.AccessSecret(projectID, "litmus-service-url")
	username := "admin"
	password, _ := utils.AccessSecret(projectID, "litmus-password")

	noAServiceURL := utils.RemoveAnsiEscapeSequences(serviceURL)

	parsedURL, err := url.Parse(noAServiceURL)
	if err != nil {
		panic(err)
	}

	parsedURL.User = url.UserPassword(username, password)

	finalURL := parsedURL.String()
	openBrowser(finalURL)
}

// openBrowser opens the specified URL in the default browser.
func openBrowser(url string) {
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