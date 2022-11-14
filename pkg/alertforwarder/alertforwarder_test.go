package alertforwarder

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/specklesystems/alertmanager-discord/pkg/alertmanager"
	"github.com/specklesystems/alertmanager-discord/pkg/discord"
	"github.com/specklesystems/alertmanager-discord/pkg/prometheus"
	. "github.com/specklesystems/alertmanager-discord/test"

	"github.com/stretchr/testify/assert"
)

func Test_TransformAndForward_HappyPath(t *testing.T) {
	ao := alertmanager.Out{
		Alerts: []alertmanager.Alert{
			{
				Status: alertmanager.StatusFiring,
			},
		},
		CommonAnnotations: struct {
			Summary string `json:"summary"`
		}{
			Summary: "a_common_annotation_summary",
		},
	}

	mockClientRecorder, res := triggerAndRecordRequest(t, ao, http.StatusOK)
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode, "http response status code")

	assert.Equal(t, 1, len(mockClientRecorder.Requests), "Should have sent one request to Discord")
	assert.Equal(t, "application/json", mockClientRecorder.Requests[0].ContentType, "content type")

	do := readerToDiscordOut(t, mockClientRecorder.Requests[0].Body)
	assert.Equal(t, 1, len(do.Embeds), "Discord message embed length")
	assert.Equal(t, 10038562, do.Embeds[0].Color, "Discord message embed color")
	assert.Contains(t, do.Content, "a_common_annotation_summary", "Discord message content")
}

func Test_TransformAndForward_InvalidInput_NoValue_ReturnsErrorResponseCode(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	req.Host = "testing.localhost"

	mockClientRecorder := MockClientRecorder{}
	mockClient := mockClientRecorder.NewMockClientWithResponse(http.StatusBadRequest)

	SUT := NewAlertForwarder(mockClient, "https://discordapp.com/api/webhooks/123456789123456789/abc", 100*time.Millisecond)

	w := httptest.NewRecorder()
	SUT.TransformAndForward(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode, "Should expect an http response status code indicating request was bad.")

	assert.Equal(t, 0, len(mockClientRecorder.Requests), "should not have sent a request to Discord")
}

func Test_TransformAndForward_InvalidInput_LongString_ReturnsErrorResponseCode(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(strings.Repeat("a", 1025)))
	req.Host = "testing.localhost"

	mockClientRecorder := MockClientRecorder{}
	mockClient := mockClientRecorder.NewMockClientWithResponse(http.StatusBadRequest)

	SUT := NewAlertForwarder(mockClient, "https://discordapp.com/api/webhooks/123456789123456789/abc", 100*time.Millisecond)

	w := httptest.NewRecorder()
	SUT.TransformAndForward(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode, "Should expect an http response status code indicating request was bad.")

	assert.Equal(t, 0, len(mockClientRecorder.Requests), "should not have sent a request to Discord")
}

func Test_TransformAndForward_InvalidInput_PrometheusAlert_ReturnsErrorResponseCode(t *testing.T) {
	promAlert := []prometheus.Alert{
		{
			Status: "",
		},
	}
	promAlertJson, err := json.Marshal(promAlert)
	assert.NoError(t, err, "marshalling prometheus alert")
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(promAlertJson))
	req.Host = "testing.localhost"

	mockClientRecorder := MockClientRecorder{}
	mockClient := mockClientRecorder.NewMockClientWithResponse(http.StatusBadRequest)

	SUT := NewAlertForwarder(mockClient, "https://discordapp.com/api/webhooks/123456789123456789/abc", 100*time.Millisecond)

	w := httptest.NewRecorder()
	SUT.TransformAndForward(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode, "Should expect an http response status code indicating server internal error.")

	assert.Equal(t, 1, len(mockClientRecorder.Requests), "should have sent one request to Discord (with a message stating there is a problem)")
	// TODO test message content sent to Discord
}

// FIXME may not be able to simulate error in http Client?
func Test_TransformAndForward_PrometheusAlert_And_DiscordClientResponsdsWithError_RespondsWithErrorCode(t *testing.T) {
	promAlert := []prometheus.Alert{
		{
			Status: "",
		},
	}
	promAlertJson, err := json.Marshal(promAlert)
	assert.NoError(t, err, "marshalling prometheus alert")
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(promAlertJson))
	req.Host = "testing.localhost"

	mockClientRecorder := MockClientRecorder{}
	mockClient := mockClientRecorder.NewMockClientReturnsNil()

	SUT := NewAlertForwarder(mockClient, "https://discordapp.com/api/webhooks/123456789123456789/abc", 100*time.Millisecond)

	w := httptest.NewRecorder()
	SUT.TransformAndForward(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode, "Should expect an http response status code indicating request was unprocessable.")

	assert.GreaterOrEqual(t, len(mockClientRecorder.Requests), 1, "should have sent at least one request to Discord (with a message stating there is a problem)")
	// TODO test message content sent to Discord
}

func Test_TransformAndForward_PrometheusAlert_And_DiscordClientResponsdsWithErrorStatusCode_RespondsWithErrorStatusCode(t *testing.T) {
	promAlert := []prometheus.Alert{
		{
			Status: "",
		},
	}
	promAlertJson, err := json.Marshal(promAlert)
	assert.NoError(t, err, "marshalling prometheus alert")
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(promAlertJson))
	req.Host = "testing.localhost"

	mockClientRecorder := MockClientRecorder{}
	mockClient := mockClientRecorder.NewMockClientWithResponse(http.StatusBadRequest)

	SUT := NewAlertForwarder(mockClient, "https://discordapp.com/api/webhooks/123456789123456789/abc", 100*time.Millisecond)

	w := httptest.NewRecorder()
	SUT.TransformAndForward(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode, "Should expect an http response status code indicating internal server error.")

	assert.Equal(t, 1, len(mockClientRecorder.Requests), "should have sent a request to Discord (with a message stating there is a problem)")
	// TODO test message content sent to Discord
}

func Test_TransformAndForward_NoAlerts_DoesNotSendToDiscord(t *testing.T) {
	ao := alertmanager.Out{}

	mockClientRecorder, res := triggerAndRecordRequest(t, ao, http.StatusBadRequest)
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode, "http response status code")

	assert.Equal(t, 0, len(mockClientRecorder.Requests), "mock client should not be triggered")
}

func Test_TransformAndForward_NoCommonAnnotationSummary_HappyPath(t *testing.T) {
	ao := alertmanager.Out{
		Alerts: []alertmanager.Alert{
			{
				Status: alertmanager.StatusFiring,
			},
		},
	}

	mockClientRecorder, res := triggerAndRecordRequest(t, ao, http.StatusOK)
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode, "http response status code")

	assert.Equal(t, 1, len(mockClientRecorder.Requests), "mock client should be triggered")
	assert.Equal(t, "application/json", mockClientRecorder.Requests[0].ContentType, "content type")

	do := readerToDiscordOut(t, mockClientRecorder.Requests[0].Body)
	assert.Equal(t, 1, len(do.Embeds), "Discord message embed length")
	assert.Equal(t, 10038562, do.Embeds[0].Color, "Discord message embed color")
	assert.Equal(t, "", do.Content, "Discord message content")
}

func Test_TransformAndForward_StatusResolved_HappyPath(t *testing.T) {
	ao := alertmanager.Out{
		Alerts: []alertmanager.Alert{
			{
				Status: alertmanager.StatusResolved,
			},
		},
	}

	mockClientRecorder, res := triggerAndRecordRequest(t, ao, http.StatusOK)
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode, "http response status code")

	assert.Equal(t, 1, len(mockClientRecorder.Requests), "mock client should be triggered")

	do := readerToDiscordOut(t, mockClientRecorder.Requests[0].Body)
	assert.Equal(t, 1, len(do.Embeds), "Discord message embed length")
	assert.Equal(t, 3066993, do.Embeds[0].Color, "Discord message embed color")
}

// alert with a label 'instance'='localhost' and 'exported_instance' label is set, should have the instance replaced by 'exported_instance'
func Test_TransformAndForward_ExportedInstance_HappyPath(t *testing.T) {
	ao := alertmanager.Out{
		Alerts: []alertmanager.Alert{
			{
				Status: alertmanager.StatusFiring,
				Labels: map[string]string{
					"instance":          "localhost",
					"exported_instance": "exported_instance_value",
				},
			},
		},
	}

	mockClientRecorder, res := triggerAndRecordRequest(t, ao, http.StatusOK)
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode, "http response status code")

	assert.Equal(t, 1, len(mockClientRecorder.Requests), "mock client should be triggered")
	assert.Equal(t, "application/json", mockClientRecorder.Requests[0].ContentType, "content type")

	do := readerToDiscordOut(t, mockClientRecorder.Requests[0].Body)
	assert.Equal(t, 1, len(do.Embeds), "Discord message embed length")
	assert.Equal(t, 10038562, do.Embeds[0].Color, "Discord message embed color")
	assert.Equal(t, 1, len(do.Embeds[0].Fields), "Discord message embed fields length")
	assert.Contains(t, do.Embeds[0].Fields[0].Name, "exported_instance_value", "Discord message embed field Name should contain instance")
	assert.Equal(t, "", do.Content, "Discord message content")
}

// FIXME may not be able to create an error in http.Client
// Discord client returns an error (e.g. a closed connection, network outage or similar)
func Test_TransformAndForward_DiscordClientReturnsError(t *testing.T) {
	ao := alertmanager.Out{
		Alerts: []alertmanager.Alert{
			{
				Status: alertmanager.StatusFiring,
			},
		},
		CommonAnnotations: struct {
			Summary string `json:"summary"`
		}{
			Summary: "a_common_annotation_summary",
		},
	}

	mockClientRecorder, res := triggerAndRecordRequest(t, ao, http.StatusBadRequest)
	defer res.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode, "http response status code")

	assert.GreaterOrEqual(t, len(mockClientRecorder.Requests), 1, "Should have sent a request to Discord")
	assert.Equal(t, "application/json", mockClientRecorder.Requests[0].ContentType, "content type")

	do := readerToDiscordOut(t, mockClientRecorder.Requests[0].Body)
	assert.Equal(t, 1, len(do.Embeds), "Discord message embed length")
	assert.Equal(t, 10038562, do.Embeds[0].Color, "Discord message embed color")
	assert.Contains(t, do.Content, "a_common_annotation_summary", "Discord message content")
}

func Test_TransformAndForward_DiscordReturnsWithErrorStatusCode_ReturnInternalServerErrorStatusCode(t *testing.T) {
	ao := alertmanager.Out{
		Alerts: []alertmanager.Alert{
			{
				Status: alertmanager.StatusFiring,
			},
		},
		CommonAnnotations: struct {
			Summary string `json:"summary"`
		}{
			Summary: "a_common_annotation_summary",
		},
	}

	mockClientRecorder, res := triggerAndRecordRequest(t, ao, http.StatusUnauthorized)
	defer res.Body.Close()

	assert.Equal(t, 1, len(mockClientRecorder.Requests), "Should have sent a request to Discord")
	assert.Equal(t, "application/json", mockClientRecorder.Requests[0].ContentType, "content type")

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode, "http response status code should be 500")
}

// TODO Add a test for context with multiple alerts: if some are firing and some resolved we should publish two separate messages to Discord - alerts with matching statuses should be grouped together
func Test_TransformAndForward_MultipleAlerts_DifferentStatus_HappyPath(t *testing.T) {
	ao := alertmanager.Out{
		Alerts: []alertmanager.Alert{
			{
				Status: alertmanager.StatusFiring,
			},
			{
				Status: alertmanager.StatusFiring,
			},
			{
				Status: alertmanager.StatusResolved,
			},
		},
	}

	mockClientRecorder, res := triggerAndRecordRequest(t, ao, http.StatusOK)
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode, "http response status code")

	assert.Equal(t, 2, len(mockClientRecorder.Requests), "Should have sent two requests to Discord")
}

// HELPERS

func triggerAndRecordRequest(t *testing.T, request alertmanager.Out, discordStatusCode int) (mockClientRecorder MockClientRecorder, httpResponse *http.Response) {
	aoJson, err := json.Marshal(request)
	assert.NoError(t, err, "marshalling alertmanager out")

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(aoJson))
	req.Host = "testing.localhost"

	mockClientRecorder = MockClientRecorder{}
	mockClient := mockClientRecorder.NewMockClientWithResponse(discordStatusCode)

	SUT := NewAlertForwarder(mockClient, "https://discordapp.com/api/webhooks/123456789123456789/abc", 100*time.Millisecond)

	w := httptest.NewRecorder()
	SUT.TransformAndForward(w, req)

	httpResponse = w.Result()
	return mockClientRecorder, httpResponse
}

func readerToDiscordOut(t *testing.T, reader io.Reader) discord.Out {
	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)

	do := discord.Out{}
	err := json.Unmarshal(buf.Bytes(), &do)
	if err != nil {
		t.Errorf("Unexpected error marshalling to Discord Object from the Discord client request body.")
	}
	return do
}
