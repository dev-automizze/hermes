# HERMES - Network Bandwidth Monitor

HERMES is a lightweight, self-hosted network bandwidth monitoring tool built with Go and Vue.js. It uses SNMP to poll your routers and switches, stores historical data in SQLite, and provides a sleek, customizable dashboard for visualization and alerting.

## ✨ Features

*   **Multi-Host Support:** Monitor multiple routers/switches simultaneously.
*   **SNMP v1, v2c, and v3 Support.**
*   **Real-time Dashboard:** Beautiful, interactive charts showing download/upload speeds.
*   **Customize Trigger:** Free Customize alert to webhook for high bandwidth usage. 
*   **Per-User Persistent Layouts:** Each user account can customize their dashboard grid, time ranges, and monitored interfaces. These settings are saved securely to the database and load on any device.
*   **Telegram Alerts:** Configure custom thresholds and durations to receive alert and recovery notifications directly via Telegram.
*   **Docker Ready:** Easy deployment using Docker and Docker Compose.
*   **Data Retention:** Configurable database purging to keep your SQLite file small.



🚀 Quick Start (Using GitHub Packages / GHCR)
If you prefer to pull the pre-built image directly from GitHub Packages, you can do so without downloading the source code.

Open your terminal and run:
```bash
mkdir hermes
cd hermes

# Pull the latest image from GHCR
docker pull ghcr.io/dev-automizze/hermes:latest

# Run the container with the correct port and volume mapping
docker run -d \
  --name hermes-network-monitor \
  -p 8014:8014 \
  -v ./data:/app/data \
  --restart unless-stopped \
  ghcr.io/dev-automizze/hermes:latest
```

Once running, access the dashboard at http://<your-server-ip>:8014.



Alternative way with docker compose

🚀Docker compose
```yaml
services:
  hermes-monitor:
    # Use the pre-built image from GHCR
    image: ghcr.io/dev-automizze/hermes:latest
    container_name: hermes-network-monitor
    ports:
      - "8014:8014"
    volumes:
      # This binds a local directory named 'data' on your real machine 
      # to the container's safe directory, making your DB immortal!
      - ./data:/app/data
    restart: unless-stopped
```

```bash
docker compose up -d
```

#Login

User: admin

Pass: Hermes


⚙️ Configuration Guide
Adding Hosts and Interfaces
Navigate to the Settings tab.

Add a new Host with its IP address and SNMP Community String.

Once added, the system will automatically discover its interfaces.

Enable "Monitoring" on the specific interfaces you want to track.

SNMP v3 Support Trick
HERMES uses a custom prefix to detect SNMP v3 authentication.

For SNMP v2c or v1, simply enter your community string (e.g., public).

For SNMP v3, enter v3: followed by your SNMP v3 Username in the Community String box.

Example: v3:myv3username

(Note: This implementation currently uses NoAuthNoPriv security level).

Telegram Alerts
Go to Settings -> Telegram.

Enter your Bot Token and Chat ID.

Go to Settings -> Alerts to create rules based on Host, Port, Direction (Upload/Download), and Thresholds (Mbps).


🛠️ Local Development
If you want to build the project from source:

Backend (Go):
```bash
cd backend
go mod download
go run cmd/hermes/main.go
```
Frontend (Vue 3 + Vite):
```bash
cd frontend
npm install
npm run dev
```
📦 Built With
*Go - Backend logic, SNMP polling, and API.
*Vue.js 3 - Frontend framework.
*SQLite - Embedded database.

No fking web coding here just Gemini and & Deepseek.






