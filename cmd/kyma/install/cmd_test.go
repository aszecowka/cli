package install

import (
	"testing"

	"errors"

	stepMocks "github.com/kyma-project/cli/internal/step/mocks"
	trustMocks "github.com/kyma-project/cli/internal/trust/mocks"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/stretchr/testify/require"
)

func Test_RemoveActionLabel(t *testing.T) {
	testData := []struct {
		testName       string
		data           []map[string]interface{}
		expectedResult []map[string]interface{}
		shouldFail     bool
	}{
		{
			testName: "correct data test",
			data: []map[string]interface{}{
				{
					"apiVersion": "installer.kyma-project.io/v1alpha1",
					"kind":       "Installation",
					"metadata": map[interface{}]interface{}{
						"name": "kyma-installation",
						"labels": map[interface{}]interface{}{
							"action": "install",
						},
					},
				},
			},
			expectedResult: []map[string]interface{}{
				{
					"apiVersion": "installer.kyma-project.io/v1alpha1",
					"kind":       "Installation",
					"metadata": map[interface{}]interface{}{
						"name":   "kyma-installation",
						"labels": map[interface{}]interface{}{},
					},
				},
			},
			shouldFail: false,
		},
		{
			testName: "incorrect data test",
			data: []map[string]interface{}{
				{
					"apiVersion": "installer.kyma-project.io/v1alpha1",
					"kind":       "Installation",
					"metadata": map[interface{}]interface{}{
						"name":   "kyma-installation",
						"labels": map[interface{}]interface{}{},
					},
				},
			},
			expectedResult: []map[string]interface{}{
				{
					"apiVersion": "installer.kyma-project.io/v1alpha1",
					"kind":       "Installation",
					"metadata": map[interface{}]interface{}{
						"name":   "kyma-installation",
						"labels": map[interface{}]interface{}{},
					},
				},
			},
			shouldFail: true,
		},
	}

	cmd := &command{
		opts: nil,
	}

	for _, tt := range testData {
		err := cmd.removeActionLabel(tt.data)
		if !tt.shouldFail {
			require.Nil(t, err, tt.testName)
			require.Equal(t, tt.data, tt.expectedResult, tt.testName)
		} else {
			require.Equal(t, tt.data, tt.expectedResult, tt.testName)
		}
	}
}

func Test_ReplaceDockerImageURL(t *testing.T) {
	const replacedWithData = "testImage!"
	testData := []struct {
		testName       string
		data           []map[string]interface{}
		expectedResult []map[string]interface{}
		shouldFail     bool
	}{
		{
			testName: "correct data test",
			data: []map[string]interface{}{
				{
					"apiVersion": "installer.kyma-project.io/v1alpha1",
					"kind":       "Deployment",
					"spec": map[interface{}]interface{}{
						"template": map[interface{}]interface{}{
							"spec": map[interface{}]interface{}{
								"serviceAccountName": "kyma-installer",
								"containers": []interface{}{
									map[interface{}]interface{}{
										"name":  "kyma-installer-container",
										"image": "eu.gcr.io/kyma-project/kyma-installer:63f27f76",
									},
								},
							},
						},
					},
				},
			},
			expectedResult: []map[string]interface{}{
				{
					"apiVersion": "installer.kyma-project.io/v1alpha1",
					"kind":       "Deployment",
					"spec": map[interface{}]interface{}{
						"template": map[interface{}]interface{}{
							"spec": map[interface{}]interface{}{
								"serviceAccountName": "kyma-installer",
								"containers": []interface{}{
									map[interface{}]interface{}{
										"name":  "kyma-installer-container",
										"image": replacedWithData,
									},
								},
							},
						},
					},
				},
			},
			shouldFail: false,
		},
	}

	cmd := &command{
		opts: nil,
	}

	for _, tt := range testData {
		res, err := cmd.replaceDockerImageURL(tt.data, replacedWithData)
		if !tt.shouldFail {
			require.Nil(t, err, tt.testName)
			require.Equal(t, res, tt.expectedResult, tt.testName)
		} else {
			require.NotNil(t, err, tt.testName)
		}
	}
}

func TestImportCertificate(t *testing.T) {
	cases := []struct {
		// params
		name        string
		description string
		cert        trustMocks.Certifier
		wait        bool
		// results
		success            bool
		stopped            bool
		expectedStepStatus []string
		expectedStepInfos  []string
		expectedStepErrors []string
		expectedErr        error
	}{
		{
			name:        "Certificate import",
			description: "Imports the correct certificate",
			cert: trustMocks.Certifier{
				Crt: "Hi, I am a fake certificate!",
			},
			wait:               true,
			success:            true,
			stopped:            false,
			expectedStepStatus: []string{"Kyma root certificate imported"},
			expectedErr:        nil,
		},
		{
			name:        "Certificate retrieval failed",
			description: "Not possible to retrieve the certificate",
			cert: trustMocks.Certifier{
				Crt: "",
			},
			wait:        true,
			success:     false,
			stopped:     false,
			expectedErr: errors.New("Could not retrieve the certificate"),
		},
		{
			name:        "No Wait",
			description: "Certificate not imported due to not waiting for Kyma installation",
			cert: trustMocks.Certifier{
				Crt: "", // empty because certificate retrieval should not be attempted
			},
			wait:               false,
			success:            false,
			stopped:            false,
			expectedStepErrors: []string{"Manual OS-specific instructions for certificate import"},
			expectedErr:        nil,
		},
	}

	cmd := command{
		opts: NewOptions(cli.NewOptions()),
	}

	mockStep := &stepMocks.Step{}
	cmd.CurrentStep = mockStep

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			cmd.opts.NoWait = !test.wait
			err := cmd.importCertificate(test.cert)

			require.Equal(t, test.expectedErr, err, "Error not as expected")
			require.Equal(t, test.success, mockStep.IsSuccessful(), "Import certificate step must be successful")
			require.Equal(t, test.stopped, mockStep.IsStopped(), "Import certificate step must not be stopped")
			require.Equal(t, test.expectedStepStatus, mockStep.Statuses(), "Status messages not as expected")
			require.Equal(t, test.expectedStepInfos, mockStep.Infos(), "Logged info messages not as expected")
			require.Equal(t, test.expectedStepErrors, mockStep.Errors(), "Logged error messages not as expected")

			mockStep.Reset()
		})
	}

}
