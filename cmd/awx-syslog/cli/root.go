package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string = ""

func init() {
	if len(os.Args) > 1 &&
		(os.Args[1] == "version") {
		return
	}

	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().
		StringVarP(&cfgFile, "config", "c", "", "config file (default is /etc/awx-syslog/config.yaml)")
}

func initConfig() {
	if cfgFile == "" {
		cfgFile = os.Getenv("AWX_SYSLOG_CONFIG")
	}

	if cfgFile != "" {
		err := loadConfig(cfgFile, true)
		if err != nil {
			log.Fatal().Caller().Err(err).Msgf("Error loading config file %s", cfgFile)
		}
	} else {
		err := loadConfig("", false)
		if err != nil {
			log.Fatal().Caller().Err(err).Msgf("Error loading config")
		}
	}

	viper.SetEnvPrefix("awx-syslog")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	logLevelStr := viper.GetString("log_level")
	if logLevelStr == "" {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	} else {
		logLevel, err := zerolog.ParseLevel(logLevelStr)
		if err != nil {
			logLevel = zerolog.InfoLevel
		}
		zerolog.SetGlobalLevel(logLevel)
	}
}

func loadConfig(path string, isFile bool) error {
	if isFile {
		viper.SetConfigFile(path)
	} else {
		viper.SetConfigName("config")
		if path == "" {
			viper.AddConfigPath("/etc/awx-syslog")
			viper.AddConfigPath("$HOME/.awx-syslog")
			viper.AddConfigPath(".")
		} else {
			// For testing
			viper.AddConfigPath(path)
		}
	}

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	return nil
}

var rootCmd = &cobra.Command{
	Use:   "awx-syslog",
	Short: "awx-syslog - A syslog forwarder for Ansible AWX logs",
	Long:  `awx-syslog takes care of receiving JSON via HTTP and forwarding it to a syslog server`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
