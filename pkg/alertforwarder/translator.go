package alertforwarder

import (
	"fmt"
	"strings"

	"github.com/specklesystems/alertmanager-discord/pkg/alertmanager"
	"github.com/specklesystems/alertmanager-discord/pkg/discord"
)

func TranslateAlertManagerToDiscord(status string, amo *alertmanager.Out, alerts []alertmanager.Alert) discord.Out {
	DO := discord.Out{}

	RichEmbed := discord.Embed{
		Title:       fmt.Sprintf("[%s:%d] %s", strings.ToUpper(status), len(alerts), amo.CommonLabels.Alertname),
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

	if amo.CommonAnnotations.Summary != "" {
		DO.Content = fmt.Sprintf(" === %s === \n", amo.CommonAnnotations.Summary)
	}

	for _, alert := range alerts {
		realname := alert.Labels["instance"]
		if strings.Contains(realname, "localhost") && alert.Labels["exported_instance"] != "" {
			realname = alert.Labels["exported_instance"]
		}

		RichEmbed.Fields = append(RichEmbed.Fields, discord.EmbedField{
			Name:  fmt.Sprintf("[%s]: %s on %s", strings.ToUpper(status), alert.Labels["alertname"], realname),
			Value: alert.Annotations.Description,
		})
	}

	DO.Embeds = []discord.Embed{RichEmbed}

	return DO
}
