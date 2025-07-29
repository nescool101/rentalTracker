# Docker Compose Integration

This document explains how the frontend and backend services are integrated using Docker Compose.

## Components

### Frontend (React with Vite)

- Built from `./frontend/Dockerfile`
- Uses a multi-stage build with Node.js and Nginx
- Exposes port 80 (Nginx web server)
- Nginx configuration:
  - Serves static Vite-built React files
  - Proxies API requests to the backend service
  - Handles client-side routing with fallback to index.html

### Backend (Go)

- Built from `./backend/Dockerfile`
- Exposes port 8081
- Environment variables:
  - `SERVER_PORT`: Sets the port for the HTTP server
  - `GIN_MODE`: Set to 'release' for production
  - `PAYERS_FILE_PATH`: Path to the payers.json file
  - `SUPABASE_URL`: Supabase project URL
  - `SUPABASE_KEY`: Supabase public API key

## API Integration

The frontend connects to the backend via:

1. `frontend/src/api/apiService.ts` - Contains API services to interact with the backend
2. `frontend/src/api/supabaseClient.ts` - Configures the Supabase client for direct Supabase access
3. `backend/controller/http_controller.go` - Implements CORS handling and RESTful endpoints
4. Nginx reverse proxy for seamless API integration

## Available Endpoints

### Legacy Endpoints
- `GET /payers` - Fetches all payers data
- `GET /validate_email` - Triggers email validation/notification logic

### RESTful API Endpoints
- `GET /api/persons` - Get all persons
- `GET /api/persons/:id` - Get person by ID
- `GET /api/properties` - Get all properties
- `GET /api/properties/:id` - Get property by ID
- `GET /api/rentals` - Get all rentals
- `GET /api/rentals/:id` - Get rental by ID

## Data Persistence

- The `payers.json` file is mounted into the backend container to persist data
- Supabase is used as the primary database for storing application data
- The file path is configurable via the `PAYERS_FILE_PATH` environment variable

## Network

- Both services are connected via the 'rental-network' Docker network
- Nginx handles proxying requests between the frontend and backend

## Running the Application

```bash
# Start both services
docker-compose up

# Access the web application
open http://localhost

# Access the API directly
curl http://localhost:8081/payers
``` 