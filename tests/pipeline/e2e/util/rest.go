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

package testUtil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type PipelineRequest struct {
	DisplayName              string `json:"display_name"`
	PipelineVersionReference struct {
		PipelineID string `json:"pipeline_id"`
	} `json:"pipeline_version_reference"`
	RuntimeConfig struct {
		Parameters map[string]interface{} `json:"parameters"`
	} `json:"runtime_config"`
}

type Pipeline struct {
	Pipelines []struct {
		PipelineID  string `json:"pipeline_id"`
		DisplayName string `json:"display_name"`
	} `json:"pipelines"`
}

func RetrievePipelineId(t *testing.T, pipelineServerURL, pipelineDisplayName, bearerToken string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/apis/v2beta1/pipelines", pipelineServerURL), nil)
	require.NoError(t, err, "Failed to create request")

	// Add the Bearer token to the Authorization header
	req.Header.Add("Authorization", "Bearer "+bearerToken)

	response, err := client.Do(req)
	require.NoError(t, err, "Failed to retrieve pipelines")
	defer response.Body.Close()

	responseData, err := io.ReadAll(response.Body)
	require.NoError(t, err, "Failed to read response body")

	var pipelineData Pipeline
	err = json.Unmarshal(responseData, &pipelineData)
	require.NoError(t, err, "Failed to parse pipeline data")

	for _, pipeline := range pipelineData.Pipelines {
		if pipeline.DisplayName == pipelineDisplayName {
			return pipeline.PipelineID, nil
		}
	}

	return "", fmt.Errorf("pipeline with display name '%s' not found", pipelineDisplayName)
}

// TriggerPipeline starts the pipeline and returns the run ID
func TriggerPipeline(t *testing.T, pipelineServerURL, pipelineID, pipelineDisplayName string, parameters map[string]interface{}, bearerToken string) (string, error) {
	client := &http.Client{}
	payload := PipelineRequest{
		DisplayName: pipelineDisplayName,
		PipelineVersionReference: struct {
			PipelineID string `json:"pipeline_id"`
		}{PipelineID: pipelineID},
		RuntimeConfig: struct {
			Parameters map[string]interface{} `json:"parameters"`
		}{Parameters: parameters},
	}

	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err, "Failed to marshal pipeline request payload")
	url := fmt.Sprintf("%s/apis/v2beta1/runs", pipelineServerURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	require.NoError(t, err, "Failed to create HTTP request")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+bearerToken)

	resp, err := client.Do(req)
	require.NoError(t, err, "Failed to execute HTTP request")
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response body")

	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	require.NoError(t, err, "Failed to parse response")

	runID, ok := response["run_id"].(string)
	if !ok {
		return "", fmt.Errorf("run_id not found in response")
	}
	return runID, nil
}

// WaitForPipelineSuccess polls the pipeline run status until it succeeds or times out
func WaitForPipelineSuccess(t *testing.T, pipelineServerURL, runID string, bearerToken string) error {
	client := &http.Client{}

	url := fmt.Sprintf("%s/apis/v2beta1/runs/%s", pipelineServerURL, runID)

	timeout := time.After(2*time.Hour + 10*time.Minute)
	tick := time.Tick(1 * time.Minute) // Poll every 1 minute

	for {
		select {
		case <-timeout:
			return fmt.Errorf("pipeline run %s timed out", runID)
		case <-tick:
			req, err := http.NewRequest("GET", url, nil)
			require.NoError(t, err, "Failed to create HTTP request")

			// Add Bearer token for authorization
			req.Header.Add("Authorization", "Bearer "+bearerToken)

			resp, err := client.Do(req)
			require.NoError(t, err, "Failed to retrieve pipeline run status")

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err, "Failed to read response body")

			var data map[string]interface{}
			err = json.Unmarshal(body, &data)
			require.NoError(t, err, "Failed to parse pipeline run status")

			state, ok := data["state"].(string)
			if !ok {
				return fmt.Errorf("invalid state format in pipeline run status")
			}

			switch state {
			case "SUCCEEDED":
				return nil
			case "SKIPPED", "FAILED", "CANCELING", "CANCELED", "PAUSED":
				return fmt.Errorf("pipeline run failed with status: %s", state)
			}
		}
	}
}
