package cli

import (
	awxsyslog "github.com/juanfont/awx-syslog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(serveCmd)
}

var serveCmd = &cobra.Command{
	Use:     "serve",
	Short:   "Serve the AWX JSON->Syslog server",
	Aliases: []string{"s"},
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := getAWXSyslogConfig()
		if err != nil {
			log.Fatal().Err(err)
		}

		app, err := awxsyslog.NewApp(cfg)
		if err != nil {
			log.Fatal().Err(err).Msg("Could not create AWX syslog server")
		}

		err = app.Serve()
		if err != nil {
			log.Fatal().Err(err).Msg("Could not serve AWX syslog server")
		}
	},
}

func getAWXSyslogConfig() (*awxsyslog.Config, error) {
	cfg := &awxsyslog.Config{
		ListenAddr:    viper.GetString("listen_addr"),
		LogLevel:      viper.GetString("log_level"),
		HostnameField: viper.GetString("hostname_field"),
		Syslog: awxsyslog.SyslogConfig{
			ServerAddr: viper.GetString("syslog.server_addr"),
			Protocol:   viper.GetString("syslog.protocol"),
		},
	}

	return cfg, nil
}
