package awxsyslog

type SyslogConfig struct {
	ServerAddr string
	Protocol   string // tcp, udp, tls
}

type Config struct {
	ListenAddr    string
	LogLevel      string // debug, info, warn, error, critical
	HostnameField string // The hostname field to use for the syslog message
	Syslog        SyslogConfig
}
