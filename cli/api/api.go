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

package api

// RunInfo holds information about a Litmus run.
type RunInfo struct {
	EndTime   string `json:"end_time"`
	Progress  string `json:"progress"`
	RunID     string `json:"run_id"`
	StartTime string `json:"start_time"`
	Status    string `json:"status"`
	TemplateID string `json:"template_id"`
    URL       string `json:"url"` // Add the URL field
}

// Structs to represent the JSON response
type RunDetails struct {
	Progress            string        `json:"progress"`
	Status              string        `json:"status"`
	TemplateID         string        `json:"template_id"`
	TemplateInputField  string        `json:"template_input_field"`
	TemplateOutputField string        `json:"template_output_field"`
	TestCases           []TestCase `json:"testCases"`
}

type TestCase struct {
	GoldenResponse string   `json:"golden_response"`
	ID            string   `json:"id"`
	Request       Request `json:"request"`
	Response      Response `json:"response"`
	TracingID     string   `json:"tracing_id"`
}

type Request struct {
	Body    interface{} `json:"body"` // Can be more specific if needed
	Headers interface{} `json:"headers"` // Can be more specific if needed
	Method  string      `json:"method"`
	URL     string      `json:"url"`
}

type Response struct {
	Note     string     `json:"note"`
	Response ResponseData `json:"response"`
	Status   string     `json:"status"`
}
type ResponseData struct {
	Error  string `json:"error"`
	Status string `json:"status"`
}