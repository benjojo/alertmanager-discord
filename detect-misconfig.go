package main

import "encoding/json"

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
