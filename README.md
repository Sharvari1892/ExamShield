# 🛡️ ExamShield

A comprehensive exam proctoring and integrity monitoring platform designed to ensure secure and fair online examinations. ExamShield provides real-time monitoring, automated integrity checks, and detailed audit logs to maintain exam security.

> ⚠️ **Status**: This project is currently under active development. Features and documentation are subject to change.

## 📋 Table of Contents

- [Features](#features)
- [Tech Stack](#tech-stack)
- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Backend Setup](#backend-setup)
  - [Frontend Setup](#frontend-setup)
- [Configuration](#configuration)
- [Development](#development)
- [API Documentation](#api-documentation)
- [Roadmap](#roadmap)

## ✨ Features

### Core Functionality
- **🔐 Secure Authentication**: JWT-based authentication with secure password hashing
- **📝 Exam Management**: Create, manage, and publish exams with multiple question types
- **👥 User Management**: Role-based access control for students and administrators
- **⏱️ Timed Sessions**: Automatic session management with grace periods

### Integrity & Monitoring
- **🔍 Real-time Integrity Monitoring**: Continuous monitoring of exam sessions
- **⚠️ Automated Alerts**: Intelligent alert system for suspicious activities
- **📊 Risk Scoring**: Dynamic risk assessment based on user behavior
- **🔄 Integrity Worker**: Background workers for continuous integrity verification

### Advanced Features
- **📡 Real-time Updates**: WebSocket-based real-time communication for instant alerts
- **📈 Metrics & Analytics**: Prometheus integration for performance monitoring
- **🗂️ Audit Logging**: Comprehensive event logging for compliance and review
- **💾 Offline Support**: IndexedDB integration for offline exam capabilities
- **🔒 Data Synchronization**: Redis-backed synchronization for distributed systems
- **🚦 Rate Limiting**: Built-in rate limiting to prevent abuse

## 🛠️ Tech Stack

### Backend
- **Language**: Go 1.25.0
- **Framework**: Gin (HTTP web framework)
- **Database**: PostgreSQL (via pgx/v5)
- **Cache/PubSub**: Redis
- **Authentication**: JWT (golang-jwt/jwt/v5)
- **Logging**: Zap (structured logging)
- **Metrics**: Prometheus
- **WebSockets**: Gorilla WebSocket
- **Database Migrations**: SQL migrations

### Frontend
- **Framework**: React 19.2.0
- **Language**: TypeScript 5.9.3
- **Build Tool**: Vite 7.3.1
- **Routing**: React Router DOM 7.13.1
- **State Management**: Zustand 5.0.11
- **Data Fetching**: TanStack Query 5.90.21
- **HTTP Client**: Axios 1.13.5
- **Charts**: Recharts 3.7.0
- **Offline Storage**: IndexedDB (via idb 8.0.3)
- **Styling**: CSS modules

## 🏗️ Architecture

```
┌─────────────────┐         ┌──────────────────┐
│  React Frontend │◄────────┤   Gin REST API   │
│   (Vite/TS)     │         │   (Go Backend)   │
└────────┬────────┘         └────────┬─────────┘
         │                           │
         │                  ┌────────┴─────────┐
         │                  │                  │
         │              ┌───▼───┐        ┌────▼────┐
         │              │ Redis │        │ Postgres│
         │              │PubSub │        │   DB    │
         │              └───┬───┘        └─────────┘
         │                  │
    ┌────▼───────────────────▼────┐
    │   WebSocket Hub (Real-time) │
    │   Integrity Worker          │
    │   Metrics (Prometheus)      │
    └─────────────────────────────┘
```

### Key Components

- **API Server**: RESTful API with middleware for auth, logging, rate limiting, and CORS
- **WebSocket Hub**: Real-time bidirectional communication for instant updates
- **Integrity Worker**: Background service monitoring exam sessions for anomalies
- **Audit System**: Comprehensive event tracking and audit trail
- **Redis Subscriber**: Pub/Sub pattern for distributed real-time updates

## 📁 Project Structure

```
ExamShield/
├── backend/                    # Go backend service
│   ├── cmd/                   # Application entrypoints
│   │   ├── api/              # Main API server
│   │   └── worker/           # Background workers
│   ├── internal/             # Private application code
│   │   ├── config/          # Configuration management
│   │   ├── domain/          # Domain models (alerts, audit, sync)
│   │   ├── middleware/      # HTTP middleware (auth, logging, rate limiting)
│   │   ├── realtime/        # WebSocket hub and subscribers
│   │   ├── repository/      # Data access layer
│   │   ├── service/         # Business logic
│   │   ├── worker/          # Background job processors
│   │   ├── logger/          # Structured logging
│   │   └── metrics/         # Prometheus metrics
│   ├── migrations/          # Database migrations
│   ├── pkg/                 # Public packages
│   └── docker-compose.yml   # Docker services configuration
│
└── frontend/                   # React frontend application
    ├── src/
    │   ├── app/             # App-level configuration
    │   ├── components/      # Reusable UI components
    │   ├── features/        # Feature-based modules
    │   │   ├── auth/       # Authentication (login, registration)
    │   │   ├── exam/       # Exam taking interface
    │   │   ├── admin/      # Admin dashboard
    │   │   ├── integrity/  # Integrity monitoring views
    │   │   └── sync/       # Synchronization features
    │   ├── offline/         # IndexedDB offline storage
    │   ├── routes/          # Application routing
    │   ├── services/        # API services
    │   ├── store/           # State management (Zustand)
    │   ├── utils/           # Utility functions
    │   └── workers/         # Web workers
    └── public/              # Static assets
```

## 🚀 Getting Started

### Prerequisites

- **Go** 1.25.0 or higher
- **Node.js** 18+ and npm/yarn
- **PostgreSQL** 14+
- **Redis** 6+
- **Docker** (optional, for containerized development)

### Backend Setup

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd ExamShield/backend
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up environment variables**
   
   Create a `.env` file in the `backend` directory:
   ```env
   DB_URL=postgres://user:password@localhost:5432/examshield?sslmode=disable
   REDIS_ADDR=localhost:6379
   JWT_SECRET=your-secret-key-change-in-production
   PORT=8080
   ```

4. **Start PostgreSQL and Redis** (using Docker)
   ```bash
   docker-compose up -d
   ```

5. **Run database migrations**
   ```bash
   # Install migrate tool if not already installed
   go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
   
   # Run migrations
   migrate -path migrations -database "${DB_URL}" up
   ```

6. **Start the backend server**
   ```bash
   # Using Make
   make run
   
   # Or directly
   go run cmd/api/main.go
   ```

   The API server will start at `http://localhost:8080`

### Frontend Setup

1. **Navigate to the frontend directory**
   ```bash
   cd frontend
   ```

2. **Install dependencies**
   ```bash
   npm install
   # or
   yarn install
   ```

3. **Set up environment variables**
   
   Create a `.env` file in the `frontend` directory:
   ```env
   VITE_API_URL=http://localhost:8080
   ```

4. **Start the development server**
   ```bash
   npm run dev
   # or
   yarn dev
   ```

   The frontend will be available at `http://localhost:5173`

## ⚙️ Configuration

### Backend Configuration

The backend configuration is managed through environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_URL` | PostgreSQL connection string | Required |
| `REDIS_ADDR` | Redis server address | `localhost:6379` |
| `JWT_SECRET` | Secret key for JWT token signing | Required |
| `PORT` | API server port | `8080` |

### Frontend Configuration

Frontend configuration uses Vite environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `VITE_API_URL` | Backend API URL | `http://localhost:8080` |

## 💻 Development

### Running Tests

**Backend:**
```bash
cd backend
go test ./...

# With coverage
go test -cover ./...

# Specific package
go test ./internal/service/...
```

**Frontend:**
```bash
cd frontend
npm run test
```

### Code Quality

**Backend:**
```bash
# Format code
go fmt ./...

# Lint
golangci-lint run
```

**Frontend:**
```bash
# Lint
npm run lint

# Format with Prettier
npm run format
```

### Building for Production

**Backend:**
```bash
cd backend
make build
# Binary will be in ./bin/api
```

**Frontend:**
```bash
cd frontend
npm run build
# Production files will be in ./dist
```

## 📚 API Documentation

### Authentication Endpoints
- `POST /register` - Register a new user
- `POST /login` - Login and receive JWT token

### Exam Endpoints
- `GET /api/exams` - List available exams
- `GET /api/exams/:id` - Get exam details
- `POST /api/sessions` - Start an exam session
- `PUT /api/sessions/:id/answers` - Submit answers

### Admin Endpoints
- `POST /admin/exams` - Create new exam (requires admin role)
- `GET /admin/audit` - View audit logs (requires admin role)

### Monitoring
- `GET /health` - Health check endpoint
- `GET /metrics` - Prometheus metrics endpoint

### WebSocket
- `WS /ws` - Real-time updates and integrity alerts

## 🗺️ Roadmap

### Planned Features
- [ ] Video proctoring with facial recognition
- [ ] Screen recording and monitoring
- [ ] Advanced analytics dashboard
- [ ] Multi-language support
- [ ] Mobile application
- [ ] AI-powered cheating detection
- [ ] Integration with LMS platforms
- [ ] Advanced reporting and export features
- [ ] Two-factor authentication
- [ ] Browser lockdown mode

### In Progress
- [ ] Core exam functionality
- [ ] Integrity monitoring system
- [ ] Real-time alerts
- [ ] Admin dashboard

## 🤝 Contributing

This project is currently in development. Contribution guidelines will be added soon.

---

**Built with ❤️ using Go and React**
