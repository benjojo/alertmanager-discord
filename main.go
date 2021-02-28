package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

// Discord color values
const (
	ColorRed   = 10038562
	ColorGreen = 3066993
	ColorGrey  = 9807270
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

const defaultListenAddress = "127.0.0.1:9094"

func main() {
	envWhURL := os.Getenv("DISCORD_WEBHOOK")
	whURL := flag.String("webhook.url", envWhURL, "Discord WebHook URL.")

	envListenAddress := os.Getenv("LISTEN_ADDRESS")
	listenAddress := flag.String("listen.address", envListenAddress, "Address:Port to listen on.")

	flag.Parse()

	if *whURL == "" {
		log.Fatalf("Environment variable 'DISCORD_WEBHOOK' or CLI parameter 'webhook.url' not found.")
	}

	if *listenAddress == "" {
		*listenAddress = defaultListenAddress
	}

	_, err := url.Parse(*whURL)
	if err != nil {
		log.Fatalf("The Discord WebHook URL doesn't seem to be a valid URL.")
	}

	re := regexp.MustCompile(`https://discord(?:app)?.com/api/webhooks/[0-9]{18}/[a-zA-Z0-9_-]+`)
	if ok := re.Match([]byte(*whURL)); !ok {
		log.Printf("The Discord WebHook URL doesn't seem to be valid.")
	}

	log.Printf("Listening on: %s", *listenAddress)
	http.ListenAndServe(*listenAddress, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s - [%s] %s", r.Host, r.Method, r.URL.RawPath)

		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}

		amo := alertManOut{}
		err = json.Unmarshal(b, &amo)
		if err != nil {
			if isRawPromAlert(b) {
				badString := `This program is suppose to be fed by alertmanager.` + "\n" +
					`It is not a replacement for alertmanager, it is a ` + "\n" +
					`webhook target for it. Please read the README.md  ` + "\n" +
					`for guidance on how to configure it for alertmanager` + "\n" +
					`or https://prometheus.io/docs/alerting/latest/configuration/#webhook_config`

				log.Print(`/!\ -- You have misconfigured this software -- /!\`)
				log.Print(`--- --                                      -- ---`)
				log.Print(badString)

				DO := discordOut{
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

				DOD, _ := json.Marshal(DO)
				http.Post(*whURL, "application/json", bytes.NewReader(DOD))
				return
			}

			if len(b) > 1024 {
				log.Printf("Failed to unpack inbound alert request - %s...", string(b[:1023]))

			} else {
				log.Printf("Failed to unpack inbound alert request - %s", string(b))
			}

			return
		}

		groupedAlerts := make(map[string][]alertManAlert)

		for _, alert := range amo.Alerts {
			groupedAlerts[alert.Status] = append(groupedAlerts[alert.Status], alert)
		}

		for status, alerts := range groupedAlerts {
			DO := discordOut{}

			RichEmbed := discordEmbed{
				Title:       fmt.Sprintf("[%s:%d] %s", strings.ToUpper(status), len(alerts), amo.CommonLabels.Alertname),
				Description: amo.CommonAnnotations.Summary,
				Color:       ColorGrey,
				Fields:      []discordEmbedField{},
			}

			if status == "firing" {
				RichEmbed.Color = ColorRed
			} else if status == "resolved" {
				RichEmbed.Color = ColorGreen
			}

			if amo.CommonAnnotations.Summary != "" {
				DO.Content = fmt.Sprintf(" === %s === \n", amo.CommonAnnotations.Summary)
			}

			for _, alert := range alerts {
				realname := alert.Labels["instance"]
				if strings.Contains(realname, "localhost") && alert.Labels["exported_instance"] != "" {
					realname = alert.Labels["exported_instance"]
				}

				RichEmbed.Fields = append(RichEmbed.Fields, discordEmbedField{
					Name:  fmt.Sprintf("[%s]: %s on %s", strings.ToUpper(status), alert.Labels["alertname"], realname),
					Value: alert.Annotations.Description,
				})
			}

			DO.Embeds = []discordEmbed{RichEmbed}

			DOD, _ := json.Marshal(DO)
			http.Post(*whURL, "application/json", bytes.NewReader(DOD))
		}
	}))
}
