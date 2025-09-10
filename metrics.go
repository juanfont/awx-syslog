package awxsyslog

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	prometheusNamespace = "awx_syslog"
)

// audit_log_events received
var awxSyslogEventsReceived = promauto.NewCounter(prometheus.CounterOpts{
	Namespace: prometheusNamespace,
	Name:      "awx_syslog_logs_received",
	Help:      "Number of AWX syslog events received",
})
