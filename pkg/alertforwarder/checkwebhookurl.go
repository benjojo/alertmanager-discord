package alertforwarder

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"

	"github.com/specklesystems/alertmanager-discord/pkg/flags"
)

func CheckWebhookURL(webhookURL string) (bool, *url.URL, error) {
	if webhookURL == "" {
		return false, &url.URL{}, fmt.Errorf("'%s' has not been set. This can be provided via configuration file, command line, or environment variable ('%s')", flags.DiscordWebhookUrlFlagKey, strings.ToUpper(flags.DiscordWebhookUrlFlagKey))
	}

	parsedUrl, err := url.Parse(webhookURL)
	if err != nil {
		return false, &url.URL{}, fmt.Errorf("The Discord WebHook URL ('%s') cannot be parsed as a url: %w", webhookURL, err)
	}

	host, _, err := net.SplitHostPort(parsedUrl.Host)
	if err != nil {
		// return false, parsedUrl, fmt.Errorf("The Discord WebHook URL ('%s') host ('%s') cannot be separated into domain/ip and port components: %w", webhookURL, parsedUrl.Host, err)
	}
	if host == "" {
		host = parsedUrl.Host
	}

	// localhost is allowed, for testing or for proxied routes etc..
	if host == "127.0.0.1" || host == "::1" || host == "localhost" {
		return true, parsedUrl, nil
	}

	re := regexp.MustCompile(`https://discord(?:app)?.com/api/webhooks/[0-9]{18,19}/[a-zA-Z0-9_-]+`)

	ok := re.Match([]byte(webhookURL))
	if !ok {
		return false, parsedUrl, fmt.Errorf("The Discord WebHook URL doesn't seem to be a valid Discord Webhook API url: '%s'", webhookURL)
	}

	return ok, parsedUrl, nil
}
