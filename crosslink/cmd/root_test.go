// Copyright The OpenTelemetry Authors
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
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	cl "go.opentelemetry.io/build-tools/crosslink/internal"
)

func TestTransform(t *testing.T) {
	tests := []struct {
		testName   string
		inputSlice []string
	}{
		{
			testName: "with items",
			inputSlice: []string{
				"example.com/testA",
				"example.com/testB",
				"example.com/testC",
				"example.com/testD",
				"example.com/testE",
			},
		},
		{
			testName:   "with empty",
			inputSlice: []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			actual := transformExclude(test.inputSlice)

			//len must match
			assert.Len(t, actual, len(test.inputSlice))

			//test for existence
			for _, val := range test.inputSlice {
				_, exists := actual[val]
				assert.True(t, exists)
			}
		})
	}
}

// Validate run config is valid after pre run.
func TestPreRun(t *testing.T) {
	configReset := func() {
		rc = cl.DefaultRunConfig()
		rootCmd.SetArgs([]string{})
	}

	tests := []struct {
		testName       string
		args           []string
		mockConfig     cl.RunConfig
		expectedConfig cl.RunConfig
	}{
		{
			testName:       "Default Config",
			args:           []string{},
			mockConfig:     cl.DefaultRunConfig(),
			expectedConfig: cl.DefaultRunConfig(),
		},
		{
			testName: "with overwrite",
			mockConfig: cl.RunConfig{
				Overwrite: true,
			},
			expectedConfig: cl.RunConfig{
				Overwrite: true,
				Verbose:   true,
			},
			args: []string{"--overwrite"},
		},
		{
			testName: "with overwrite and verbose=false",
			mockConfig: cl.RunConfig{
				Overwrite: true,
				Verbose:   false,
			},
			expectedConfig: cl.RunConfig{
				Overwrite: true,
				Verbose:   false,
			},
			args: []string{"--overwrite", "--verbose=false"},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			t.Cleanup(configReset)

			err := rootCmd.ParseFlags(test.args)
			if err != nil {
				t.Errorf("Failed to parse flags: %v", err)
			}
			rootCmd.DebugFlags()

			rc = test.mockConfig

			preRunSetup(rootCmd, nil)

			if diff := cmp.Diff(test.expectedConfig, rc, cmpopts.IgnoreFields(cl.RunConfig{}, "Logger", "ExcludedPaths")); diff != "" {
				t.Errorf("TestCase: %s \n Replace{} mismatch (-want +got):\n%s", test.testName, diff)
			}
		})
	}
}
