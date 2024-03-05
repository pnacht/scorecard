// Copyright 2024 OpenSSF Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package patch

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/ossf/scorecard/v4/checker"
)

func Test_GeneratePatch(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input_filepath    string
		expected_filepath string
		// err      error
	}{
		// Extracted from real Angular fix: https://github.com/angular/angular/pull/51026/files
		{
			name: "Real Example 1",
			input_filepath: "realExample1.yaml",
			expected_filepath: "realExample1_fixed.yaml",
		},
		// Inspired on a real fix: https://github.com/googleapis/google-cloud-go/pull/9011/files
		{
			name: "Real Example 2",
			input_filepath: "realExample2.yaml",
			expected_filepath: "realExample2_fixed.yaml",
		},
		// Inspired from a real lit/lit fix: https://github.com/lit/lit/pull/3669/files
		{
			name: "Real Example 3",
			input_filepath: "realExample3.yaml",
			expected_filepath: "realExample3_fixed.yaml",
		},
		{
			name: "Test all (or most) types of user input that should be detected",
			input_filepath: "allKindsOfUserInput.yaml",
			expected_filepath: "allKindsOfUserInput_fixed.yaml",
		},
		{
			name: "User's input is assigned to a variable before used",
			input_filepath: "userInputAssignedToVariable.yaml",
			expected_filepath: "userInputAssignedToVariable_fixed.yaml",
		},
		{
			name: "Two incidences in different jobs",
			input_filepath: "twoInjectionsDifferentJobs.yaml",
			expected_filepath: "twoInjectionsDifferentJobs_fixed.yaml",
		},
		{
			name: "Two incidences in same job",
			input_filepath: "twoInjectionsSameJob.yaml",
			expected_filepath: "twoInjectionsSameJob_fixed.yaml",
		},
		{
			name: "Two incidences in same step",
			input_filepath: "twoInjectionsSameStep.yaml",
			expected_filepath: "twoInjectionsSameStep_fixed.yaml",
		},
		{
			name: "Bad indentation is kept the same",
			input_filepath: "badIndentationMultipleInjections.yaml",
			expected_filepath: "badIndentationMultipleInjections_fixed.yaml",
		},
		{
			// Currently we're not keeping this pattern, as we always add a blankline after the env block
			name: "File with no blank lines between blocks",
			input_filepath: "noLineBreaksBetweenBlocks.yaml",
			expected_filepath: "noLineBreaksBetweenBlocks_fixed.yaml",
		},
		{
			name: "Ignore if user input regex is just part of a comment",
			input_filepath: "safeExample.yaml",
			expected_filepath: "safeExample.yaml",
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			input_file := checker.File{
				Path: tt.input_filepath,
			}

			expected_content, err := os.ReadFile("./testdata/" + tt.expected_filepath)
			if err != nil {
				t.Errorf("Couldn't read expected testfile. Error:\n%s", err)
			}

			output := GeneratePatch(input_file)
			if diff := cmp.Diff(string(expected_content[:]), output); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}