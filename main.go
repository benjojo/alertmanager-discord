package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

// Discord color values.
const (
	ColorRed   = 0x992D22
	ColorGreen = 0x2ECC71
	ColorGrey  = 0x95A5A6
)

type alertManAlert struct {
	Annotations struct {
		Description string `json:"description"`
		Summary     string `json:"summary"`
	} `json:"annotations"`
	EndsAt       string            `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	Labels       map[string]string `json:"labels"`
	StartsAt     string            `json:"startsAt"`
	Status       string            `json:"status"`
}

type rawPromAlert struct {
	Annotations struct {
		Description string `json:"description"`
		Summary     string `json:"summary"`
	} `json:"annotations"`
	EndsAt       string            `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	Labels       map[string]string `json:"labels"`
	StartsAt     string            `json:"startsAt"`
	Status       string            `json:"status"`
}

type alertManOut struct {
	Alerts            []alertManAlert `json:"alerts"`
	CommonAnnotations struct {
		Summary string `json:"summary"`
	} `json:"commonAnnotations"`
	CommonLabels struct {
		Alertname string `json:"alertname"`
	} `json:"commonLabels"`
	ExternalURL string `json:"externalURL"`
	GroupKey    string `json:"groupKey"`
	GroupLabels struct {
		Alertname string `json:"alertname"`
	} `json:"groupLabels"`
	Receiver string `json:"receiver"`
	Status   string `json:"status"`
	Version  string `json:"version"`
}

type discordOut struct {
	Content string         `json:"content"`
	Embeds  []discordEmbed `json:"embeds"`
}

type discordEmbed struct {
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Color       int                 `json:"color"`
	Fields      []discordEmbedField `json:"fields"`
}

type discordEmbedField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type discordBot struct {
	WebhookURL string
}

const defaultListenAddress = "127.0.0.1:9094"

func isRawPromAlert(b []byte) bool {
	alertTest := make([]rawPromAlert, 0)
	err := json.Unmarshal(b, &alertTest)
	if err == nil {
		if len(alertTest) != 0 {
			if alertTest[0].Status == "" {
				// Ok it's more than likely then
				return true
			}
		}
	}
	return false
}

func checkWhURL() (string, string) {
	flag.Parse()
	whURL := flag.String("webhook.url", os.Getenv("DISCORD_WEBHOOK"), "Discord WebHook URL.")
	listenAddress := flag.String("listen.address", os.Getenv("LISTEN_ADDRESS"), "Address:Port to listen on.")
	if *whURL == "" {
		log.Fatalf("Environment variable 'DISCORD_WEBHOOK' or CLI parameter 'webhook.url' not found.")
	}
	_, err := url.Parse(*whURL)
	if err != nil {
		log.Fatalf("The Discord WebHook URL doesn't seem to be a valid URL.")
	}

	re := regexp.MustCompile(`https://discord(?:app)?.com/api/webhooks/[0-9]{18,19}/[a-zA-Z0-9_-]+`)
	if ok := re.Match([]byte(*whURL)); !ok {
		log.Error("The Discord WebHook URL doesn't seem to be valid.")
	}
	return *whURL, *listenAddress
}

func (d *discordBot) sendWebhook(amo *alertManOut) {
	groupedAlerts := make(map[string][]alertManAlert)

	for _, alert := range amo.Alerts {
		groupedAlerts[alert.Status] = append(groupedAlerts[alert.Status], alert)
	}

	for status, alerts := range groupedAlerts {
		do := discordOut{}

		richEmbed := discordEmbed{
			Title:       fmt.Sprintf("[%s:%d] %s", strings.ToUpper(status), len(alerts), amo.CommonLabels.Alertname),
			Description: amo.CommonAnnotations.Summary,
			Color:       ColorGrey,
			Fields:      []discordEmbedField{},
		}

		if status == "firing" {
			richEmbed.Color = ColorRed
		} else if status == "resolved" {
			richEmbed.Color = ColorGreen
		}

		if amo.CommonAnnotations.Summary != "" {
			do.Content = fmt.Sprintf(" === %s === \n", amo.CommonAnnotations.Summary)
		}

		for _, alert := range alerts {
			realname := alert.Labels["instance"]
			if strings.Contains(realname, "localhost") && alert.Labels["exported_instance"] != "" {
				realname = alert.Labels["exported_instance"]
			}

			richEmbed.Fields = append(richEmbed.Fields, discordEmbedField{
				Name:  fmt.Sprintf("[%s]: %s on %s", strings.ToUpper(status), alert.Labels["alertname"], realname),
				Value: alert.Annotations.Description,
			})
		}

		do.Embeds = []discordEmbed{richEmbed}

		dod, _ := json.Marshal(do)
		r, err := http.Post(d.WebhookURL, "application/json", bytes.NewReader(dod))
		if err != nil {
			log.Error(err)
		}
		err = r.Body.Close()
		if err != nil {
			log.Error(err)
		}
	}
}

func (d *discordBot) sendRawPromAlertWarn() {
	badString := `This program is suppose to be fed by alertmanager.` + "\n" +
		`It is not a replacement for alertmanager, it is a ` + "\n" +
		`webhook target for it. Please read the README.md  ` + "\n" +
		`for guidance on how to configure it for alertmanager` + "\n" +
		`or https://prometheus.io/docs/alerting/latest/configuration/#webhook_config`

	log.Error(`/!\ -- You have misconfigured this software -- /!\`)
	log.Error(`--- --                                      -- ---`)
	log.Error(badString)

	do := discordOut{
		Content: "",
		Embeds: []discordEmbed{
			{
				Title:       "You have misconfigured this software",
				Description: badString,
				Color:       ColorGrey,
				Fields:      []discordEmbedField{},
			},
		},
	}

	dod, _ := json.Marshal(do)
	r, err := http.Post(d.WebhookURL, "application/json", bytes.NewReader(dod))
	if err != nil {
		log.Error(err)
	}
	err = r.Body.Close()
	if err != nil {
		log.Error(err)
	}
}

func (d *discordBot) alertMessage(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	amo := alertManOut{}
	err = json.Unmarshal(b, &amo)
	if err != nil {
		if isRawPromAlert(b) {
			d.sendRawPromAlertWarn()
			return
		}

		if len(b) > 1024 {
			log.Errorf("Failed to unpack inbound alert request - %s...", string(b[:1023]))
		} else {
			log.Errorf("Failed to unpack inbound alert request - %s", string(b))
		}

		return
	}
	d.sendWebhook(&amo)
}

func main() {
	whURL, listenAddress := checkWhURL()
	d := discordBot{WebhookURL: whURL}

	if listenAddress == "" {
		listenAddress = defaultListenAddress
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/healtz", func(http.ResponseWriter, *http.Request){}).Methods(http.MethodGet)
	mux.HandleFunc("/", d.alertMessage)
	srv := http.Server{
		Addr:              ":9094",
		WriteTimeout:      5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		Handler:           mux,
	}

	log.Infof("Listening on: %s", listenAddress)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %s\n", err)
	}
}
