package alertforwarder

import (
	"fmt"
	"sort"
	"strings"

	"github.com/specklesystems/alertmanager-discord/pkg/alertmanager"
	"github.com/specklesystems/alertmanager-discord/pkg/discord"
)

const (
	keySummary         = "summary"
	keyEnvironmentType = "source_environment_type"
	keyEnvironmentName = "source_environment_name"
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
		fieldName := fmt.Sprintf("[%s/%s] Alert details", alert.Labels[keyEnvironmentType], alert.Labels[keyEnvironmentName])
		if summary, ok := alert.Annotations[keySummary]; ok {
			fieldName = fmt.Sprintf("[%s/%s] %s", alert.Labels[keyEnvironmentType], alert.Labels[keyEnvironmentName], summary)
		}

		var details strings.Builder
		details.WriteString("Annotations:\n")

		// sort into alphabetical order
		annotationKeys := make([]string, 0, len(alert.Annotations))
		for key := range alert.Annotations {
			annotationKeys = append(annotationKeys, key)
		}
		sort.Strings(annotationKeys)

		for _, key := range annotationKeys {
			if key == keySummary {
				// if there is a summary, it is already the field name so no need to repeat it
				continue
			}
			details.WriteString(fmt.Sprintf("\t%s: %s\n", key, alert.Annotations[key]))
		}

		details.WriteString("Labels:\n")

		// sort into alphabetical order
		labelKeys := make([]string, 0, len(alert.Labels))
		for key := range alert.Labels {
			labelKeys = append(labelKeys, key)
		}
		sort.Strings(labelKeys)

		for _, key := range labelKeys {
			if key == keyEnvironmentName || key == keyEnvironmentType {
				// if these keys exist, we have already added them to the field name
				continue
			}
			details.WriteString(fmt.Sprintf("\t%s: %s\n", key, alert.Labels[key]))
		}

		RichEmbed.Fields = append(RichEmbed.Fields, discord.EmbedField{
			Name:  fieldName,
			Value: details.String(),
		})
	}

	DO.Embeds = []discord.Embed{RichEmbed}

	return DO
}
