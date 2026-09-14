package models

import "time"

// Host represents a monitored network device (Mikrotik, Fortigate, etc.)
type Host struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	IP        string `json:"ip"`
	Community string `json:"community"`
	Enabled   bool   `json:"enabled"`
}

// Interface represents a physical or virtual port on a host
type Interface struct {
	ID         int64  `json:"id"`
	HostID     int64  `json:"host_id"`
	IndexNum   int    `json:"index_num"`
	Name       string `json:"name"`
	Monitoring bool   `json:"monitoring"` // True if we are actively recording this port's speed
}

// Metric represents a single speed recording snapshot
type Metric struct {
	ID          int64     `json:"id"`
	InterfaceID int64     `json:"interface_id"`
	Download    float64   `json:"download_mbps"`
	Upload      float64   `json:"upload_mbps"`
	Timestamp   time.Time `json:"timestamp"`
}

// AlertRule matches the exact structure we built in the Vue frontend
type AlertRule struct {
	ID                    int    `json:"id"`
	HostID                int    `json:"host_id"`
	Ports                 []string `json:"ports"`
	Direction             string `json:"direction"`
	AlertThresholdMbps    int    `json:"alert_threshold_mbps"`
	AlertDurationMins     int    `json:"alert_duration_mins"`
	AlertTemplate         string `json:"alert_template"`
	RecoveryThresholdMbps int    `json:"recovery_threshold_mbps"`
	RecoveryDurationMins  int    `json:"recovery_duration_mins"`
	RecoveryTemplate      string `json:"recovery_template"`
	Enabled               bool   `json:"enabled"`
}