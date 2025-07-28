# Virtual Garden Management System API - Setup Guide

## Quick Start

### Prerequisites

1. **Go 1.21 or higher**
   ```bash
   go version
   ```

2. **PostgreSQL 12 or higher**
   - Install PostgreSQL from https://www.postgresql.org/download/
   - Create a database named `my_garden`

3. **Redis 6 or higher** (optional, for caching)
   - Install Redis from https://redis.io/download
   - Or use Docker: `docker run -d -p 6379:6379 redis:alpine`

### Installation Steps

1. **Clone and navigate to the project**
   ```bash
   cd my-garden
   ```

2. **Install dependencies**
   ```bash
   go mod tidy
   ```

3. **Set up environment variables**
   ```bash
   cp env.example .env
   ```
   
   Edit `.env` with your database credentials:
   ```env
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=postgres
   DB_PASSWORD=your_password
   DB_NAME=my_garden
   ```

4. **Build the application**
   ```bash
   go build -o bin/server cmd/server/main.go
   ```

5. **Run the server**
   ```bash
   ./bin/server
   # Or on Windows:
   bin\server.exe
   ```

The API will be available at `http://localhost:8080`

## Testing the API

### 1. Health Check
```bash
curl http://localhost:8080/health
```

### 2. Register a User
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123",
    "first_name": "Test",
    "last_name": "User"
  }'
```

### 3. Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123"
  }'
```

### 4. Create a Garden (use token from login)
```bash
curl -X POST http://localhost:8080/api/v1/gardens \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -d '{
    "name": "My First Garden",
    "description": "A beautiful garden"
  }'
```

### 5. Check Weather
```bash
curl http://localhost:8080/api/v1/weather/current
```

## Project Structure

```
my-garden/
├── cmd/
│   └── server/
│       └── main.go                # Application entry point
├── internal/
│   ├── audit/
│   │   ├── handler.go             # Audit log handlers
│   │   ├── middleware.go
│   │   ├── model.go
│   │   ├── repository.go
│   │   ├── service.go
│   │   └── types.go
│   ├── auth/
│   │   ├── handler.go             # Auth handlers
│   │   ├── middleware.go
│   │   ├── model.go
│   │   ├── repository.go
│   │   └── service.go
│   ├── config/
│   │   └── config.go              # Configuration management
│   ├── database/
│   │   └── database.go            # Database connection
│   ├── garden/
│   │   ├── handler.go             # Garden handlers
│   │   ├── model.go
│   │   └── share_handler.go
│   ├── middleware/
│   │   └── cors.go                # CORS middleware
│   ├── migrations/
│   │   └── migrations.go          # DB migrations
│   ├── store/
│   │   ├── handler.go             # Store handlers
│   │   ├── repository.go
│   │   └── service.go
│   └── weather/
│       ├── handler.go             # Weather handlers
│       ├── model.go
│       ├── repository.go
│       └── service.go
├── pkg/
│   ├── auth/
│   │   └── jwt.go                 # JWT token management
│   ├── db/
│   │   └── database.go            # DB helpers
│   ├── game/
│   │   └── engine.go              # Game mechanics engine
│   └── utils/
│       └── math.go                # Utility functions
├── frontend/
│   ├── src/
│   │   ├── App.tsx                # Main React app
│   │   ├── main.tsx
│   │   ├── index.css
│   │   ├── components/
│   │   │   ├── (many component files)
│   │   │   ├── ui/
│   │   │   │   ├── button.tsx, card.tsx, ...
│   │   │   └── animate-ui/
│   │   │       ├── base/
│   │   │       ├── components/
│   │   │       ├── effects/
│   │   │       └── ui-elements/
│   │   ├── contexts/
│   │   │   ├── AuthContext.tsx
│   │   │   └── ThemeContext.tsx
│   │   ├── services/
│   │   │   └── api.ts
│   │   ├── types/
│   │   │   └── api.ts
│   │   └── lib/
│   │       └── utils.ts
│   ├── package.json
│   ├── vite.config.ts
│   └── ... (other config files)
├── docs/
│   ├── API.md                     # API documentation
│   ├── AUDIT_SYSTEM.md
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── tmp/
│   ├── main.exe                   # Compiled binary (example)
│   └── build-errors.log
├── go.mod                         # Go module file
├── go.sum                         # Dependency checksums
├── env.example                    # Environment variables template
├── README.md                      # Project overview
└── SETUP.md                       # This file
```

## Features Implemented

### ✅ Core Features
- **User Authentication**: JWT-based registration, login, and token management
- **Garden Management**: Create, read, update, delete gardens
- **Plant System**: Plant seeds, water, fertilize, and harvest plants
- **Weather System**: Dynamic weather affecting plant growth
- **Game Engine**: Real-time plant growth and weather updates
- **Database**: PostgreSQL with GORM ORM
- **Caching**: Redis integration (optional)

### ✅ Game Mechanics
- **Plant Growth Stages**: Seed → Sprout → Growing → Mature → Harvestable → Withered
- **Weather Effects**: Different weather conditions affect growth rates
- **Experience System**: Users gain XP for various actions
- **Leveling System**: Level up based on experience points
- **Seasons**: Dynamic seasons affecting weather patterns

### ✅ API Endpoints
- **Authentication**: Register, login, refresh, logout
- **Users**: Profile management
- **Gardens**: CRUD operations for gardens
- **Plants**: Plant management and care
- **Weather**: Current weather, forecast, history
- **Game**: Status and leaderboard

### 🔄 Real-time Features (Planned)
- WebSocket support for live garden updates
- Real-time weather changes
- Plant growth notifications

## Configuration Options

Key environment variables in `.env`:

```env
# Server
PORT=8080
ENV=development

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=my_garden

# Redis (optional)
REDIS_HOST=localhost
REDIS_PORT=6379

# JWT
JWT_SECRET=your-super-secret-jwt-key
JWT_EXPIRY=24h

# Game Settings
GAME_TICK_INTERVAL=300s      # 5 minutes
WEATHER_UPDATE_INTERVAL=600s # 10 minutes
PLANT_GROWTH_INTERVAL=900s   # 15 minutes
```

## Development

### Running in Development Mode
```bash
go run cmd/server/main.go
```

### Running Tests
```bash
go test ./...
```

### Code Formatting
```bash
go fmt ./...
```

### Building for Production
```bash
go build -ldflags="-s -w" -o bin/server cmd/server/main.go
```

## Troubleshooting

### Common Issues

1. **Database Connection Error**
   - Ensure PostgreSQL is running
   - Check database credentials in `.env`
   - Verify database `my_garden` exists

2. **Port Already in Use**
   - Change `PORT` in `.env` file
   - Or kill the process using the port

3. **Redis Connection Error**
   - Redis is optional, the app will work without it
   - Check Redis is running if you want caching

4. **Permission Denied**
   - On Windows, run PowerShell as Administrator
   - On Linux/Mac, check file permissions

### Logs

The application logs to stdout. Look for:
- Database connection messages
- Game engine startup
- Weather updates
- Plant growth processing

## Next Steps

1. **Frontend Development**: Build a web or mobile client
2. **Additional Features**: 
   - Garden tools and upgrades
   - More plant types
   - Social features (friends, trading)
   - Advanced weather patterns
3. **Deployment**: Deploy to cloud platform
4. **Monitoring**: Add metrics and monitoring
5. **Testing**: Comprehensive test suite

## Support

For issues and questions:
1. Check the API documentation in `docs/API.md`
2. Review the code structure
3. Check the logs for error messages
4. Verify your configuration

The API is production-ready with proper error handling, authentication, and scalable architecture! 