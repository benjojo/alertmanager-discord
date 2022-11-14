package cmd

import (
	"os"
	"strings"
	"time"

	. "github.com/specklesystems/alertmanager-discord/pkg/flags"
	"github.com/specklesystems/alertmanager-discord/pkg/server"
	"github.com/specklesystems/alertmanager-discord/pkg/version"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	defaultConfigurationPath     = "/etc/alertmanager-discord/config.yaml"
	defaultMaxBackoffTimeSeconds = 10
)

var (
	configurationFilePath     string
	webhookURL                string
	listenAddress             string
	maximumBackoffTimeSeconds int
)

func init() {
	viper.SetDefault(ConfigurationPathFlagKey, defaultConfigurationPath)

	viper.BindEnv(ConfigurationPathFlagKey, strings.ToUpper(ConfigurationPathFlagKey))
	rootCmd.Flags().StringVarP(&configurationFilePath, ConfigurationPathFlagKey, "c", defaultConfigurationPath, "Path to the configuration file.")
	viper.BindPFlag(ConfigurationPathFlagKey, rootCmd.Flags().Lookup(ConfigurationPathFlagKey))

	viper.BindEnv(DiscordWebhookUrlFlagKey, strings.ToUpper(DiscordWebhookUrlFlagKey))
	rootCmd.Flags().StringVarP(&webhookURL, DiscordWebhookUrlFlagKey, "d", "", "Url to the Discord webhook API endpoint.")
	viper.BindPFlag(DiscordWebhookUrlFlagKey, rootCmd.Flags().Lookup(DiscordWebhookUrlFlagKey))

	viper.SetDefault(ListenAddressFlagKey, server.DefaultListenAddress)
	viper.BindEnv(ListenAddressFlagKey, strings.ToUpper(ListenAddressFlagKey))
	rootCmd.Flags().StringVarP(&listenAddress, ListenAddressFlagKey, "l", "", "The address (host:port) which the server will attempt to bind to and listen on.")
	viper.BindPFlag(ListenAddressFlagKey, rootCmd.Flags().Lookup(ListenAddressFlagKey))

	viper.SetDefault(MaxBackoffTimeSecondsFlagKey, defaultMaxBackoffTimeSeconds)
	viper.BindEnv(MaxBackoffTimeSecondsFlagKey, strings.ToUpper(MaxBackoffTimeSecondsFlagKey))
	rootCmd.Flags().IntVarP(&maximumBackoffTimeSeconds, MaxBackoffTimeSecondsFlagKey, "", defaultMaxBackoffTimeSeconds, "The maximum elapsed duration (expressed as an integer number of seconds) to allow the Discord client to continue retrying to send messages to the Discord API.")
	viper.BindPFlag(MaxBackoffTimeSecondsFlagKey, rootCmd.Flags().Lookup(MaxBackoffTimeSecondsFlagKey))
}

var rootCmd = &cobra.Command{
	Use:     "alertmanager-discord",
	Version: version.Version,
	Short:   "Forwards AlertManager alerts to Discord.",
	Long: `A simple web server that accepts AlertManager webhooks,
translates the data to match Discord's message specifications,
and forwards that to Discord's message API endpoint.`,
	Run: func(cmd *cobra.Command, args []string) {
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

		log.Debug().Msgf("Attempting to read from configuration file path: ('%s')", configurationFilePath)
		viper.SetConfigFile(configurationFilePath)
		if err := viper.ReadInConfig(); err != nil {
			log.Info().Err(err).Msgf("Unable to read configuration file at path ('%s'). Attempting to parse command line arguments or environment variables, the command line argument has higher order of precedence.", configurationFilePath)
		}

		if viper.GetString(DiscordWebhookUrlFlagKey) != "" {
			webhookURL = viper.GetString(DiscordWebhookUrlFlagKey)
		}
		if viper.GetString(ListenAddressFlagKey) != "" {
			listenAddress = viper.GetString(ListenAddressFlagKey)
		}
		if viper.GetString(MaxBackoffTimeSecondsFlagKey) != "" {
			maximumBackoffTimeSeconds = viper.GetInt(MaxBackoffTimeSecondsFlagKey)
		}

		amds := server.AlertManagerDiscordServer{
			MaximumBackoffTimeSeconds: time.Duration(maximumBackoffTimeSeconds) * time.Second,
		}
		stopCh, err := amds.ListenAndServe(webhookURL, listenAddress)
		defer func() {
			if err = amds.Shutdown(); err != nil {
				log.Fatal().Err(err).Msg("Error while shutting down server.")
			}
		}()
		if err != nil {
			log.Error().Err(err).Msg("Error in AlertManager-Discord server")
			close(stopCh)
		}

		// Waits here for SIGINT (kill -2) or for channel to be closed (which can occur if there is an error in the server)
		<-stopCh
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Error().Err(err).Msg("Error when executing command. Exiting program...")
		os.Exit(1)
	}
}
