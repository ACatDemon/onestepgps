# Setup

## Installation
1. Install Go from https://go.dev/doc/install
2. Install Node.js from https://nodejs.org

## Environment Variables
Create two `.env` files:

1. Root directory:
```
ONESTEPGPS_API_KEY=your_api_key
```

2. device-map directory:
```
VUE_APP_GOOGLE_MAPS_API_KEY=your_google_maps_api_key
```

## Run Backend
```bash
go get github.com/joho/godotenv
go run main/server.go
```
Server runs on http://localhost:3000

## Run Frontend
```bash
cd device-map
npm install
npm run serve
```
Frontend runs on http://localhost:8080
