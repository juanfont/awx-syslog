package awxsyslog

import (
	"encoding/json"
	"time"
)

// Common schema fields present in all log messages
type CommonLogFields struct {
	ClusterHostID string    `json:"cluster_host_id"` // Unique identifier of the host within the controller cluster
	Level         string    `json:"level"`           // Standard python log level (INFO, DEBUG, ERROR, etc.)
	LoggerName    string    `json:"logger_name"`     // Name of the logger (e.g., "activity_stream", "job_events")
	Timestamp     time.Time `json:"@timestamp"`      // Time of log
	Path          string    `json:"path"`            // File path in code where the log was generated
}

// ActivityStreamLog represents activity stream log messages
type ActivityStreamLog struct {
	CommonLogFields
	Actor     string                 `json:"actor"`     // Username of the user who took the action
	Changes   map[string]interface{} `json:"changes"`   // JSON summary of what fields changed, old/new values
	Operation string                 `json:"operation"` // Basic category of change (e.g., "associate")
	Object1   map[string]interface{} `json:"object1"`   // Information about the primary object being operated on
	Object2   map[string]interface{} `json:"object2"`   // If applicable, the second object involved in the action
}

// JobEventLog represents job event log messages
type JobEventLog struct {
	CommonLogFields
	EventHost string                 `json:"event_host"` // Host field from job_event model (renamed to avoid conflicts)
	EventData map[string]interface{} `json:"event_data"` // Sub-dictionary with different fields depending on Ansible event
	JobID     int                    `json:"job_id,omitempty"`
	TaskName  string                 `json:"task_name,omitempty"`
	PlayName  string                 `json:"play_name,omitempty"`
}

// SystemTrackingLog represents system tracking/fact data
type SystemTrackingLog struct {
	CommonLogFields
	Services    map[string]interface{} `json:"services,omitempty"` // For services scans (periods replaced with "_")
	Package     map[string]interface{} `json:"package,omitempty"`  // For package scans
	Files       map[string]interface{} `json:"files,omitempty"`    // For file scans
	Host        string                 `json:"host"`               // Name of host scan applies to
	InventoryID int                    `json:"inventory_id"`       // Inventory ID the host is inside of
}

// AWXLog represents generic automation controller server logs
type AWXLog struct {
	CommonLogFields
	Msg       string `json:"msg"`                 // Log message content
	Traceback string `json:"traceback,omitempty"` // Error traceback if present
}

// parseAWXLog parses incoming JSON and extracts logger type and common fields
func parseAWXLog(data []byte) (string, CommonLogFields, map[string]interface{}, error) {
	// Parse into a generic map first
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return "", CommonLogFields{}, nil, err
	}

	// Extract common fields
	var commonFields CommonLogFields
	if v, ok := raw["cluster_host_id"].(string); ok {
		commonFields.ClusterHostID = v
	}
	if v, ok := raw["level"].(string); ok {
		commonFields.Level = v
	}
	if v, ok := raw["logger_name"].(string); ok {
		commonFields.LoggerName = v
	}
	if v, ok := raw["@timestamp"].(string); ok {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			commonFields.Timestamp = t
		}
	}
	if v, ok := raw["path"].(string); ok {
		commonFields.Path = v
	}

	// Store remaining fields in remaining data map
	remainingData := make(map[string]interface{})
	for k, v := range raw {
		switch k {
		case "cluster_host_id", "level", "logger_name", "@timestamp", "path":
			// Skip common fields already processed
		default:
			remainingData[k] = v
		}
	}

	return commonFields.LoggerName, commonFields, remainingData, nil
}
