package alertforwarder

import (
	"fmt"
	"sort"
	"strings"

	"github.com/specklesystems/alertmanager-discord/pkg/alertmanager"
	"github.com/specklesystems/alertmanager-discord/pkg/discord"
)

func TranslateAlertManagerToDiscord(status string, amo *alertmanager.Out, alerts []alertmanager.Alert) discord.Out {
	DO := discord.Out{}

	if amo.CommonAnnotations.Summary != "" {
		DO.Content = fmt.Sprintf(" === %s === \n", amo.CommonAnnotations.Summary)
	}

	RichEmbed := discord.Embed{
		Title:       fmt.Sprintf("[%s: %d] %s", strings.ToUpper(status), len(alerts), amo.CommonLabels.Alertname),
		Description: amo.CommonAnnotations.Summary,
		Color:       discord.ColorGrey,
		Fields:      []discord.EmbedField{},
	}

	switch status {
	case alertmanager.StatusFiring:
		RichEmbed.Color = discord.ColorRed
	case alertmanager.StatusResolved:
		RichEmbed.Color = discord.ColorGreen
	}

	for _, alert := range alerts {
		var details strings.Builder
		details.WriteString("===Annotations===\n")

		// sort into alphabetical order
		annotationKeys := make([]string, 0, len(alert.Annotations))
		for key := range alert.Annotations {
			annotationKeys = append(annotationKeys, key)
		}
		sort.Strings(annotationKeys)

		for _, key := range annotationKeys {
			details.WriteString(fmt.Sprintf("'%s': '%s'\n", key, alert.Annotations[key]))
		}

		details.WriteString("===Labels===\n")

		// sort into alphabetical order
		labelKeys := make([]string, 0, len(alert.Labels))
		for key := range alert.Labels {
			labelKeys = append(labelKeys, key)
		}
		sort.Strings(labelKeys)

		for _, key := range labelKeys {
			details.WriteString(fmt.Sprintf("'%s': '%s'\n", key, alert.Labels[key]))
		}

		RichEmbed.Fields = append(RichEmbed.Fields, discord.EmbedField{
			Name:  "Alert details",
			Value: details.String(),
		})
	}

	DO.Embeds = []discord.Embed{RichEmbed}

	return DO
}
