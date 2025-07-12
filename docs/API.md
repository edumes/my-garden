# Virtual Garden Management System API Documentation

## Overview

The Virtual Garden Management System API provides a comprehensive set of endpoints for managing virtual gardens, plants, weather, and game mechanics. The API is built with Go using the Gin framework and follows RESTful principles.

## Base URL

```
http://localhost:8080/api/v1
```

## Authentication

The API uses JWT (JSON Web Tokens) for authentication. Include the token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

## Response Format

All API responses follow this standard format:

```json
{
  "data": {...},
  "message": "Success message",
  "error": null
}
```

Error responses:

```json
{
  "error": "Error message",
  "data": null
}
```

## Endpoints

### Authentication

#### Register User
- **POST** `/auth/register`
- **Description**: Register a new user account
- **Request Body**:
```json
{
  "username": "gardener123",
  "email": "gardener@example.com",
  "password": "securepassword123",
  "first_name": "John",
  "last_name": "Doe"
}
```
- **Response**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "uuid",
    "username": "gardener123",
    "email": "gardener@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "level": 1,
    "experience": 0,
    "coins": 100,
    "created_at": "2024-01-01T00:00:00Z"
  },
  "expires_at": "2024-01-02T00:00:00Z"
}
```

#### Login User
- **POST** `/auth/login`
- **Description**: Authenticate user and get JWT token
- **Request Body**:
```json
{
  "username": "gardener123",
  "password": "securepassword123"
}
```
- **Response**: Same as register response

#### Refresh Token
- **POST** `/auth/refresh`
- **Description**: Refresh JWT token
- **Headers**: `Authorization: Bearer <current-token>`
- **Response**:
```json
{
  "token": "new-jwt-token",
  "expires_at": "2024-01-02T00:00:00Z"
}
```

#### Logout
- **POST** `/auth/logout`
- **Description**: Logout user (client-side token removal)
- **Headers**: `Authorization: Bearer <token>`
- **Response**:
```json
{
  "message": "Logged out successfully"
}
```

### User Management

#### Get User Profile
- **GET** `/users/profile`
- **Description**: Get current user's profile and achievements
- **Headers**: `Authorization: Bearer <token>`
- **Response**:
```json
{
  "user": {
    "id": "uuid",
    "username": "gardener123",
    "email": "gardener@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "level": 5,
    "experience": 450,
    "coins": 1250,
    "achievements": [
      {
        "achievement": {
          "name": "First Garden",
          "description": "Create your first garden",
          "icon": "🌱",
          "points": 10
        },
        "unlocked_at": "2024-01-01T10:00:00Z"
      }
    ]
  }
}
```

#### Update User Profile
- **PUT** `/users/profile`
- **Description**: Update user profile information
- **Headers**: `Authorization: Bearer <token>`
- **Request Body**:
```json
{
  "first_name": "Johnny",
  "last_name": "Smith",
  "avatar": "https://example.com/avatar.jpg",
  "timezone": "America/New_York",
  "language": "en"
}
```

### Garden Management

#### Get User Gardens
- **GET** `/gardens`
- **Description**: Get all gardens for the current user
- **Headers**: `Authorization: Bearer <token>`
- **Response**:
```json
{
  "gardens": [
    {
      "id": "uuid",
      "name": "My First Garden",
      "description": "A beautiful garden",
      "size": 9,
      "soil_quality": 75,
      "water_level": 60,
      "fertilizer_level": 20,
      "has_sprinkler": false,
      "has_greenhouse": false,
      "has_composter": false,
      "plants": [
        {
          "id": "uuid",
          "position": 0,
          "stage": "growing",
          "health": 85,
          "water_level": 70,
          "growth_progress": 65.5,
          "planted_at": "2024-01-01T09:00:00Z",
          "plant_type": {
            "name": "Tomato",
            "description": "A juicy red tomato",
            "icon": "🍅",
            "growth_time": 120,
            "harvest_value": 15
          }
        }
      ]
    }
  ]
}
```

#### Create Garden
- **POST** `/gardens`
- **Description**: Create a new garden
- **Headers**: `Authorization: Bearer <token>`
- **Request Body**:
```json
{
  "name": "Spring Garden",
  "description": "A garden for spring vegetables"
}
```

#### Get Garden Details
- **GET** `/gardens/{id}`
- **Description**: Get specific garden with all plants
- **Headers**: `Authorization: Bearer <token>`
- **Response**: Same as garden object in Get User Gardens

#### Update Garden
- **PUT** `/gardens/{id}`
- **Description**: Update garden information
- **Headers**: `Authorization: Bearer <token>`
- **Request Body**:
```json
{
  "name": "Updated Garden Name",
  "description": "Updated description"
}
```

#### Delete Garden
- **DELETE** `/gardens/{id}`
- **Description**: Delete a garden and all its plants
- **Headers**: `Authorization: Bearer <token>`
- **Response**:
```json
{
  "message": "Garden deleted successfully"
}
```

### Plant Management

#### Get Available Plant Types
- **GET** `/plants`
- **Description**: Get all available plant types
- **Response**:
```json
{
  "plant_types": [
    {
      "id": "uuid",
      "name": "Tomato",
      "description": "A juicy red tomato that grows well in warm weather",
      "icon": "🍅",
      "growth_time": 120,
      "water_needs": 60,
      "fertilizer_needs": 20,
      "yield": 3,
      "harvest_value": 15,
      "experience_value": 10,
      "min_level": 1,
      "season": "summer",
      "weather": "sunny",
      "rarity": "common"
    }
  ]
}
```

#### Plant Seed
- **POST** `/gardens/{id}/plants`
- **Description**: Plant a seed in a garden
- **Headers**: `Authorization: Bearer <token>`
- **Request Body**:
```json
{
  "plant_type_id": "uuid",
  "position": 0
}
```
- **Response**:
```json
{
  "plant": {
    "id": "uuid",
    "garden_id": "uuid",
    "plant_type_id": "uuid",
    "position": 0,
    "stage": "seed",
    "health": 100,
    "water_level": 50,
    "growth_progress": 0,
    "planted_at": "2024-01-01T10:00:00Z",
    "plant_type": {
      "name": "Tomato",
      "icon": "🍅"
    }
  }
}
```

#### Water Plant
- **PUT** `/gardens/{id}/plants/{plantId}`
- **Description**: Water a plant
- **Headers**: `Authorization: Bearer <token>`
- **Request Body**:
```json
{
  "amount": 30
}
```

#### Fertilize Plant
- **POST** `/gardens/{id}/plants/{plantId}/fertilize`
- **Description**: Apply fertilizer to a plant
- **Headers**: `Authorization: Bearer <token>`
- **Request Body**:
```json
{
  "amount": 20
}
```

#### Harvest Plant
- **POST** `/gardens/{id}/plants/{plantId}/harvest`
- **Description**: Harvest a mature plant
- **Headers**: `Authorization: Bearer <token>`
- **Response**:
```json
{
  "plant": {
    "id": "uuid",
    "stage": "withered",
    "harvested_at": "2024-01-01T12:00:00Z"
  },
  "harvest": {
    "coins_earned": 45,
    "experience_earned": 10,
    "level_up": true,
    "new_level": 6
  }
}
```

#### Remove Plant
- **DELETE** `/gardens/{id}/plants/{plantId}`
- **Description**: Remove a plant from the garden
- **Headers**: `Authorization: Bearer <token>`
- **Response**:
```json
{
  "message": "Plant removed successfully"
}
```

### Weather System

#### Get Current Weather
- **GET** `/weather/current`
- **Description**: Get current weather conditions
- **Response**:
```json
{
  "weather": {
    "id": "uuid",
    "condition": "sunny",
    "temperature": 25.5,
    "humidity": 45,
    "wind_speed": 8.2,
    "pressure": 1013.25,
    "growth_multiplier": 1.2,
    "water_evaporation_rate": 1.5,
    "created_at": "2024-01-01T10:00:00Z",
    "valid_until": "2024-01-01T10:10:00Z"
  },
  "season": "summer",
  "updated_at": "2024-01-01T10:00:00Z"
}
```

#### Get Weather Forecast
- **GET** `/weather/forecast`
- **Description**: Get weather predictions for next 24 hours
- **Response**:
```json
{
  "forecasts": [
    {
      "id": "uuid",
      "condition": "sunny",
      "temperature": 28.0,
      "humidity": 40,
      "probability": 85,
      "forecast_for": "2024-01-01T16:00:00Z"
    }
  ],
  "generated_at": "2024-01-01T10:00:00Z"
}
```

#### Get Weather History
- **GET** `/weather/history`
- **Description**: Get historical weather data
- **Response**:
```json
{
  "history": [
    {
      "id": "uuid",
      "condition": "cloudy",
      "temperature": 22.0,
      "humidity": 55,
      "created_at": "2024-01-01T09:50:00Z"
    }
  ],
  "count": 24
}
```

### Game Features

#### Get Game Status
- **GET** `/game/status`
- **Description**: Get current game status and statistics
- **Headers**: `Authorization: Bearer <token>` (optional)
- **Response**:
```json
{
  "server_time": "2024-01-01T10:00:00Z",
  "season": "summer",
  "weather": {
    "condition": "sunny",
    "temperature": 25.5
  },
  "user_stats": {
    "level": 5,
    "experience": 450,
    "coins": 1250,
    "gardens_count": 2,
    "plants_count": 8
  }
}
```

#### Get Leaderboard
- **GET** `/game/leaderboard`
- **Description**: Get top players by experience
- **Query Parameters**:
  - `limit`: Number of players to return (default: 10)
  - `category`: Ranking category (level, experience, coins)
- **Response**:
```json
{
  "leaderboard": [
    {
      "rank": 1,
      "username": "master_gardener",
      "level": 25,
      "experience": 12500,
      "coins": 5000
    }
  ],
  "category": "level",
  "updated_at": "2024-01-01T10:00:00Z"
}
```

### WebSocket Endpoints

#### Garden Real-time Updates
- **WebSocket** `/ws/garden/{gardenId}`
- **Description**: Real-time updates for garden changes
- **Headers**: `Authorization: Bearer <token>`
- **Message Types**:
  - `plant_growth`: Plant growth progress updates
  - `weather_change`: Weather condition changes
  - `harvest_ready`: Plant ready for harvest
  - `plant_withered`: Plant has withered

Example WebSocket message:
```json
{
  "type": "plant_growth",
  "data": {
    "plant_id": "uuid",
    "stage": "mature",
    "growth_progress": 85.5,
    "water_level": 65
  }
}
```

## Error Codes

| Status Code | Description |
|-------------|-------------|
| 200 | Success |
| 201 | Created |
| 400 | Bad Request - Invalid input data |
| 401 | Unauthorized - Invalid or missing token |
| 403 | Forbidden - Insufficient permissions |
| 404 | Not Found - Resource not found |
| 409 | Conflict - Resource already exists |
| 422 | Unprocessable Entity - Validation error |
| 500 | Internal Server Error |

## Rate Limiting

The API implements rate limiting to prevent abuse:
- 100 requests per minute per IP address
- Rate limit headers included in responses:
  - `X-RateLimit-Limit`: Request limit
  - `X-RateLimit-Remaining`: Remaining requests
  - `X-RateLimit-Reset`: Reset time

## Game Mechanics

### Plant Growth Stages
1. **Seed** (0-20% growth): Just planted
2. **Sprout** (20-40% growth): First leaves appear
3. **Growing** (40-70% growth): Plant is developing
4. **Mature** (70-100% growth): Plant is fully grown
5. **Harvestable** (100% growth): Ready for harvest
6. **Withered** (0% health): Plant has died

### Weather Effects
- **Sunny**: +20% growth, +50% water evaporation
- **Cloudy**: Normal growth and evaporation
- **Rainy**: +10% growth, -70% water evaporation
- **Stormy**: -20% growth, -90% water evaporation
- **Foggy**: -10% growth, -20% water evaporation
- **Windy**: -5% growth, +30% water evaporation
- **Snowy**: -50% growth, -80% water evaporation

### Experience System
- Planting: 5 XP
- Watering: 1 XP
- Fertilizing: 2 XP
- Harvesting: Varies by plant type (5-50 XP)
- Level up: Every 100 XP

### Seasons
- **Spring** (Mar-May): Moderate temperatures, varied weather
- **Summer** (Jun-Aug): High temperatures, mostly sunny
- **Autumn** (Sep-Nov): Moderate temperatures, rainy
- **Winter** (Dec-Feb): Low temperatures, snowy/cloudy

## Development

### Running the API

1. Set up environment variables:
```bash
cp env.example .env
# Edit .env with your configuration
```

2. Start the server:
```bash
go run cmd/server/main.go
```

3. The API will be available at `http://localhost:8080`

### Testing

Use the provided endpoints with tools like:
- **Postman**: For API testing
- **curl**: For command-line testing
- **WebSocket clients**: For real-time features

Example curl commands:
```bash
# Register user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","email":"test@example.com","password":"password123"}'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"password123"}'

# Get gardens (with token)
curl -X GET http://localhost:8080/api/v1/gardens \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
``` 