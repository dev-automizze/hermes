# --- STAGE 1: Build the Vue Frontend ---
FROM node:20-alpine AS frontend-builder
WORKDIR /frontend

# Grab the package.json from inside your frontend folder
COPY frontend/package*.json ./
RUN npm install

# Copy the rest of the frontend source code
COPY frontend/ ./
RUN npm run build

# --- STAGE 2: Build the Go Backend ---
FROM golang:1.25-alpine AS backend-builder
RUN apk add --no-cache gcc musl-dev sqlite-dev
WORKDIR /backend

# Copy the entire backend directory (contains go.mod, cmd/, internal/, etc.)
COPY backend/ ./

# Download dependencies from the backend root folder
RUN go mod download

# Build the binary by pointing directly to main.go inside cmd/hermes
RUN go build -o /app/hermes-app ./cmd/hermes/main.go

# --- STAGE 3: Final Lightweight Production Runner ---
FROM alpine:latest
RUN apk add --no-cache libc6-compat sqlite
WORKDIR /app

# Copy the compiled binary from Stage 2
COPY --from=backend-builder /app/hermes-app .
# Copy the compiled static frontend files from Stage 1
COPY --from=frontend-builder /frontend/dist ./dist

# Create the data folder for our SQLite database
RUN mkdir -p data
EXPOSE 8014

CMD ["./hermes-app"]