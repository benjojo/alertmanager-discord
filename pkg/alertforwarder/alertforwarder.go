package alertforwarder

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/specklesystems/alertmanager-discord/pkg/alertmanager"
	"github.com/specklesystems/alertmanager-discord/pkg/discord"
	"github.com/specklesystems/alertmanager-discord/pkg/logging"
	"github.com/specklesystems/alertmanager-discord/pkg/prometheus"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

const (
	maxLogLength = 1024
)

type AlertForwarderHandler struct {
	af AlertForwarder
}

func NewAlertForwarderHandler(client *http.Client, webhookURL string, maximumBackoffTimeSeconds time.Duration) *AlertForwarderHandler {
	return &AlertForwarderHandler{
		af: NewAlertForwarder(client, webhookURL, maximumBackoffTimeSeconds),
	}
}

func (h *AlertForwarderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.af.TransformAndForward(w, r)
}

type AlertForwarder struct {
	client *discord.Client
}

func NewAlertForwarder(client *http.Client, webhookURL string, maximumBackoffTimeSeconds time.Duration) AlertForwarder {
	return AlertForwarder{
		client: discord.NewClient(client, webhookURL, maximumBackoffTimeSeconds),
	}
}

func (af *AlertForwarder) sendWebhook(correlationId string, amo *alertmanager.Out, w http.ResponseWriter) {
	if len(amo.Alerts) < 1 {
		log.Debug().
			Str(logging.FieldKeyCorrelationId, correlationId).
			Msg("There are no alerts within this notification. There is nothing to forward to Discord. Returning early...")
		w.WriteHeader(http.StatusOK)
		return
	}

	groupedAlerts := make(map[string][]alertmanager.Alert)
	for _, alert := range amo.Alerts {
		groupedAlerts[alert.Status] = append(groupedAlerts[alert.Status], alert)
	}

	failedToPublishAtLeastOne := false
	for status, alerts := range groupedAlerts {
		DO := TranslateAlertManagerToDiscord(status, amo, alerts)

		log.Info().
			Str(logging.FieldKeyEventType, logging.EventTypeRequestSending).
			Str(logging.FieldKeyCorrelationId, correlationId).
			Msg("Sending HTTP request to Discord.")
		res, err := af.client.PublishMessage(DO)
		if err != nil {
			err = fmt.Errorf("Error encountered when publishing message to discord: %w", err)
			log.Error().
				Str(logging.FieldKeyCorrelationId, correlationId).
				Err(err).
				Msg("Error when attempting to publish message to discord.")
			failedToPublishAtLeastOne = true
			continue
		}

		log.Info().
			Str(logging.FieldKeyEventType, logging.EventTypeResponseReceived).
			Str(logging.FieldKeyCorrelationId, correlationId).
			Msg("HTTP response received from Discord")

		if res.StatusCode < 200 || res.StatusCode > 399 {
			failedToPublishAtLeastOne = true
			continue
		}
	}

	if failedToPublishAtLeastOne {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (af *AlertForwarder) sendRawPromAlertWarn(correlationId string) (*http.Response, error) {

	warningMessage := `You have probably misconfigured this software.
We detected input in Prometheus Alert format but are expecting AlertManager format.
This program is intended to ingest alerts from alertmanager.
It is not a replacement for alertmanager, it is a
webhook target for it. Please read the README.md
for guidance on how to configure it for alertmanager
or https://prometheus.io/docs/alerting/latest/configuration/#webhook_config`
	log.Warn().Msg(warningMessage)
	DO := discord.Out{
		Content: "",
		Embeds: []discord.Embed{
			{
				Title:       "You have misconfigured this software",
				Description: warningMessage,
				Color:       discord.ColorGrey,
				Fields:      []discord.EmbedField{},
			},
		},
	}

	log.Info().
		Str(logging.FieldKeyEventType, logging.EventTypeRequestSending).
		Str(logging.FieldKeyCorrelationId, correlationId).
		Msg("Sending HTTP request to Discord.")
	res, err := af.client.PublishMessage(DO)
	if err != nil {
		return nil, fmt.Errorf("Error encountered when publishing message to discord: %w", err)
	}

	log.Info().
		Str(logging.FieldKeyEventType, logging.EventTypeResponseReceived).
		Str(logging.FieldKeyCorrelationId, correlationId).
		Msg("HTTP response received from Discord")
	return res, nil
}

func (af *AlertForwarder) TransformAndForward(w http.ResponseWriter, r *http.Request) {
	correlationId := uuid.New().String()
	log.Info().
		Str(logging.FieldKeyHttpHost, r.Host).
		Str(logging.FieldKeyHttpMethod, r.Method).
		Str(logging.FieldKeyHttpPath, r.URL.Path).
		Str(logging.FieldKeyEventType, logging.EventTypeRequestReceived).
		Str(logging.FieldKeyCorrelationId, correlationId).
		Msg("HTTP request received from AlertManager.")
	defer log.Info().
		Str(logging.FieldKeyEventType, logging.EventTypeResponseSending).
		Str(logging.FieldKeyCorrelationId, correlationId).
		Msg("Sending HTTP response to AlertManager.")

	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().
			Str(logging.FieldKeyCorrelationId, correlationId).
			Err(err).
			Msg("Unable to read request body.")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	amo := alertmanager.Out{}
	err = json.Unmarshal(b, &amo)
	if err != nil {
		af.handleInvalidInput(correlationId, b, w)
		return
	}

	af.sendWebhook(correlationId, &amo, w)
}

func (af *AlertForwarder) handleInvalidInput(correlationId string, b []byte, w http.ResponseWriter) {
	if prometheus.IsAlert(b) {
		log.Info().
			Str(logging.FieldKeyCorrelationId, correlationId).
			Msg("Detected a Prometheus Alert, and not an AlertManager alert, has been sent within the http request. This indicates a misconfiguration. Attempting to send a message to notify the Discord channel of the misconfiguration.")
		res, err := af.sendRawPromAlertWarn(correlationId)
		if err != nil || (res != nil && res.StatusCode < 200 || res.StatusCode > 399) {
			statusCode := 0
			if res != nil {
				statusCode = res.StatusCode
			}

			log.Error().
				Err(err).
				Str(logging.FieldKeyCorrelationId, correlationId).
				Int(logging.FieldKeyStatusCode, statusCode).
				Msg("Error when attempting to send a warning message to Discord.")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	if len(b) > maxLogLength-3 {
		log.Info().
			Str(logging.FieldKeyCorrelationId, correlationId).
			Msgf("Failed to unpack inbound alert request - %s...", string(b[:maxLogLength-3]))
	} else {
		log.Info().
			Str(logging.FieldKeyCorrelationId, correlationId).
			Msgf("Failed to unpack inbound alert request - %s", string(b))
	}

	w.WriteHeader(http.StatusBadRequest)
	return
}
