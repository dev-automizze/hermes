package polling

import (
	"fmt"
	"log"
	"project-hermes/internal/db"
	"time"

	"github.com/gosnmp/gosnmp"
)

// cacheData remembers the last seen byte counts to calculate speeds
type cacheData struct {
	LastIn       uint64
	LastOut      uint64
	LastPolledAt time.Time
}

var memoryCache = make(map[int64]cacheData) // Key is the Interface ID

// StartEngine boots up the background ticker loop
func StartEngine(interval time.Duration) {
	ticker := time.NewTicker(interval)

	log.Printf("Background Polling Engine started (Interval: %v)", interval)

	// This loop runs forever in the background
	for range ticker.C {
		pollActiveInterfaces()
	}
}

func pollActiveInterfaces() {
	// 1. Get all interfaces that the user wants to monitor, along with their Host details
	query := `
		SELECT i.id, i.index_num, i.name, h.ip, h.community 
		FROM interfaces i
		JOIN hosts h ON i.host_id = h.id
		WHERE i.monitoring = 1 AND h.enabled = 1`

	rows, err := db.DB.Query(query)
	if err != nil {
		log.Printf("DB Error fetching monitored ports: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var indexNum int
		var name, ip, community string

		if err := rows.Scan(&id, &indexNum, &name, &ip, &community); err != nil {
			continue
		}

		// 2. Fetch current bandwidth counters for this specific interface
		go pollSingleInterface(id, indexNum, name, ip, community)
	}
}

func pollSingleInterface(id int64, indexNum int, name, ip, community string) {
	agent := &gosnmp.GoSNMP{
		Target:    ip,
		Port:      161,
		Community: community,
		Version:   gosnmp.Version2c,
		Timeout:   time.Duration(3) * time.Second,
	}

	if err := agent.Connect(); err != nil {
		return
	}
	defer agent.Conn.Close()

	// High-capacity 64-bit counter OIDs
	oidIn := fmt.Sprintf(".1.3.6.1.2.1.31.1.1.1.6.%d", indexNum)
	oidOut := fmt.Sprintf(".1.3.6.1.2.1.31.1.1.1.10.%d", indexNum)

	result, err := agent.Get([]string{oidIn, oidOut})
	if err != nil {
		return
	}

	currentIn := gosnmp.ToBigInt(result.Variables[0].Value).Uint64()
	currentOut := gosnmp.ToBigInt(result.Variables[1].Value).Uint64()
	now := time.Now()

	// 3. Calculate speed if we have a previous data point in our cache
	if cache, exists := memoryCache[id]; exists {
		timeElapsed := now.Sub(cache.LastPolledAt).Seconds()
		if timeElapsed > 0 {
			// Bits per second calculation
			inBps := float64((currentIn-cache.LastIn)*8) / timeElapsed
			outBps := float64((currentOut-cache.LastOut)*8) / timeElapsed

			// Convert to Megabits per second (Mbps)
			inMbps := inBps / 1000000
			outMbps := outBps / 1000000

			// 4. Save metrics into the SQLite database
			_, err = db.DB.Exec(`
				INSERT INTO metrics (interface_id, download_mbps, upload_mbps, timestamp) 
				VALUES (?, ?, ?, ?)`,
				id, inMbps, outMbps, now,
			)
			if err != nil {
				log.Printf("Failed to save metric for %s: %v", name, err)
			} else {
				log.Printf("[%s] Speeds Saved -> Down: %.2f Mbps | Up: %.2f Mbps", name, inMbps, outMbps)
			}
		}
	}

	// Update cache with current values for the next round
	memoryCache[id] = cacheData{
		LastIn:       currentIn,
		LastOut:      currentOut,
		LastPolledAt: now,
	}
}
