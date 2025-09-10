package awxsyslog

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/influxdata/go-syslog/rfc5424"
)

// awxEventToSyslogMessage converts any AWX log event to a syslog message
func (a *App) awxEventToSyslogMessage(loggerType string, data map[string]interface{}, commonFields CommonLogFields) rfc5424.SyslogMessage {
	msg := rfc5424.SyslogMessage{}

	// Set facility based on log type
	facility := 13 // audit log
	priority := facility*8 + logLevelStringToSyslog(commonFields.Level)

	msg.SetPriority(uint8(priority))
	msg.SetVersion(1)

	// Set timestamp from the log event
	msg.SetTimestamp(commonFields.Timestamp.Format(time.RFC3339))
	msg.SetAppname(fmt.Sprintf("awx-controller-%s", commonFields.ClusterHostID))
	msg.SetHostname(a.cfg.HostnameField)

	// Add common structured data
	msg.SetParameter("awx_common", "logger_name", loggerType)
	msg.SetParameter("awx_common", "cluster_host_id", commonFields.ClusterHostID)
	msg.SetParameter("awx_common", "path", commonFields.Path)

	// Add type-specific structured data using reflection
	addStructuredDataForLogType(&msg, loggerType, data)

	// Set message ID and content based on log type
	messageID, message := getMessageForLogType(loggerType, data, commonFields)
	msg.SetMsgID(messageID)
	msg.SetMessage(message)

	return msg
}

// logLevelStringToSyslog converts AWX log level to syslog severity
func logLevelStringToSyslog(level string) int {
	switch strings.ToUpper(level) {
	case "CRITICAL":
		return 2 // Critical
	case "ERROR":
		return 3 // Error
	case "WARNING", "WARN":
		return 4 // Warning
	case "INFO":
		return 6 // Informational
	case "DEBUG":
		return 7 // Debug
	default:
		return 6 // Default to Info
	}
}

// structToMap converts any struct to a map[string]string using reflection
func structToMap(v interface{}) map[string]string {
	fieldMap := map[string]string{}
	reflectValue := reflect.ValueOf(v)

	// Handle if passed a pointer to struct
	if reflectValue.Kind() == reflect.Ptr {
		reflectValue = reflectValue.Elem()
	}

	for i := 0; i < reflectValue.NumField(); i++ {
		fieldName := reflectValue.Type().Field(i).Name
		fieldValue := reflectValue.Field(i).Interface()

		// Handle pointer fields
		if reflectValue.Field(i).Kind() == reflect.Ptr && !reflectValue.Field(i).IsNil() {
			fieldValue = reflectValue.Field(i).Elem().Interface()
		}

		// Skip nil pointers
		if reflectValue.Field(i).Kind() != reflect.Ptr || !reflectValue.Field(i).IsNil() {
			fieldMap[fieldName] = fmt.Sprintf("%v", fieldValue)
		}
	}
	return fieldMap
}

// awxLogToStructuredData adds structured data using reflection
func awxLogToStructuredData(msg *rfc5424.SyslogMessage, logType string, logStruct interface{}) {
	fieldMap := structToMap(logStruct)
	for k, v := range fieldMap {
		// Skip common fields (already handled separately)
		if k == "ClusterHostID" || k == "Level" || k == "LoggerName" || k == "Timestamp" || k == "Path" {
			continue
		}
		// Convert field name to snake_case for syslog
		fieldName := strings.ToLower(k)
		msg.SetParameter(logType, fieldName, v)
	}
}

// addStructuredDataForLogType adds structured data elements based on log type using reflection
func addStructuredDataForLogType(msg *rfc5424.SyslogMessage, loggerType string, data map[string]interface{}) {
	// Create appropriate struct and unmarshal data into it
	switch loggerType {
	case "activity_stream":
		var activityLog ActivityStreamLog
		if jsonData, err := json.Marshal(data); err == nil {
			if err := json.Unmarshal(jsonData, &activityLog); err == nil {
				awxLogToStructuredData(msg, "activity_stream", activityLog)
			}
		}
	case "job_events":
		var jobLog JobEventLog
		if jsonData, err := json.Marshal(data); err == nil {
			if err := json.Unmarshal(jsonData, &jobLog); err == nil {
				awxLogToStructuredData(msg, "job_events", jobLog)
			}
		}
	case "system_tracking":
		var systemLog SystemTrackingLog
		if jsonData, err := json.Marshal(data); err == nil {
			if err := json.Unmarshal(jsonData, &systemLog); err == nil {
				awxLogToStructuredData(msg, "system_tracking", systemLog)
			}
		}
	case "awx":
		var awxLog AWXLog
		if jsonData, err := json.Marshal(data); err == nil {
			if err := json.Unmarshal(jsonData, &awxLog); err == nil {
				awxLogToStructuredData(msg, "awx_log", awxLog)
			}
		}
	default:
		// For unknown log types, add raw data as structured data
		for k, v := range data {
			msg.SetParameter("unknown_log", k, fmt.Sprintf("%v", v))
		}
	}
}

// getMessageForLogType generates appropriate message ID and content based on log type
func getMessageForLogType(loggerType string, data map[string]interface{}, commonFields CommonLogFields) (string, string) {
	switch loggerType {
	case "activity_stream":
		return getActivityStreamMessage(data)
	case "job_events":
		return getJobEventMessage(data)
	case "system_tracking":
		return getSystemTrackingMessage(data)
	case "awx":
		return getAWXLogMessage(data)
	default:
		return "UNKNOWN", fmt.Sprintf("Unknown log type: %s", loggerType)
	}
}

// getActivityStreamMessage generates message for activity stream logs
func getActivityStreamMessage(data map[string]interface{}) (string, string) {
	actor := getStringFromData(data, "actor", "unknown")
	operation := getStringFromData(data, "operation", "unknown")

	var objectInfo string
	if obj1, ok := data["object1"].(map[string]interface{}); ok {
		objType := getStringFromData(obj1, "type", "object")
		objName := getStringFromData(obj1, "name", "unnamed")
		objectInfo = fmt.Sprintf("%s '%s'", objType, objName)
	} else {
		objectInfo = "object"
	}

	messageID := "ACTIVITY_STREAM"
	message := fmt.Sprintf("User %s performed %s on %s", actor, operation, objectInfo)

	return messageID, message
}

// getJobEventMessage generates message for job event logs
func getJobEventMessage(data map[string]interface{}) (string, string) {
	eventHost := getStringFromData(data, "event_host", "unknown")
	taskName := getStringFromData(data, "task_name", "")

	messageID := "JOB_EVENT"
	var message string

	if taskName != "" {
		message = fmt.Sprintf("Job event on host %s: task '%s'", eventHost, taskName)
	} else {
		message = fmt.Sprintf("Job event on host %s", eventHost)
	}

	return messageID, message
}

// getSystemTrackingMessage generates message for system tracking logs
func getSystemTrackingMessage(data map[string]interface{}) (string, string) {
	host := getStringFromData(data, "host", "unknown")

	messageID := "SYSTEM_TRACKING"
	var scanType string

	if _, hasServices := data["services"]; hasServices {
		scanType = "services"
	} else if _, hasPackage := data["package"]; hasPackage {
		scanType = "package"
	} else if _, hasFiles := data["files"]; hasFiles {
		scanType = "files"
	} else {
		scanType = "unknown"
	}

	message := fmt.Sprintf("System tracking scan (%s) for host %s", scanType, host)

	return messageID, message
}

// getAWXLogMessage generates message for generic AWX logs
func getAWXLogMessage(data map[string]interface{}) (string, string) {
	msg := getStringFromData(data, "msg", "No message")

	messageID := "AWX_LOG"

	// Check if it's an error log with traceback
	if traceback, ok := data["traceback"].(string); ok && traceback != "" {
		messageID = "AWX_ERROR"
		return messageID, fmt.Sprintf("AWX Error: %s", msg)
	}

	return messageID, msg
}

// getStringFromData safely extracts string value from data map
func getStringFromData(data map[string]interface{}, key, defaultValue string) string {
	if value, ok := data[key].(string); ok {
		return value
	}
	return defaultValue
}
