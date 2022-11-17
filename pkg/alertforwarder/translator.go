package alertforwarder

import (
	"fmt"
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
		for key, val := range alert.Annotations {
			details.WriteString(fmt.Sprintf("'%s': '%s'\n", key, val))
		}
		details.WriteString("===Labels===\n")
		for key, val := range alert.Labels {
			details.WriteString(fmt.Sprintf("'%s': '%s'\n", key, val))
		}

		RichEmbed.Fields = append(RichEmbed.Fields, discord.EmbedField{
			Name:  "Alert details",
			Value: details.String(),
		})
	}

	DO.Embeds = []discord.Embed{RichEmbed}

	return DO
}
