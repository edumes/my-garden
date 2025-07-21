# Virtual Garden Management System API

A robust RESTful API for a virtual garden management game built with Go, featuring plant growth mechanics, weather systems, and real-time updates.

<img width="1740" height="820" alt="{180D47E5-04B8-434E-BC21-EE084CB99C26}" src="https://github.com/user-attachments/assets/e5a769e1-416d-490e-b497-1e311bdd0498" />

## Features

- 🌱 **Plant Management**: Plant, water, harvest, and manage various plant types
- 🌤️ **Weather System**: Dynamic weather affecting plant growth and garden conditions
- 👤 **User Authentication**: JWT-based authentication with role-based access
- 🤝 **Garden Sharing**: Share gardens with other users via access links with customizable permissions
- 🎮 **Game Mechanics**: Experience points, levels, achievements, and garden progression
- 🔄 **Real-time Updates**: WebSocket support for live garden updates
- 📊 **Analytics**: Garden statistics and performance tracking
- 🗄️ **Data Persistence**: PostgreSQL database with Redis caching
- 🚀 **Scalable Architecture**: Clean architecture with dependency injection
- 🔍 **Audit System**: Comprehensive logging and monitoring of all garden actions and system events

## Tech Stack

- **Language**: Go 1.21+
- **Framework**: Gin (HTTP router)
- **Database**: PostgreSQL with GORM
- **Cache**: Redis
- **Authentication**: JWT
- **Real-time**: WebSocket
- **Validation**: Go validator
- **Configuration**: Environment variables

## Quick Start

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 12 or higher
- Redis 6 or higher

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/edumes/my-garden.git
   cd my-garden
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up environment variables**
   ```bash
   cp env.example .env
   # Edit .env with your configuration
   ```

4. **Set up database**
   ```bash
   # Create PostgreSQL database
   createdb my_garden
   ```

5. **Run the application**
   ```bash
   go run cmd/server/main.go
   ```

The API will be available at `http://localhost:8080`

## API Endpoints

### Authentication
- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/refresh` - Refresh JWT token
- `POST /api/v1/auth/logout` - User logout

### Users
- `GET /api/v1/users/profile` - Get user profile
- `PUT /api/v1/users/profile` - Update user profile
- `GET /api/v1/users/achievements` - Get user achievements

### Gardens
- `GET /api/v1/gardens` - Get user gardens (owned and shared)
- `POST /api/v1/gardens` - Create new garden
- `GET /api/v1/gardens/{id}` - Get garden details
- `PUT /api/v1/gardens/{id}` - Update garden
- `DELETE /api/v1/gardens/{id}` - Delete garden

### Garden Sharing
- `POST /api/v1/garden-shares/access-links` - Create garden access link
- `POST /api/v1/garden-shares/join` - Join garden via access link
- `GET /api/v1/garden-shares/shared-with-me` - Get gardens shared with user
- `GET /api/v1/garden-shares/garden/{garden_id}` - Get garden shares
- `PUT /api/v1/garden-shares/garden/{garden_id}/permissions` - Update share permissions
- `DELETE /api/v1/garden-shares/garden/{garden_id}/user/{user_id}` - Remove garden share
- `GET /api/v1/garden-shares/garden/{garden_id}/access-links` - Get access links for garden
- `POST /api/v1/garden-shares/garden/{garden_id}/access-links/{link_id}/deactivate` - Deactivate access link

### Plants
- `GET /api/v1/plants` - Get available plant types
- `POST /api/v1/gardens/{id}/plants` - Plant in garden
- `PUT /api/v1/gardens/{id}/plants/{plantId}` - Update plant (water, fertilize)
- `DELETE /api/v1/gardens/{id}/plants/{plantId}` - Remove plant
- `POST /api/v1/gardens/{id}/plants/{plantId}/harvest` - Harvest plant

### Weather
- `GET /api/v1/weather/current` - Get current weather
- `GET /api/v1/weather/forecast` - Get weather forecast

### Game
- `GET /api/v1/game/status` - Get game status
- `POST /api/v1/game/actions` - Perform game actions

### Audit System
- `GET /api/v1/audit/logs` - Get audit logs with filtering
- `GET /api/v1/audit/users/{user_id}/activity` - Get user activity
- `GET /api/v1/audit/gardens/{garden_id}/activity` - Get garden activity
- `GET /api/v1/audit/stats` - Get audit statistics
- `GET /api/v1/game/leaderboard` - Get leaderboard

### WebSocket
- `WS /api/v1/ws/garden/{gardenId}` - Real-time garden updates

## Game Mechanics

### Plant Growth
- Plants have different growth stages: seed, sprout, growing, mature, harvestable
- Growth time varies by plant type and weather conditions
- Water and fertilizer affect growth speed and yield

### Weather System
- Dynamic weather affects plant growth rates
- Different weather conditions: sunny, cloudy, rainy, stormy
- Weather changes every 10 minutes

### Experience & Levels
- Users gain XP for various actions
- Leveling up unlocks new plant types and garden features
- Achievement system for milestones

### Garden Management
- Multiple garden plots per user
- Soil quality affects plant growth
- Garden tools and upgrades available

### Garden Sharing System
- **Access Links**: Create shareable links with customizable permissions
- **Permission Levels**: View, Plant, Harvest, and Manage permissions
- **Link Management**: Set expiration dates and usage limits
- **Shared Gardens**: Access and collaborate on gardens shared by other users
- **Permission Control**: Garden owners can manage user access and permissions

## Development

### Development with Live Reload (Air)

For the best development experience, use Air for automatic rebuilding and restarting when you make changes:

1. **Install Air** (if not already installed):
   ```bash
   go install github.com/cosmtrek/air@latest
   ```

2. **Start development server with live reload**:
   
   **On Unix/Linux/macOS:**
   ```bash
   make dev
   # or directly: air
   ```

Air will automatically:
- Watch for changes in your Go files
- Rebuild the application when changes are detected
- Restart the server automatically
- Show build errors in real-time

### Alternative Development Commands

**On Unix/Linux/macOS:**
```bash
# Install dependencies
make deps

# Build the application
make build

# Run the application (without live reload)
make run

# Run tests
make test

# Clean build artifacts
make clean
```

### Running Tests
```bash
go test ./...
```

### Code Formatting
```bash
go fmt ./...
```

### Linting
```bash
golangci-lint run
```

### Building
```bash
go build -o bin/server cmd/server/main.go
```

### Generate Swagger
```bash
swag init -g cmd/server/main.go -o docs
```
