package alertforwarder

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_WebhookUrl_HappyPath(t *testing.T) {
	ok, _, _ := CheckWebhookURL("https://discordapp.com/api/webhooks/123456789123456789/abc")
	assert.True(t, ok, "Should be a valid webhook url")

	ok, _, _ = CheckWebhookURL("https://discord.com/api/webhooks/123456789123456789/abc")
	assert.True(t, ok, "Should be a valid webhook url")

	ok, _, _ = CheckWebhookURL("http://localhost/")
	assert.True(t, ok, "Should be a valid webhook url")

	ok, _, _ = CheckWebhookURL("http://127.0.0.1/")
	assert.True(t, ok, "Should be a valid webhook url")

	ok, _, _ = CheckWebhookURL("http://::1/")
	assert.True(t, ok, "Should be a valid webhook url")
}

func Test_WebhookUrl_EmptyUrl_ReturnsFalse(t *testing.T) {
	ok, _, err := CheckWebhookURL("")
	assert.False(t, ok, "Empty url should be identified as invalid")
	assert.Error(t, err, "Empty url should return an error message")
}

func Test_WebhookUrl_InvalidUrl_ReturnsFalse(t *testing.T) {
	ok, _, err := CheckWebhookURL("::::::::::")
	assert.False(t, ok, "Malformed urls should be identified as invalid")
	assert.Error(t, err, "Invalid url should return an error message")
}

func Test_WebhookUrl_InvalidAPIUrl_ReturnsFalse(t *testing.T) {
	ok, _, err := CheckWebhookURL("https://discordapp.com/api/webhooks/12/abc")
	assert.False(t, ok, "Malformed Discord API urls should be identified as invalid")
	assert.Error(t, err, "Malformed Discord API url should return an error message")

	ok, _, err = CheckWebhookURL("https://example.org/api/webhooks/12/abc")
	assert.False(t, ok, "Non-Discord urls should be identified as invalid")
	assert.Error(t, err, "Non-Discord urls should return an error message")
}
