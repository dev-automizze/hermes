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

## 🚀 Quick Start (Docker Compose)

The easiest way to run HERMES is using Docker Compose.

1. Create a folder for your HERMES installation and create a `docker-compose.yml` file with the following content:

```yaml
services:
  hermes-monitor:
    image: <your-dockerhub-username>/<your-repo-name>:latest
    container_name: hermes-network-monitor
    ports:
      - "8014:8014"
    volumes:
      # This binds a local directory named 'data' on your real machine 
      # to the container's safe directory, making your DB immortal!
      - ./data:/app/data
    restart: unless-stopped
