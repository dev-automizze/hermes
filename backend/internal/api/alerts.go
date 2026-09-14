package api

import (
	"encoding/json"
	"log"
	"net/http"
	"project-hermes/internal/db"
	"project-hermes/internal/models"

	"github.com/gin-gonic/gin"
)

// GetAlertRules fetches all rules from SQLite and sends them to Vue
func GetAlertRules(c *gin.Context) {
	rows, err := db.DB.Query(`
		SELECT id, host_id, port, direction, alert_threshold_mbps, alert_duration_mins, 
			   alert_template, recovery_threshold_mbps, recovery_duration_mins, recovery_template, enabled 
		FROM alert_rules`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch alert rules"})
		return
	}
	defer rows.Close()

	var rules []models.AlertRule
	for rows.Next() {
		var r models.AlertRule
		var enabledInt int
		var portsJSON string // We will extract the JSON string from the DB here
		
		err := rows.Scan(
			&r.ID, &r.HostID, &portsJSON, &r.Direction, &r.AlertThresholdMbps, &r.AlertDurationMins,
			&r.AlertTemplate, &r.RecoveryThresholdMbps, &r.RecoveryDurationMins, &r.RecoveryTemplate, &enabledInt,
		)
		if err != nil {
			log.Printf("Row scan error: %v", err)
			continue
		}
		
		// Unmarshal the string back into the Go string slice
		json.Unmarshal([]byte(portsJSON), &r.Ports)
		
		r.Enabled = enabledInt == 1
		rules = append(rules, r)
	}

	if rules == nil {
		rules = []models.AlertRule{}
	}

	c.JSON(http.StatusOK, rules)
}

// SaveAlertRules takes the JSON array from Vue and saves it to SQLite
func SaveAlertRules(c *gin.Context) {
	var rules []models.AlertRule
	
	if err := c.ShouldBindJSON(&rules); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	_, err = tx.Exec("DELETE FROM alert_rules")
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear old rules"})
		return
	}

	stmt, err := tx.Prepare(`
		INSERT INTO alert_rules (
			host_id, port, direction, alert_threshold_mbps, alert_duration_mins, 
			alert_template, recovery_threshold_mbps, recovery_duration_mins, recovery_template, enabled
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare statement"})
		return
	}
	defer stmt.Close()

	for _, r := range rules {
		enabledInt := 0
		if r.Enabled {
			enabledInt = 1
		}

		// Marshal the array back into a JSON string for SQLite
		portsJSON, _ := json.Marshal(r.Ports)

		_, err = stmt.Exec(
			r.HostID, string(portsJSON), r.Direction, r.AlertThresholdMbps, r.AlertDurationMins,
			r.AlertTemplate, r.RecoveryThresholdMbps, r.RecoveryDurationMins, r.RecoveryTemplate, enabledInt,
		)
		if err != nil {
			tx.Rollback()
			log.Printf("Failed to insert rule: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save rule"})
			return
		}
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "Alert rules saved successfully"})
}