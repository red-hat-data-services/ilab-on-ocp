/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package odh

import (
	"os"
	"testing"

	TestUtil "github.com/opendatahub-io/ilab-on-ocp/tests/pipeline/e2e/util"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestPipelineRun(t *testing.T) {
	t.Log("Starting TestPipelineRun...")

	if os.Getenv("ENABLE_ILAB_PIPELINE_TEST") != "true" {
		t.Skip("Skipping iLab pipeline test. Set ENABLE_ILAB_PIPELINE_TEST=true to enable.")
	}

	t.Log("Checking required environment variables...")

	pipelineServerURL := os.Getenv("PIPELINE_SERVER_URL")
	require.NotEmpty(t, pipelineServerURL, "PIPELINE_SERVER_URL environment variable must be set")

	bearerToken := os.Getenv("BEARER_TOKEN")
	require.NotEmpty(t, bearerToken, "BEARER_TOKEN environment variable must be set")

	pipelineDisplayName := os.Getenv("PIPELINE_DISPLAY_NAME")
	require.NotEmpty(t, pipelineDisplayName, "PIPELINE_DISPLAY_NAME environment variable must be set")

	t.Logf("Retrieving pipeline ID for display name: %s", pipelineDisplayName)

	// Retrieve the pipeline ID
	pipelineID, err := TestUtil.RetrievePipelineId(t, pipelineServerURL, pipelineDisplayName, bearerToken)
	require.NoError(t, err, "Failed to retrieve pipeline ID")
	t.Log("Pipeline loaded successfully.")

	// Load input parameters for the pipeline
	t.Log("Loading pipeline parameters")
	viper.SetConfigName("pipeline_params")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("../e2e/resources/")

	err = viper.ReadInConfig()
	require.NoError(t, err, "Error loading pipeline parameters")
	t.Log("Parameter config loaded successfully.")

	paramsMap := viper.AllSettings()
	t.Log("Successfully loaded and converted pipeline parameters.")
	// Trigger the pipeline run
	runID, err := TestUtil.TriggerPipeline(t, pipelineServerURL, pipelineID, pipelineDisplayName, paramsMap, bearerToken)
	require.NoError(t, err, "Failed to trigger pipeline")
	t.Logf("Pipeline with name %s and run ID %s started....", pipelineDisplayName, runID)

	// Verify the pipeline's successful completion
	t.Log("Waiting for pipeline to complete successfully...")
	err = TestUtil.WaitForPipelineSuccess(t, pipelineServerURL, runID, bearerToken)
	require.NoError(t, err, "Pipeline did not complete successfully")
	t.Logf("Pipeline with name %s and run ID %s finished successfully!", pipelineDisplayName, runID)
}
