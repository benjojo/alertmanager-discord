package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type alertManOut struct {
	Alerts []struct {
		Annotations struct {
			Description string `json:"description"`
			Summary     string `json:"summary"`
		} `json:"annotations"`
		EndsAt       string            `json:"endsAt"`
		GeneratorURL string            `json:"generatorURL"`
		Labels       map[string]string `json:"labels"`
		StartsAt     string            `json:"startsAt"`
		Status       string            `json:"status"`
	} `json:"alerts"`
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
	Content string `json:"content"`
	Name    string `json:"username"`
}

func main() {
	webhookUrl := os.Getenv("webhook")
	whURL := flag.String("webhook.url", webhookUrl, "")
	flag.Parse()
	http.ListenAndServe(":9094", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}

		amo := alertManOut{}
		err = json.Unmarshal(b, &amo)
		if err != nil {
			panic(err)
		}

		DO := discordOut{
			Name: amo.Status,
		}

		Content := "```"
		if amo.CommonAnnotations.Summary != "" {
			Content = fmt.Sprintf(" === %s === \n```", amo.CommonAnnotations.Summary)
		}

		for _, alert := range amo.Alerts {
			realname := alert.Labels["instance"]
			if strings.Contains(realname, "localhost") && alert.Labels["exported_instance"] != "" {
				realname = alert.Labels["exported_instance"]
			}
			Content += fmt.Sprintf("[%s]: %s on %s\n%s\n", strings.ToUpper(amo.Status), alert.Labels["alertname"], realname, alert.Annotations.Description)
		}

		DO.Content = Content + "```"

		DOD, _ := json.Marshal(DO)
		http.Post(*whURL, "application/json", bytes.NewReader(DOD))
	}))
}
