package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"project-hermes/internal/db"

	"github.com/gin-gonic/gin"
)

type TelegramSettings struct {
	BotToken string `json:"bot_token"`
	ChatID   string `json:"chat_id"`
	Enabled  bool   `json:"enabled"`
}

// 1. Fetch Current Settings
func GetTelegramSettings(c *gin.Context) {
	var settings TelegramSettings

	// Helper function to get a value from the DB safely
	getValue := func(key string) string {
		var val string
		db.DB.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&val)
		return val
	}

	settings.BotToken = getValue("telegram_bot_token")
	settings.ChatID = getValue("telegram_chat_id")
	settings.Enabled = getValue("telegram_enabled") == "true"

	c.JSON(http.StatusOK, settings)
}

// 2. Save Settings from the UI
func SaveTelegramSettings(c *gin.Context) {
	var req TelegramSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	enabledStr := "false"
	if req.Enabled {
		enabledStr = "true"
	}

	// Insert or Replace settings in the database
	stmt := `INSERT INTO settings (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value=excluded.value;`
	db.DB.Exec(stmt, "telegram_bot_token", req.BotToken)
	db.DB.Exec(stmt, "telegram_chat_id", req.ChatID)
	db.DB.Exec(stmt, "telegram_enabled", enabledStr)

	c.JSON(http.StatusOK, gin.H{"message": "Settings saved successfully"})
}

// 3. The "Test Connection" Trigger
func TestTelegramConnection(c *gin.Context) {
	var req TelegramSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	if req.BotToken == "" || req.ChatID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bot Token and Chat ID are required"})
		return
	}

	err := sendTelegramMessage(req.BotToken, req.ChatID, "✅ *Hermes Bandwidth Monitor*\nTest connection successful! Your notification engine is online.")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Test message sent!"})
}

// The actual HTTP request to Telegram's API
func sendTelegramMessage(token, chatID, text string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)

	payload := map[string]string{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "Markdown",
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API rejected the request with status: %d", resp.StatusCode)
	}
	return nil
}
