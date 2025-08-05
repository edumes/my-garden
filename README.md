# Virtual Garden Management System

A full-stack virtual garden management game featuring a robust Go backend API and modern React frontend. Players can plant, grow, harvest, and share gardens while experiencing dynamic weather systems and real-time updates.

![Virtual Garden System](https://github.com/user-attachments/assets/e5a769e1-416d-490e-b497-1e311bdd0498)

## 🚀 Quick Start

### Prerequisites

- **Go** 1.21 or higher
- **Node.js** 18 or higher
- **PostgreSQL** 12 or higher
- **Redis** 6 or higher

### One-Command Setup

```bash
# Clone and setup
git clone https://github.com/edumes/my-garden.git
cd my-garden

# Setup environment
cp env.example .env
# Edit .env with your database credentials

# Install dependencies and start
make setup
make dev
```

The application will be available at:
- **Backend API**: http://localhost:8080
- **Frontend**: http://localhost:5173

## 🏗️ Architecture Overview

This is a **full-stack application** with a clean separation between frontend and backend:

### Backend (Go)
- **Framework**: Gin HTTP router
- **Database**: PostgreSQL with GORM ORM
- **Cache**: Redis for session management
- **Authentication**: JWT-based with role-based access
- **Real-time**: WebSocket support
- **Documentation**: Swagger/OpenAPI

### Frontend (React + TypeScript)
- **Framework**: React 18 with TypeScript
- **Styling**: Tailwind CSS with shadcn/ui components
- **State Management**: React Context API
- **Build Tool**: Vite
- **UI Components**: Custom component library

## 📁 Project Structure

```
my-garden/
├── cmd/server/           # Application entry point
│   └── main.go          # Main server file
├── internal/             # Private application code
│   ├── auth/            # Authentication & authorization
│   ├── garden/          # Garden management logic
│   ├── marketplace/     # Marketplace features
│   ├── weather/         # Weather system
│   ├── audit/           # Audit logging system
│   ├── store/           # Store/inventory system
│   ├── config/          # Configuration management
│   ├── database/        # Database connection & migrations
│   └── middleware/      # HTTP middleware
├── pkg/                 # Public packages (reusable)
│   ├── auth/            # JWT utilities
│   ├── db/              # Database utilities
│   ├── game/            # Game engine logic
│   └── utils/           # Common utilities
├── frontend/            # React frontend application
│   ├── src/
│   │   ├── components/  # React components
│   │   ├── contexts/    # React contexts
│   │   ├── services/    # API services
│   │   ├── types/       # TypeScript types
│   │   └── lib/         # Utility functions
│   └── package.json
├── docs/                # API documentation
├── migrations/          # Database migrations
└── Makefile            # Build automation
```

## 🎮 Features

### Core Features
- **Garden Management**: Create, plant, water, and harvest virtual gardens
- **Weather System**: Dynamic weather affecting plant growth
- **Marketplace**: Buy/sell plants and delivery requests
- **Social Features**: Share gardens with other users
- **Audit System**: Comprehensive activity logging
- **Settings Management**: User profile and preference management

### Settings Page
The application includes a comprehensive settings page (`/settings`) with the following features:

- **Profile Management**: Update first name, last name, timezone, and language
- **Account Information**: View username, email, level, experience, and coins (read-only)
- **Preferences**: Configure timezone and language settings
- **Security**: Password change, 2FA, and account deletion options
- **Notifications**: Configure email and push notification preferences
- **Responsive Design**: Works seamlessly on desktop and mobile devices

## 🛠️ Development Setup

### 1. Environment Configuration

```bash
# Copy environment template
cp env.example .env

# Edit .env with your settings
DATABASE_URL=postgres://username:password@localhost:5432/my_garden
REDIS_URL=redis://localhost:6379
JWT_SECRET=your-secret-key
```

### 2. Database Setup

```bash
# Create PostgreSQL database
createdb my_garden

# Run migrations (if using a migration tool)
# The application will auto-migrate on startup
```

### 3. Install Dependencies

```bash
# Backend dependencies
go mod download

# Frontend dependencies
cd frontend
npm install
cd ..
```

## 🚀 Development Commands

### Backend Development

```bash
# Development with live reload (recommended)
make dev

# Alternative: direct air command
air

# Run without live reload
make run

# Run tests
make test

# Build for production
make build

# Clean build artifacts
make clean
```

### Frontend Development

```bash
# Start frontend development server
cd frontend
npm run dev

# Build for production
npm run build

# Run tests
npm test
```

### Database & Migrations

```bash
# Generate new migration
make migrate-create name=migration_name

# Run migrations
make migrate-up

# Rollback migrations
make migrate-down
```

### Code Quality

```bash
# Format Go code
make fmt

# Lint Go code
make lint

# Format frontend code
cd frontend && npm run format

# Lint frontend code
cd frontend && npm run lint
```

### Documentation

```bash
# Generate Swagger documentation
make swagger

# View API documentation
# Open docs/swagger.json in Swagger UI
```

## 🎮 Core Features

### Plant Management
- **Growth Stages**: Seed → Sprout → Growing → Mature → Harvestable
- **Plant Types**: Various plants with different growth times and yields
- **Care System**: Watering, fertilizing, and maintenance
- **Weather Impact**: Dynamic weather affects growth rates

### Garden System
- **Multiple Gardens**: Users can own and manage multiple garden plots
- **Soil Quality**: Different soil types affect plant growth
- **Garden Tools**: Upgrades and tools for better garden management
- **Real-time Updates**: WebSocket-based live garden updates

### Social Features
- **Garden Sharing**: Create shareable access links with customizable permissions
- **Permission Levels**: View, Plant, Harvest, and Manage permissions
- **Collaboration**: Multiple users can work on shared gardens
- **Access Control**: Set expiration dates and usage limits

### Weather System
- **Dynamic Weather**: Changes every 10 minutes
- **Weather Types**: Sunny, cloudy, rainy, stormy
- **Plant Impact**: Weather affects growth rates and plant health
- **Forecasting**: Weather predictions for planning

### Game Mechanics
- **Experience System**: Gain XP for various actions
- **Leveling**: Unlock new features and plant types
- **Achievements**: Milestone-based achievement system
- **Leaderboards**: Competitive rankings

### Audit System
- **Comprehensive Logging**: Track all user actions and system events
- **Activity Monitoring**: User and garden activity tracking
- **Security**: Detailed audit trails for security compliance
- **Analytics**: System statistics and performance metrics

## 🔌 API Endpoints

### Authentication
```http
POST /api/v1/auth/register     # Register new user
POST /api/v1/auth/login        # User login
POST /api/v1/auth/refresh      # Refresh JWT token
POST /api/v1/auth/logout       # User logout
```

### Gardens
```http
GET    /api/v1/gardens                    # Get user gardens
POST   /api/v1/gardens                    # Create new garden
GET    /api/v1/gardens/{id}              # Get garden details
PUT    /api/v1/gardens/{id}              # Update garden
DELETE /api/v1/gardens/{id}              # Delete garden
```

### Plants
```http
GET    /api/v1/plants                                    # Get available plants
POST   /api/v1/gardens/{id}/plants                      # Plant in garden
PUT    /api/v1/gardens/{id}/plants/{plantId}           # Update plant
DELETE /api/v1/gardens/{id}/plants/{plantId}           # Remove plant
POST   /api/v1/gardens/{id}/plants/{plantId}/harvest   # Harvest plant
```

### Garden Sharing
```http
POST   /api/v1/garden-shares/access-links              # Create access link
POST   /api/v1/garden-shares/join                      # Join garden
GET    /api/v1/garden-shares/shared-with-me           # Get shared gardens
PUT    /api/v1/garden-shares/garden/{id}/permissions  # Update permissions
```

### Weather & Game
```http
GET /api/v1/weather/current    # Current weather
GET /api/v1/weather/forecast   # Weather forecast
GET /api/v1/game/status        # Game status
GET /api/v1/game/leaderboard   # Leaderboard
```

### Audit & Monitoring
```http
GET /api/v1/audit/logs                    # Audit logs
GET /api/v1/audit/users/{id}/activity     # User activity
GET /api/v1/audit/gardens/{id}/activity   # Garden activity
GET /api/v1/audit/stats                   # Audit statistics
```

### WebSocket
```http
WS /api/v1/ws/garden/{gardenId}  # Real-time garden updates
```

## 💻 Code Style Guidelines

### Go Backend

#### File Organization
- Use **snake_case** for file names
- Group related functionality in packages
- Keep packages focused and cohesive

#### Naming Conventions
```go
// Variables and functions: camelCase
userName := "john"
func getUserProfile() {}

// Constants: UPPER_SNAKE_CASE
const MAX_PLANTS_PER_GARDEN = 100

// Types and interfaces: PascalCase
type Garden struct {}
type PlantService interface {}
```

#### Error Handling
```go
// Always check errors
if err != nil {
    return fmt.Errorf("failed to create garden: %w", err)
}

// Use custom error types for specific errors
var ErrGardenNotFound = errors.New("garden not found")
```

#### API Response Structure
```go
type APIResponse struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
}
```

### Frontend (TypeScript/React)

#### File Organization
- Use **PascalCase** for component files
- Use **camelCase** for utility files
- Group related components in folders

#### Component Structure
```typescript
// Component naming: PascalCase
export const GardenCard: React.FC<GardenCardProps> = ({ garden }) => {
  // Hooks first
  const [isLoading, setIsLoading] = useState(false);
  
  // Event handlers
  const handleClick = useCallback(() => {
    // handler logic
  }, []);
  
  // Render
  return (
    <div className="garden-card">
      {/* JSX */}
    </div>
  );
};
```

#### TypeScript Guidelines
```typescript
// Use interfaces for object shapes
interface Garden {
  id: string;
  name: string;
  plants: Plant[];
}

// Use type aliases for unions
type Permission = 'view' | 'plant' | 'harvest' | 'manage';

// Use enums sparingly
enum WeatherType {
  SUNNY = 'sunny',
  CLOUDY = 'cloudy',
  RAINY = 'rainy'
}
```

## 🧪 Testing

### Backend Testing
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/garden

# Run tests with verbose output
go test -v ./...
```

### Frontend Testing
```bash
cd frontend

# Run tests
npm test

# Run tests with coverage
npm run test:coverage

# Run tests in watch mode
npm run test:watch
```

## 🚀 Deployment

### Backend Deployment
```bash
# Build for production
make build

# Run with environment variables
./bin/server
```

### Frontend Deployment
```bash
cd frontend

# Build for production
npm run build

# Serve static files
npm run preview
```

### Docker Deployment
```bash
# Build and run with Docker
docker-compose up -d
```

## 📚 Documentation

- **API Documentation**: Available at `/docs/swagger.json`
- **Code Comments**: Follow Go documentation standards
- **README**: This file for project overview
- **SETUP.md**: Detailed setup instructions

## 🤝 Contributing

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Commit** your changes (`git commit -m 'Add amazing feature'`)
4. **Push** to the branch (`git push origin feature/amazing-feature`)
5. **Open** a Pull Request

### Development Workflow
1. Ensure all tests pass
2. Follow code style guidelines
3. Add tests for new features
4. Update documentation as needed
5. Squash commits before merging

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- **Issues**: Create an issue on GitHub
- **Discussions**: Use GitHub Discussions for questions
- **Documentation**: Check the `/docs` folder for detailed guides

---

**Happy Gardening! 🌱**
