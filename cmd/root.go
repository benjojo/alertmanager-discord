package cmd

import (
	"os"
	"strings"
	"time"

	"github.com/specklesystems/alertmanager-discord/pkg/flags"
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
	defaultLogLevel              = "info"
)

var (
	configurationFilePath     string
	webhookURL                string
	listenAddress             string
	logLevel                  string
	maximumBackoffTimeSeconds int
)

func init() {
	defineConfigurationVariable(&configurationFilePath, rootCmd.Flags().StringVarP, flags.ConfigurationPathFlagKey, "c", defaultConfigurationPath, "Path to the configuration file.")
	defineConfigurationVariable(&webhookURL, rootCmd.Flags().StringVarP, flags.DiscordWebhookUrlFlagKey, "d", "", "Url to the Discord webhook API endpoint.")
	defineConfigurationVariable(&listenAddress, rootCmd.Flags().StringVarP, flags.ListenAddressFlagKey, "l", server.DefaultListenAddress, "The address (host:port) which the server will attempt to bind to and listen on.")
	defineConfigurationVariable(&logLevel, rootCmd.Flags().StringVarP, flags.LogLevelFlagKey, "", defaultLogLevel, "The minimum level of logging to be produced by the pod. Acceptable values, in ascending order, are 'trace', 'debug', 'info', 'warn', 'error', 'fatal', 'panic', or 'disabled'.")
	defineConfigurationVariable(&maximumBackoffTimeSeconds, rootCmd.Flags().IntVarP, flags.MaxBackoffTimeSecondsFlagKey, "", defaultMaxBackoffTimeSeconds, "The maximum elapsed duration (expressed as an integer number of seconds) to allow the Discord client to continue retrying to send messages to the Discord API.")
}

func defineConfigurationVariable[K int | string](variable *K, flagParser func(*K, string, string, K, string), flagKey string, shorthand string, defaultValue K, description string) {
	viper.SetDefault(flagKey, defaultValue)
	viper.BindEnv(flagKey, strings.ToUpper(flagKey))
	flagParser(variable, flagKey, shorthand, defaultValue, description)
	viper.BindPFlag(flagKey, rootCmd.Flags().Lookup(flagKey))
}

var rootCmd = &cobra.Command{
	Use:     "alertmanager-discord",
	Version: version.Version,
	Short:   "Forwards AlertManager alerts to Discord.",
	Long: `A simple web server that accepts AlertManager webhooks,
translates the data to match Discord's message specifications,
and forwards that to Discord's message API endpoint.`,
	Run: func(cmd *cobra.Command, args []string) {
		zerolog.TimeFieldFormat = time.RFC3339
		zerolog.SetGlobalLevel(zerolog.InfoLevel)

		// these log messages are generated before the log level is set
		log.Debug().Msgf("Attempting to read from configuration file path: ('%s')", configurationFilePath)
		viper.SetConfigFile(configurationFilePath)
		if err := viper.ReadInConfig(); err != nil {
			log.Info().Err(err).Msgf("Unable to read configuration file at path ('%s'). Attempting to parse command line arguments or environment variables, the command line argument has higher order of precedence.", configurationFilePath)
		}

		if viper.GetString(flags.DiscordWebhookUrlFlagKey) != "" {
			webhookURL = viper.GetString(flags.DiscordWebhookUrlFlagKey)
		}
		if viper.GetString(flags.ListenAddressFlagKey) != "" {
			listenAddress = viper.GetString(flags.ListenAddressFlagKey)
		}

		setGlobalLogLevel(viper.GetString(flags.LogLevelFlagKey))

		if viper.GetString(flags.MaxBackoffTimeSecondsFlagKey) != "" {
			maximumBackoffTimeSeconds = viper.GetInt(flags.MaxBackoffTimeSecondsFlagKey)
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

func setGlobalLogLevel(logLevel string) {
	switch logLevel {
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	case "disabled":
		zerolog.SetGlobalLevel(zerolog.Disabled)
	default:
		break
	}
}
