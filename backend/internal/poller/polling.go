package poller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"project-hermes/internal/db"

	"github.com/gosnmp/gosnmp"
)

type cacheKey struct {
	interfaceID int64
	direction   string
}

type cacheSample struct {
	value     uint64
	timestamp time.Time
}

type ruleState struct {
	breachStartTime   time.Time
	recoveryStartTime time.Time
	isAlerting        bool
}

type speedData struct {
	downMbps float64
	upMbps   float64
}

var pollerCache = make(map[cacheKey]cacheSample)
var alertTracker = make(map[int]*ruleState)

// Live cache that holds current Mbps of every port for alert evaluations
var liveSpeedCache sync.Map

type MonitoredTarget struct {
	HostID        int64
	HostName      string
	IP            string
	Community     string
	InterfaceID   int64
	IndexNum      string
	InterfaceName string
}

type HostGroup struct {
	HostID    int64
	HostName  string
	IP        string
	Community string
	Ports     []MonitoredTarget
}

func Start(interval time.Duration) {
	log.Printf("Initializing Background Polling Engine (Interval: %v)...", interval)
	ticker := time.NewTicker(interval)

	go func() {
		for range ticker.C {
			executePollCycle()
		}
	}()
}

func executePollCycle() {
	query := `
        SELECT h.id, h.name, h.ip, h.community, i.id, i.index_num, i.name 
        FROM interfaces i
        JOIN hosts h ON i.host_id = h.id
        WHERE h.enabled = 1 AND i.monitoring = 1;`

	rows, err := db.DB.Query(query)
	if err != nil {
		log.Printf("[Poller Error] Failed to fetch monitored ports: %v", err)
		return
	}
	defer rows.Close()

	var targets []MonitoredTarget
	for rows.Next() {
		var t MonitoredTarget
		if err := rows.Scan(&t.HostID, &t.HostName, &t.IP, &t.Community, &t.InterfaceID, &t.IndexNum, &t.InterfaceName); err == nil {
			targets = append(targets, t)
		}
	}

	if len(targets) == 0 {
		return
	}

	// 1. Group all monitored interfaces by Host ID
	hostMap := make(map[int64]*HostGroup)
	var hostOrder []int64

	for _, t := range targets {
		if _, exists := hostMap[t.HostID]; !exists {
			hostMap[t.HostID] = &HostGroup{
				HostID:    t.HostID,
				HostName:  t.HostName,
				IP:        t.IP,
				Community: t.Community,
				Ports:     []MonitoredTarget{},
			}
			hostOrder = append(hostOrder, t.HostID)
		}
		hostMap[t.HostID].Ports = append(hostMap[t.HostID].Ports, t)
	}

	// 2. Poll each HOST in parallel (1 connection per physical router/host)
	var wg sync.WaitGroup
	for i, hostID := range hostOrder {
		group := hostMap[hostID]
		wg.Add(1)

		// 50ms stagger between different hosts
		time.Sleep(time.Duration(i*50) * time.Millisecond)

		go pollHostGroup(*group, &wg)
	}

	// 3. Once all hosts respond, evaluate alert rules
	go func() {
		wg.Wait()
		evaluateAllAlerts(targets)
	}()
}

func pollHostGroup(group HostGroup, wg *sync.WaitGroup) {
	defer wg.Done()

	if len(group.Ports) == 0 {
		return
	}

	isV3 := strings.HasPrefix(group.Community, "v3:")
	authName := strings.TrimPrefix(group.Community, "v3:")

	agent := &gosnmp.GoSNMP{
		Target:  group.IP,
		Port:    161,
		Timeout: time.Duration(2) * time.Second,
		Retries: 1,
	}

	if isV3 {
		agent.Version = gosnmp.Version3
		agent.MsgFlags = gosnmp.NoAuthNoPriv
		agent.SecurityModel = gosnmp.UserSecurityModel
		agent.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName: authName,
		}
	} else {
		agent.Version = gosnmp.Version2c
		agent.Community = authName
	}

	if err := agent.Connect(); err != nil {
		return
	}
	defer agent.Conn.Close()

	// Build a single batch array of all OIDs for all interfaces on this host
	var oids []string
	for _, p := range group.Ports {
		oidIn := ".1.3.6.1.2.1.31.1.1.1.6." + p.IndexNum
		oidOut := ".1.3.6.1.2.1.31.1.1.1.10." + p.IndexNum
		oids = append(oids, oidIn, oidOut)
	}

	// Query ALL interface OIDs in ONE single UDP request
	result, err := agent.Get(oids)
	if err != nil {
		return
	}

	// Map returned values by OID string for fast lookup
	valMap := make(map[string]uint64)
	for _, pdu := range result.Variables {
		val := gosnmp.ToBigInt(pdu.Value).Uint64()
		valMap[pdu.Name] = val
		valMap[strings.TrimPrefix(pdu.Name, ".")] = val
	}

	now := time.Now()

	// Process each interface using the fetched batch data
	for _, p := range group.Ports {
		oidIn := ".1.3.6.1.2.1.31.1.1.1.6." + p.IndexNum
		oidOut := ".1.3.6.1.2.1.31.1.1.1.10." + p.IndexNum

		currentIn, okIn := getPDUVal(valMap, oidIn)
		currentOut, okOut := getPDUVal(valMap, oidOut)

		if !okIn || !okOut {
			continue
		}

		downloadMbps := calculateMbps(p.InterfaceID, "in", currentIn, now)
		uploadMbps := calculateMbps(p.InterfaceID, "out", currentOut, now)

		if downloadMbps >= 0 && uploadMbps >= 0 {
			// Save to Database
			_, _ = db.DB.Exec(`
                INSERT INTO metrics (interface_id, download_mbps, upload_mbps, timestamp)
                VALUES (?, ?, ?, ?)`,
				p.InterfaceID, downloadMbps, uploadMbps, now,
			)

			// Save to live cache for alert evaluations
			cacheKey := fmt.Sprintf("%d_%s", p.HostID, p.InterfaceName)
			liveSpeedCache.Store(cacheKey, speedData{downMbps: downloadMbps, upMbps: uploadMbps})
		}
	}
}

func getPDUVal(valMap map[string]uint64, oid string) (uint64, bool) {
	if v, ok := valMap[oid]; ok {
		return v, true
	}
	trimmed := strings.TrimPrefix(oid, ".")
	if v, ok := valMap[trimmed]; ok {
		return v, true
	}
	return 0, false
}

func calculateMbps(interfaceID int64, direction string, currentValue uint64, now time.Time) float64 {
	key := cacheKey{interfaceID: interfaceID, direction: direction}
	prevSample, exists := pollerCache[key]

	pollerCache[key] = cacheSample{value: currentValue, timestamp: now}

	if !exists {
		return -1
	}

	timeElapsed := now.Sub(prevSample.timestamp).Seconds()
	if timeElapsed <= 0 || currentValue < prevSample.value {
		return 0
	}

	deltaBits := (currentValue - prevSample.value) * 8
	return (float64(deltaBits) / timeElapsed) / 1000000.0
}

// --- SUMMATION EXPERT ENGINE ---
func evaluateAllAlerts(targets []MonitoredTarget) {
	var botToken, chatID, enabledStr string
	db.DB.QueryRow("SELECT value FROM settings WHERE key = 'telegram_bot_token'").Scan(&botToken)
	db.DB.QueryRow("SELECT value FROM settings WHERE key = 'telegram_chat_id'").Scan(&chatID)
	db.DB.QueryRow("SELECT value FROM settings WHERE key = 'telegram_enabled'").Scan(&enabledStr)

	if enabledStr != "true" || botToken == "" || chatID == "" {
		return
	}

	// Fetch all enabled rules regardless of host
	rows, err := db.DB.Query(`
        SELECT id, host_id, port, direction, alert_threshold_mbps, alert_duration_mins, alert_template, 
               recovery_threshold_mbps, recovery_duration_mins, recovery_template 
        FROM alert_rules 
        WHERE enabled = 1`)

	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var ruleID, hostID int
		var portsJSON, direction, alertTpl, recovTpl string
		var alertThresh, recovThresh float64
		var alertDur, recovDur int

		if err := rows.Scan(&ruleID, &hostID, &portsJSON, &direction, &alertThresh, &alertDur, &alertTpl, &recovThresh, &recovDur, &recovTpl); err != nil {
			continue
		}

		// Convert JSON string back to []string
		var ports []string
		json.Unmarshal([]byte(portsJSON), &ports)

		var totalSpeed float64
		var hostName string

		// Aggregate speeds across configured ports
		for _, p := range ports {
			cacheKey := fmt.Sprintf("%d_%s", hostID, p)
			if val, ok := liveSpeedCache.Load(cacheKey); ok {
				sd := val.(speedData)
				if strings.ToLower(direction) == "upload" {
					totalSpeed += sd.upMbps
				} else {
					totalSpeed += sd.downMbps
				}
			}

			for _, t := range targets {
				if t.HostID == int64(hostID) {
					hostName = t.HostName
					break
				}
			}
		}

		if alertTracker[ruleID] == nil {
			alertTracker[ruleID] = &ruleState{}
		}
		state := alertTracker[ruleID]

		groupedPortNames := strings.Join(ports, " + ")

		// State machine logic for alerts
		if !state.isAlerting {
			if totalSpeed > alertThresh {
				if state.breachStartTime.IsZero() {
					state.breachStartTime = time.Now()
				} else if time.Since(state.breachStartTime).Minutes() >= float64(alertDur) {
					state.isAlerting = true
					state.breachStartTime = time.Time{}
					msg := formatTemplate(alertTpl, hostName, groupedPortNames, totalSpeed)
					sendTelegramNotification(botToken, chatID, msg)
				}
			} else {
				state.breachStartTime = time.Time{}
			}
		} else {
			if totalSpeed < recovThresh {
				if state.recoveryStartTime.IsZero() {
					state.recoveryStartTime = time.Now()
				} else if time.Since(state.recoveryStartTime).Minutes() >= float64(recovDur) {
					state.isAlerting = false
					state.recoveryStartTime = time.Time{}
					msg := formatTemplate(recovTpl, hostName, groupedPortNames, totalSpeed)
					sendTelegramNotification(botToken, chatID, msg)
				}
			} else {
				state.recoveryStartTime = time.Time{}
			}
		}
	}
}

func formatTemplate(tpl, device, port string, speed float64) string {
	msg := strings.ReplaceAll(tpl, "[DEVICE]", device)
	msg = strings.ReplaceAll(msg, "[PORT]", port)
	msg = strings.ReplaceAll(msg, "[SPEED]", fmt.Sprintf("%.2f", speed))
	return msg
}

func sendTelegramNotification(token, chatID, text string) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	payload := map[string]string{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "Markdown",
	}
	body, _ := json.Marshal(payload)
	http.Post(url, "application/json", bytes.NewBuffer(body))
}
