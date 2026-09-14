package main

import (
	"database/sql"
	"fmt"
	"log"
	"project-hermes/internal/api"
	"project-hermes/internal/db"
	"project-hermes/internal/housekeeper"
	"project-hermes/internal/models"
	"project-hermes/internal/poller"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	// 1. Initialize the SQLite Database
	db.InitDB("data/hermes.db")
	defer db.DB.Close()

	// 2. Automatically Prep Auth Tables & Seed Default Admin User
	ensureAuthSystem()

	fmt.Println("\nInitialization successful! Hermes database is prepped and ready.")
	fmt.Println("-------------------------------------------------------------------")

	// 3. START THE BACKGROUND POLLING ENGINE 
	poller.Start(60 * time.Second)

	// 4. START THE BACKGROUND HOUSEKEEPER (Janitor)
	housekeeper.Start() 

	// 5. START THE WEB API SERVER
	fmt.Println("Starting Hermes API Server on http://localhost:8014...")
	router := api.SetupRouter()
	if err := router.Run(":8014"); err != nil {
		log.Fatalf("Failed to start API server: %v", err)
	}
}

// ensureAuthSystem automatically bootstraps our login security infrastructure
func ensureAuthSystem() {
	// Create Users Table
	_, err := db.DB.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL
		);`)
	if err != nil {
		log.Fatalf("Failed to build users security layer: %v", err)
	}

	// Create Sessions Table
	_, err = db.DB.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
			token TEXT PRIMARY KEY,
			user_id INTEGER NOT NULL,
			expires_at DATETIME NOT NULL,
			FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
		);`)
	if err != nil {
		log.Fatalf("Failed to build sessions layer: %v", err)
	}

	// Check if any administrator accounts exist
	var count int
	err = db.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		log.Fatalf("Failed checking database user state: %v", err)
	}

	// Seed the default system account if completely empty
	if count == 0 {
		defaultPassword := "Hermes" //default password for first-time setup (should be changed immediately after login)
		hashed, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("Failed hashing default administration credential: %v", err)
		}

		_, err = db.DB.Exec("INSERT INTO users (username, password_hash, role) VALUES (?, ?, ?)", "admin", string(hashed), "admin")
		if err != nil {
			log.Fatalf("Failed seeding default administrator: %v", err)
		}
		fmt.Println("\n🔒 [SECURITY]: No users detected. Generated default credentials:")
		fmt.Println("👉 Username: admin")
		fmt.Println("👉 Password: Hermes")
	}
}

// ==============================================================================
// KEEP THESE FUNCTIONS DOWN HERE FOR NOW! 
// ==============================================================================

func syncHostInDB(host models.Host) (int64, error) {
	var id int64
	err := db.DB.QueryRow("SELECT id FROM hosts WHERE ip = ?", host.IP).Scan(&id)

	if err == sql.ErrNoRows {
		result, err := db.DB.Exec(
			"INSERT INTO hosts (name, ip, community, enabled) VALUES (?, ?, ?, ?)",
			host.Name, host.IP, host.Community, 1,
		)
		if err != nil {
			return 0, err
		}
		id, err = result.LastInsertId()
		if err != nil {
			return 0, err
		}
		fmt.Printf("Added new host to database: %s (ID: %d)\n", host.Name, id)
	} else if err != nil {
		return 0, err
	} else {
		fmt.Printf("Host %s already exists in database (ID: %d)\n", host.Name, id)
	}

	return id, nil
}