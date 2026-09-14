package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// InitDB initializes the SQLite database and creates tables
func InitDB(filepath string) {
	var err error
	DB, err = sql.Open("sqlite3", filepath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Enable foreign keys in SQLite
	_, err = DB.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		log.Fatalf("Failed to enable foreign keys: %v", err)
	}

	createTables()
}

func createTables() {
	// 1. Hosts Table
	hostsTable := `
	CREATE TABLE IF NOT EXISTS hosts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		ip TEXT NOT NULL UNIQUE,
		community TEXT NOT NULL,
		enabled INTEGER DEFAULT 1
	);`

	// 2. Interfaces Table
	interfacesTable := `
	CREATE TABLE IF NOT EXISTS interfaces (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		host_id INTEGER,
		index_num INTEGER NOT NULL,
		name TEXT NOT NULL,
		monitoring INTEGER DEFAULT 0,
		FOREIGN KEY(host_id) REFERENCES hosts(id) ON DELETE CASCADE,
		UNIQUE(host_id, index_num)
	);`

	// 3. Metrics Table
	metricsTable := `
	CREATE TABLE IF NOT EXISTS metrics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		interface_id INTEGER,
		download_mbps REAL NOT NULL,
		upload_mbps REAL NOT NULL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(interface_id) REFERENCES interfaces(id) ON DELETE CASCADE
	);`

	// 4. Settings Table (For Telegram & Uptime Kuma style configs)
	settingsTable := `
	CREATE TABLE IF NOT EXISTS settings (
		key TEXT PRIMARY KEY,
		value TEXT
	);`

	// 5. Alert Rules Table (NEW: Connects hosts to their specific traffic rules)
	alertRulesTable := `
	CREATE TABLE IF NOT EXISTS alert_rules (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		host_id INTEGER NOT NULL,
		port TEXT NOT NULL,
		direction TEXT NOT NULL, 
		alert_threshold_mbps INTEGER NOT NULL,
		alert_duration_mins INTEGER NOT NULL,
		alert_template TEXT,
		recovery_threshold_mbps INTEGER NOT NULL,
		recovery_duration_mins INTEGER NOT NULL,
		recovery_template TEXT,
		enabled INTEGER DEFAULT 1,
		FOREIGN KEY(host_id) REFERENCES hosts(id) ON DELETE CASCADE
	);`

	// Execute tables creation
	_, err := DB.Exec(hostsTable)
	if err != nil {
		log.Fatalf("Error creating hosts table: %v", err)
	}

	_, err = DB.Exec(interfacesTable)
	if err != nil {
		log.Fatalf("Error creating interfaces table: %v", err)
	}

	_, err = DB.Exec(metricsTable)
	if err != nil {
		log.Fatalf("Error creating metrics table: %v", err)
	}

	_, err = DB.Exec(settingsTable)
	if err != nil {
		log.Fatalf("Error creating settings table: %v", err)
	}

	// Run our new alert rules table execution block
	_, err = DB.Exec(alertRulesTable)
	if err != nil {
		log.Fatalf("Error creating alert_rules table: %v", err)
	}

	log.Println("Database initialized successfully with all tables.")
}