package housekeeper

import (
	"log"
	"time"

	"project-hermes/internal/db"
)

// Start begins the daily background cleanup task
func Start() {
	log.Println("Initializing Background Housekeeper (Runs every 24h)...")

	// Run once immediately on startup, then tick every 24 hours
	go func() {
		for {
			RunCleanup()
			time.Sleep(24 * time.Hour)
		}
	}()
}

// RunCleanup deletes old graph data from the metrics table
func RunCleanup() error {
	// 1. Fetch retention setting (default to 90 days if missing)
	var daysStr string
	err := db.DB.QueryRow("SELECT value FROM settings WHERE key = 'data_retention_days'").Scan(&daysStr)
	if err != nil {
		daysStr = "90" // Fallback safety net
	}

	// 2. Execute the database purge
	// SQLite syntax to delete records older than X days
	query := `DELETE FROM metrics WHERE timestamp < datetime('now', '-' || ? || ' days')`
	result, err := db.DB.Exec(query, daysStr)
	if err != nil {
		log.Printf("[Housekeeper Error] Failed to purge old metrics: %v", err)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		log.Printf("[Housekeeper] 🧹 Successfully purged %d old graph data points.", rowsAffected)
	}
	
	return nil
}