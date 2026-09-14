package api

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"project-hermes/internal/db"
	"project-hermes/internal/models"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gosnmp/gosnmp"
	"golang.org/x/crypto/bcrypt"
)

// SetupRouter initializes the Gin web server and defines our endpoints
func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
	}))

	// --- PUBLIC ENDPOINTS ---
	apiGroup := r.Group("/api")
	{
		apiGroup.POST("/auth/login", LoginHandler)
		apiGroup.GET("/auth/check", CheckAuthHandler)
	}

	// --- SECURE PROTECTED ENDPOINTS ---
	protected := r.Group("/api")
	protected.Use(AuthMiddleware())
	{
		protected.GET("/interfaces", getInterfaces)
		protected.PUT("/interfaces/:id", updateInterface)
		protected.GET("/metrics/:id", getMetrics)

		// Host Management Routes
		protected.GET("/hosts", getHosts)
		protected.POST("/hosts", saveHost)
		protected.DELETE("/hosts/:id", DeleteHost)

		// Alert Engine Routes (Assuming defined in alerts.go)
		protected.GET("/alerts", GetAlertRules)
		protected.POST("/alerts", SaveAlertRules)

		// Telegram Notification Routes (Assuming defined in telegram.go)
		protected.GET("/settings/telegram", GetTelegramSettings)
		protected.POST("/settings/telegram", SaveTelegramSettings)
		protected.POST("/settings/telegram/test", TestTelegramConnection)

		// Database & History Routes
		protected.GET("/settings/database", GetDatabaseStats)
		protected.POST("/settings/database", SaveDatabaseSettings)
		protected.POST("/settings/database/purge", PurgeMetrics)

		// NEW: User Management Routes
		protected.GET("/users", GetUsers)
		protected.POST("/users", CreateUser)
		protected.PUT("/users/:id", UpdateUser)
		protected.DELETE("/users/:id", DeleteUser)
	}

	// -------------------------------------------------------------
	// SERVE EMBEDDED FRONTEND STATIC FILES
	// -------------------------------------------------------------
	// Serve the compiled static assets directory (js, css, images)
	r.Static("/assets", "./dist/assets")

	// Catch-all route: If the user refreshes on any frontend page, send them index.html
	r.NoRoute(func(c *gin.Context) {
    c.File("./dist/index.html")
	})

	return r
}

// ==============================================================================
// USER SECURITY & AUTHENTICATION ENGINES
// ==============================================================================

// AuthMiddleware intercepts API calls to ensure a valid session exists and extracts roles
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Access Denied: Missing auth credentials"})
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")

		var userID int
		var role, username string
		
		// Extract the user's role and username directly from their active session
		err := db.DB.QueryRow(`
			SELECT s.user_id, u.role, u.username FROM sessions s
			JOIN users u ON s.user_id = u.id
			WHERE s.token = ? AND s.expires_at > datetime('now')`, token).Scan(&userID, &role, &username)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Access Denied: Session expired or invalid"})
			return
		}

		// Inject security context into the request for downstream RBAC checks
		c.Set("userID", userID)
		c.Set("role", role)
		c.Set("username", username)
		c.Next()
	}
}

// LoginHandler processes user logins, updates hash checks, and generates session keys
func LoginHandler(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and password required"})
		return
	}

	var userID int
	var passwordHash string
	var role string

	err := db.DB.QueryRow("SELECT id, password_hash, role FROM users WHERE username = ?", req.Username).Scan(&userID, &passwordHash, &role)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// Evaluate cryptographic hash matches
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// Generate secure session identifier string
	randBytes := make([]byte, 24)
	rand.Read(randBytes)
	token := hex.EncodeToString(randBytes)

	// Set session life (e.g., 7 Days session permanence)
	expiration := time.Now().Add(7 * 24 * time.Hour)

	_, err = db.DB.Exec("INSERT INTO sessions (token, user_id, expires_at) VALUES (?, ?, ?)", token, userID, expiration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed creating an active session mapping"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":    token,
		"username": req.Username,
		"role":     role,
	})
}

// CheckAuthHandler allows web client dashboards to verify token safety on page boots
func CheckAuthHandler(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{"authenticated": false})
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	var username, role string
	err := db.DB.QueryRow(`
		SELECT u.username, u.role FROM users u
		JOIN sessions s ON u.id = s.user_id
		WHERE s.token = ? AND s.expires_at > datetime('now')`, token).Scan(&username, &role)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"authenticated": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{"authenticated": true, "username": username, "role": role})
}

// ==============================================================================
// USER MANAGEMENT ENDPOINTS
// ==============================================================================

func GetUsers(c *gin.Context) {
	rows, err := db.DB.Query("SELECT id, username, role FROM users")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	defer rows.Close()

	var users []map[string]interface{}
	for rows.Next() {
		var id int
		var username, role string
		if err := rows.Scan(&id, &username, &role); err == nil {
			users = append(users, map[string]interface{}{
				"id":       id,
				"username": username,
				"role":     role,
			})
		}
	}
	
	if users == nil {
		users = []map[string]interface{}{}
	}
	
	c.JSON(http.StatusOK, users)
}

func CreateUser(c *gin.Context) {
	if c.GetString("role") == "guest" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Guests cannot create users"})
		return
	}

	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Role     string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to secure password"})
		return
	}

	_, err = db.DB.Exec("INSERT INTO users (username, password_hash, role) VALUES (?, ?, ?)", req.Username, string(hashed), req.Role)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User created successfully"})
}

func UpdateUser(c *gin.Context) {
	if c.GetString("role") == "guest" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Guests cannot modify users"})
		return
	}

	targetID := c.Param("id")
	callerUsername := c.GetString("username")

	// SUDO PROTECTION: If trying to modify ID 1 (admin), the caller must be admin
	if targetID == "1" && callerUsername != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the Sudo Administrator can modify this account"})
		return
	}

	var req struct {
		Password string `json:"password"`
		Role     string `json:"role"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	if req.Password != "" {
		hashed, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		db.DB.Exec("UPDATE users SET password_hash = ? WHERE id = ?", string(hashed), targetID)
	}

	// Sudo role cannot be downgraded
	if req.Role != "" && targetID != "1" {
		db.DB.Exec("UPDATE users SET role = ? WHERE id = ?", req.Role, targetID)
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

func DeleteUser(c *gin.Context) {
	if c.GetString("role") == "guest" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Guests cannot delete users"})
		return
	}

	targetID := c.Param("id")

	// SUDO PROTECTION: ID 1 can NEVER be deleted
	if targetID == "1" {
		c.JSON(http.StatusForbidden, gin.H{"error": "The Sudo Administrator account cannot be deleted"})
		return
	}

	// Prevent user from deleting themselves
	if fmt.Sprintf("%d", c.GetInt("userID")) == targetID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You cannot delete your own active session"})
		return
	}

	db.DB.Exec("DELETE FROM users WHERE id = ?", targetID)
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// ==============================================================================
// CORE HERMES HANDLERS
// ==============================================================================

func getHosts(c *gin.Context) {
	rows, err := db.DB.Query("SELECT id, name, ip, community, enabled FROM hosts")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch hosts from database"})
		return
	}
	defer rows.Close()

	type HostResponse struct {
		ID        int    `json:"id"`
		Name      string `json:"name"`
		IP        string `json:"ip"`
		Community string `json:"community"`
		Enabled   bool   `json:"enabled"`
	}

	var hosts []HostResponse

	for rows.Next() {
		var h HostResponse
		var enabledInt int
		if err := rows.Scan(&h.ID, &h.Name, &h.IP, &h.Community, &enabledInt); err != nil {
			continue
		}
		h.Enabled = enabledInt == 1
		hosts = append(hosts, h)
	}

	if hosts == nil {
		hosts = []HostResponse{}
	}

	c.JSON(http.StatusOK, hosts)
}

func saveHost(c *gin.Context) {
	if c.GetString("role") == "guest" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Read-only access"})
		return
	}

	var req struct {
		ID        int    `json:"id"` // <--- WE NOW CATCH THE ID
		Name      string `json:"name" binding:"required"`
		IP        string `json:"ip" binding:"required"`
		Community string `json:"community" binding:"required"`
		Enabled   bool   `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data"})
		return
	}

	enabledInt := 0
	if req.Enabled {
		enabledInt = 1
	}

	var hostID int64

	// If ID > 0, it's an UPDATE. Otherwise, it's a NEW host.
	if req.ID > 0 {
		_, err := db.DB.Exec("UPDATE hosts SET name=?, ip=?, community=?, enabled=? WHERE id=?", 
			req.Name, req.IP, req.Community, enabledInt, req.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update host configuration"})
			return
		}
		hostID = int64(req.ID)
	} else {
		result, err := db.DB.Exec("INSERT INTO hosts (name, ip, community, enabled) VALUES (?, ?, ?, ?)", 
			req.Name, req.IP, req.Community, enabledInt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save new host configuration"})
			return
		}
		hostID, err = result.LastInsertId()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Host saved, but failed to retrieve internal ID"})
			return
		}
	}

	go func() {
		log.Printf("Spawning background SNMP interface discovery for %s (%s)...", req.Name, req.IP)
		if err := discoverAndSaveInterfaces(hostID, req.IP, req.Community); err != nil {
			log.Printf("Background interface discovery failed for host %s: %v", req.Name, err)
		} else {
			log.Printf("Background interface discovery successfully complete for %s!", req.Name)
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "Host configuration saved and discovery started successfully"})
}

func DeleteHost(c *gin.Context) {
	if c.GetString("role") == "guest" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Read-only access"})
		return
	}

	id := c.Param("id")
    
	// 1. Delete the host from the hosts table
	_, err := db.DB.Exec("DELETE FROM hosts WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete host"})
		return
	}

	// 2. Cleanup: Delete all associated interfaces so they don't get left behind as ghosts
	_, _ = db.DB.Exec("DELETE FROM interfaces WHERE host_id = ?", id)

	c.JSON(http.StatusOK, gin.H{"message": "Host deleted successfully"})
}

func getInterfaces(c *gin.Context) {
	hostID := c.Query("host_id")

	var rows *sql.Rows
	var err error

	if hostID != "" {
		rows, err = db.DB.Query("SELECT id, host_id, index_num, name, monitoring FROM interfaces WHERE host_id = ?", hostID)
	} else {
		rows, err = db.DB.Query("SELECT id, host_id, index_num, name, monitoring FROM interfaces")
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch interfaces"})
		return
	}
	defer rows.Close()

	var interfaces []models.Interface

	for rows.Next() {
		var i models.Interface
		var monitoringInt int
		if err := rows.Scan(&i.ID, &i.HostID, &i.IndexNum, &i.Name, &monitoringInt); err != nil {
			continue
		}

		i.Monitoring = monitoringInt == 1
		interfaces = append(interfaces, i)
	}

	if interfaces == nil {
		interfaces = []models.Interface{}
	}

	c.JSON(http.StatusOK, interfaces)
}

func updateInterface(c *gin.Context) {
	if c.GetString("role") == "guest" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Read-only access"})
		return
	}

	interfaceID := c.Param("id")

	var req struct {
		Monitoring bool `json:"monitoring"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	monitoringInt := 0
	if req.Monitoring {
		monitoringInt = 1
	}

	_, err := db.DB.Exec(`
		UPDATE interfaces 
		SET monitoring = ? 
		WHERE id = ?`, monitoringInt, interfaceID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update monitoring status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Interface updated successfully"})
}

func getMetrics(c *gin.Context) {
	interfaceID := c.Param("id")
	durationStr := c.DefaultQuery("duration", "5m")

	var startTime, endTime time.Time
	isCustom := durationStr == "custom"

	if isCustom {
		startStr := c.Query("start")
		endStr := c.Query("end")

		var errStart, errEnd error
		startTime, errStart = time.Parse(time.RFC3339, startStr)
		if errStart != nil {
			startTime, errStart = time.Parse("2006-01-02T15:04", startStr)
		}

		endTime, errEnd = time.Parse(time.RFC3339, endStr)
		if errEnd != nil {
			endTime, errEnd = time.Parse("2006-01-02T15:04", endStr)
		}

		if errStart != nil || errEnd != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid custom date-time parameters"})
			return
		}
	} else {
		duration, err := time.ParseDuration(durationStr)
		if err != nil {
			duration = 5 * time.Minute
		}
		endTime = time.Now()
		startTime = endTime.Add(-duration)
	}

	timeDiff := endTime.Sub(startTime)

	var query string
	var rows interface{}
	var queryErr error

	if isCustom && timeDiff > 24*time.Hour {
		query = `
			SELECT AVG(download_mbps) as download_mbps, AVG(upload_mbps) as upload_mbps, 
			       strftime('%Y-%m-%dT%H:00:00Z', timestamp) as grouped_time
			FROM metrics 
			WHERE interface_id = ? AND timestamp >= ? AND timestamp <= ?
			GROUP BY grouped_time
			ORDER BY grouped_time ASC`

		rows, queryErr = db.DB.Query(query, interfaceID, startTime, endTime)
	} else {
		query = `
			SELECT download_mbps, upload_mbps, timestamp 
			FROM metrics 
			WHERE interface_id = ? AND timestamp >= ? AND timestamp <= ?
			ORDER BY timestamp ASC`

		rows, queryErr = db.DB.Query(query, interfaceID, startTime, endTime)
	}

	if queryErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query historical metrics"})
		return
	}
	dbRows := rows.(*sql.Rows)
	defer dbRows.Close()

	type MetricResponse struct {
		Download  float64 `json:"download_mbps"`
		Upload    float64 `json:"upload_mbps"`
		Timestamp string  `json:"timestamp"`
	}

	var metrics []MetricResponse

	for dbRows.Next() {
		var m MetricResponse
		var rawTimestamp interface{}
		if err := dbRows.Scan(&m.Download, &m.Upload, &rawTimestamp); err != nil {
			continue
		}

		switch v := rawTimestamp.(type) {
		case time.Time:
			m.Timestamp = v.Format(time.RFC3339)
		case string:
			m.Timestamp = v
		case []byte:
			m.Timestamp = string(v)
		default:
			m.Timestamp = time.Now().Format(time.RFC3339)
		}

		metrics = append(metrics, m)
	}

	if metrics == nil {
		metrics = []MetricResponse{}
	}

	c.JSON(http.StatusOK, metrics)
}

func discoverAndSaveInterfaces(hostID int64, ip, community string) error {
	agent := &gosnmp.GoSNMP{
		Target:    ip,
		Port:      161,
		Community: community,
		Version:   gosnmp.Version2c,
		Timeout:   time.Duration(5) * time.Second,
		Retries:   3,
	}

	err := agent.Connect()
	if err != nil {
		return err
	}
	defer agent.Conn.Close()

	type DiscoveredPort struct {
		Index string
		Name  string
	}
	var discoveredPorts []DiscoveredPort

	err = agent.Walk(".1.3.6.1.2.1.31.1.1.1.1", func(pdu gosnmp.SnmpPDU) error {
		oidParts := strings.Split(pdu.Name, ".")
		indexStr := oidParts[len(oidParts)-1]

		var name string
		switch pdu.Type {
		case gosnmp.OctetString:
			name = string(pdu.Value.([]byte))
		default:
			name = fmt.Sprintf("Interface-%s", indexStr)
		}

		name = strings.TrimSpace(name)
		if name == "" {
			name = fmt.Sprintf("Port-%s", indexStr)
		}

		discoveredPorts = append(discoveredPorts, DiscoveredPort{Index: indexStr, Name: name})
		return nil
	})

	if err != nil {
		return err
	}

	tx, err := db.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	stmt, err := tx.Prepare(`
		INSERT INTO interfaces (host_id, index_num, name, monitoring) 
		VALUES (?, ?, ?, 0)
		ON CONFLICT(host_id, index_num) DO UPDATE SET name=excluded.name`)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	for _, port := range discoveredPorts {
		if _, err := stmt.Exec(hostID, port.Index, port.Name); err != nil {
			log.Printf("Failed to save interface index %s: %v", port.Index, err)
		}
	}

	return tx.Commit()
}

func GetDatabaseStats(c *gin.Context) {
	var pageCount, pageSize int64
	err := db.DB.QueryRow("PRAGMA page_count").Scan(&pageCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read database page count"})
		return
	}
	err = db.DB.QueryRow("PRAGMA page_size").Scan(&pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read database page size"})
		return
	}

	sizeInBytes := pageCount * pageSize
	sizeInMB := float64(sizeInBytes) / (1024 * 1024)

	var metricsCount int64
	err = db.DB.QueryRow("SELECT COUNT(*) FROM metrics").Scan(&metricsCount)
	if err != nil {
		metricsCount = 0
	}

	var hostsCount int
	err = db.DB.QueryRow("SELECT COUNT(*) FROM hosts").Scan(&hostsCount)
	if err != nil {
		hostsCount = 0
	}

	var retentionDays string
	err = db.DB.QueryRow("SELECT value FROM settings WHERE key = 'data_retention_days'").Scan(&retentionDays)
	if err != nil || retentionDays == "" {
		retentionDays = "7" 
	}

	c.JSON(http.StatusOK, gin.H{
		"database_size_mb": fmt.Sprintf("%.2f MB", sizeInMB),
		"total_metrics":    metricsCount,
		"total_hosts":      hostsCount,
		"retention_days":   retentionDays,
	})
}

func SaveDatabaseSettings(c *gin.Context) {
	if c.GetString("role") == "guest" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Read-only access"})
		return
	}

	var req struct {
		RetentionDays int `json:"retention_days"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	stmt := `INSERT INTO settings (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value=excluded.value;`
	_, err := db.DB.Exec(stmt, "data_retention_days", fmt.Sprintf("%d", req.RetentionDays))
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save retention policy"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Data retention policy saved successfully!"})
}

func PurgeMetrics(c *gin.Context) {
	if c.GetString("role") == "guest" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Read-only access"})
		return
	}

	var req struct {
		RetentionDays int `json:"retention_days"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	var err error
	if req.RetentionDays <= 0 {
		_, err = db.DB.Exec("DELETE FROM metrics")
	} else {
		_, err = db.DB.Exec("DELETE FROM metrics WHERE timestamp < datetime('now', ?)", fmt.Sprintf("-%d days", req.RetentionDays))
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete historical data"})
		return
	}

	_, err = db.DB.Exec("VACUUM")
	if err != nil {
		log.Printf("Warning: VACUUM operation failed: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Database history cleaned and optimized successfully"})
}