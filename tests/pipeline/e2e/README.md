# InstructLab Pipeline Tests on RHOAI

This test validates the InstructLab pipeline functionality on the RHOAI platform.

### Prerequisites

* Admin access to an OpenShift cluster.

* Follow the instructions from [here](https://github.com/opendatahub-io/ilab-on-ocp/blob/main/README.md) to set up the InstructLab Pipeline on RHOAI.

* Once the InstructLab Pipeline is imported, you are ready to run the test.

### Setup

* Set the following environment variables:

  * ENABLE_ILAB_PIPELINE_TEST: Set to true to enable the test.
  * PIPELINE_SERVER_URL: The URL of the pipeline server.
  * BEARER_TOKEN: A valid bearer token for authentication.
  * PIPELINE_DISPLAY_NAME: The display name of the pipeline to be tested.

* Trust the cluster's self-signed certificates:

   * Download the certificates from the cluster and add them to your trusted certificate store.

### Execution

Run the test using the following command:

```bash
go test -run TestPipelineRun -v -timeout 180m ./pipeline/e2e/
```
This will execute the pipeline test and validate its successful completion.
