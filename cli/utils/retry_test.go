// Copyright 2024 Google, LLC.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"os/exec"
	"testing"
	"time"
)

func shellCmd(script string) *exec.Cmd {
	return exec.Command("sh", "-c", script)
}

func TestRunWithRetryRetriesPropagationErrors(t *testing.T) {
	calls := 0
	out, err := RunWithRetry(func() *exec.Cmd {
		calls++
		if calls < 3 {
			return shellCmd(`echo "ERROR: Service account x@p.iam.gserviceaccount.com does not exist."; exit 1`)
		}
		return shellCmd(`echo ok`)
	}, 6, time.Millisecond, IsIAMPropagationError)
	if err != nil || calls != 3 || string(out) != "ok\n" {
		t.Fatalf("expected success on 3rd attempt, got calls=%d err=%v out=%q", calls, err, out)
	}
}

func TestRunWithRetryDoesNotRetryOtherErrors(t *testing.T) {
	calls := 0
	_, err := RunWithRetry(func() *exec.Cmd {
		calls++
		return shellCmd(`echo "ERROR: PERMISSION_DENIED"; exit 1`)
	}, 6, time.Millisecond, IsIAMPropagationError)
	if err == nil || calls != 1 {
		t.Fatalf("expected a single failing attempt, got calls=%d err=%v", calls, err)
	}
}

func TestRunWithRetryGivesUpAfterAttempts(t *testing.T) {
	calls := 0
	_, err := RunWithRetry(func() *exec.Cmd {
		calls++
		return shellCmd(`echo "does not exist"; exit 1`)
	}, 4, time.Millisecond, IsIAMPropagationError)
	if err == nil || calls != 4 {
		t.Fatalf("expected 4 attempts then an error, got calls=%d err=%v", calls, err)
	}
}

func TestRunWithRetryImmediateSuccess(t *testing.T) {
	calls := 0
	_, err := RunWithRetry(func() *exec.Cmd {
		calls++
		return shellCmd(`exit 0`)
	}, 6, time.Millisecond, IsIAMPropagationError)
	if err != nil || calls != 1 {
		t.Fatalf("expected immediate success, got calls=%d err=%v", calls, err)
	}
}
