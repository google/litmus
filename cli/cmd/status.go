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

	"github.com/google/litmus/cli/utils"
)

// ShowStatus displays the status of the Litmus deployment.
func ShowStatus(projectID string) {
	serviceURL, err := utils.AccessSecret(projectID, "litmus-service-url")
	if err != nil {
		fmt.Println("Litmus is not deployed or there was an error retrieving the status.")
		return
	}

	password, err := utils.AccessSecret(projectID, "litmus-password")
	if err != nil {
		fmt.Println("Error retrieving password from Secret Manager:", err)
		return
	}

	fmt.Println("Litmus Deployment Status:")
	fmt.Println("URL:", serviceURL)
	fmt.Println("User: admin")
	fmt.Println("Password:", password)
}