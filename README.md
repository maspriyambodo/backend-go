# AdminBE

A robust backend administration API built with Go and Gin framework, providing comprehensive user management, role-based access control (RBAC), menu navigation, and audit logging capabilities.

## Features

- 🔐 JWT-based authentication system
- 👥 User management with CRUD operations
- 🏷️ Role-based access control with inheritance
- 📱 Dynamic menu system with navigation hierarchy
- 🔗 Permissions management (User-Role, Role-Menu associations)
- 📊 Audit logging for all operations
- 🏥 Health check endpoints
- 🔄 CORS support
- 📖 RESTful API design
- 🗄️ MySQL database with GORM ORM
- ⚡ Redis caching support

## Prerequisites

- Go 1.19 or higher (for local development)
- Docker and Docker Compose (recommended for running the application)
- MySQL 5.7+ or MariaDB 10.0+ (if running services locally)
- Redis 6.0+ (if running services locally)
- Git

## Installation

1. Clone the repository:
```bash
git clone https://github.com/maspriyambodo/backend-go.git
cd adminbe
```

2. Install dependencies:
```bash
go mod download
```

3. Set up the database:
   - Create a MySQL database named `adminbe`
   - Run migrations (if any) or set up schema manually

4. Generate JWT secret:
```bash
go run ./cmd/secret
```
Copy the generated secret and set it as an environment variable.

## Configuration

### Environment Variables

Create a `.env` file in the root directory or set environment variables:

```env
# Server Configuration
PORT=8080

# Database Configuration
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=adminbe

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT Configuration
JWT_SECRET=your_generated_secret_key_here
JWT_EXPIRATION=24h

# Server Mode (debug/release/test)
GIN_MODE=release
```

### Config File

Alternatively, update `configs/config.yaml` with your settings:

```yaml
server:
  port: 8080
  mode: release

database:
  host: localhost
  port: 3306
  username: root
  password: "your_password"
  database: adminbe
  charset: utf8mb4
  parseTime: True
  loc: Local

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

jwt:
  secret: "your_secret_here"
  expiration: 24h

cors:
  allow_origins: ["*"]
  allow_methods: ["GET", "POST", "PUT", "DELETE"]
  allow_headers: ["Authorization", "Content-Type"]
```

## Running the Application

### Option 1: Docker Compose (Recommended)

#### Development Environment
```bash
# Start the development environment with Docker Compose
docker-compose -f docker-compose.dev.yml up -d

# View logs
docker-compose -f docker-compose.dev.yml logs -f app

# Stop the development environment
docker-compose -f docker-compose.dev.yml down
```

The development environment includes:
- Auto-rebuilding on code changes (via volume mounts)
- Local MySQL and Redis instances
- Pre-configured database schema and sample data
- Debug logging enabled
- Access at http://localhost:8080

Default login credentials:
- Username: `admin`
- Password: `admin123`

#### Production Environment
```bash
# 1. Copy and configure production environment
cp .env.prod .env.production
# Edit .env.production with your production values

# 2. Build the production image
docker build -t adminbe:latest .

# 3. Start production environment
docker-compose -f docker-compose.prod.yml --env-file .env.production up -d

# 4. View logs
docker-compose -f docker-compose.prod.yml logs -f app

# 5. Stop production environment
docker-compose -f docker-compose.prod.yml down
```

**⚠️ Production Security Notes:**
- Change all default passwords and secrets in `.env.prod`
- Generate a strong JWT secret (minimum 32 characters)
- Configure proper firewall rules
- Use HTTPS in production (consider reverse proxy like nginx)
- Regularly update Docker images for security patches

### Option 2: Local Development

1. Ensure you have MySQL and Redis running locally

2. Build the application:
```bash
chmod +x scripts/build.sh
./scripts/build.sh
```

3. Configure your local environment:
   - Set database connection strings in `.env` or environment variables
   - Or modify `configs/config.yaml` with your local settings

4. Run the server:
```bash
./bin/adminbe
```

Or run directly with Go:
```bash
go run ./cmd/server
```

The server will start on port 8080 by default (configurable via PORT environment variable).

## API Documentation

### Authentication

#### Login
```http
POST /api/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "password"
}
```

Response:
```json
{
  "token": "jwt_token_here",
  "user": {
    "id": 1,
    "username": "admin",
    "email": "admin@example.com"
  }
}
```

### Health Check

#### Ping
```http
GET /ping
```

#### Health Status
```http
GET /health
```

### Protected Endpoints (Require JWT token in Authorization header)

All API endpoints require `Bearer <jwt_token>` in the Authorization header.

#### Users Management
- `GET /api/users` - List all users
- `GET /api/users/:id` - Get user by ID
- `POST /api/users` - Create new user
- `PUT /api/users/:id` - Update user
- `DELETE /api/users/:id` - Delete user

#### Roles Management
- `GET /api/roles` - List all roles
- `GET /api/roles/:id` - Get role by ID
- `POST /api/roles` - Create new role
- `PUT /api/roles/:id` - Update role
- `DELETE /api/roles/:id` - Delete role

#### Role Inheritances
- `GET /api/role_inheritances` - List role inheritance relationships
- `POST /api/role_inheritances` - Create role inheritance
- `PUT /api/role_inheritances/:id` - Update inheritance
- `DELETE /api/role_inheritances/:id` - Delete inheritance

#### Virtual Roles (Role Hierarchy View)
- `GET /api/v_roles` - Get flattened role hierarchy

#### Menu Management
- `GET /api/menu` - List all menu items
- `GET /api/menu/:id` - Get menu item by ID
- `POST /api/menu` - Create menu item
- `PUT /api/menu/:id` - Update menu item
- `DELETE /api/menu/:id` - Delete menu item

#### Menu Navigation (Menu Tree View)
- `GET /api/menu_navigation` - Get menu hierarchy tree

#### Role-Menu Permissions
- `GET /api/role_menu` - List role-menu associations
- `GET /api/role_menu/:roleId/:menuId` - Get specific association
- `POST /api/role_menu` - Create role-menu association
- `PUT /api/role_menu/:roleId/:menuId` - Update association
- `DELETE /api/role_menu/:roleId/:menuId` - Delete association

#### User-Role Assignments
- `GET /api/user_roles` - List user-role associations
- `GET /api/user_roles/:userId/:roleId` - Get specific association
- `POST /api/user_roles` - Create user-role association
- `PUT /api/user_roles/:userId/:roleId` - Update association
- `DELETE /api/user_roles/:userId/:roleId` - Delete association

#### User-Menu Permissions
- `GET /api/user_menu` - List user-menu associations
- `GET /api/user_menu/:userId/:menuId` - Get specific association
- `POST /api/user_menu` - Create user-menu association
- `PUT /api/user_menu/:userId/:menuId` - Update association
- `DELETE /api/user_menu/:userId/:menuId` - Delete association

#### Audit Logs
- `GET /api/audit_logs` - List all audit logs
- `GET /api/audit_logs/:id` - Get audit log by ID
- `POST /api/audit_logs` - Create audit log entry
- `PUT /api/audit_logs/:id` - Update audit log
- `DELETE /api/audit_logs/:id` - Delete audit log

## Development

### Project Structure
```
adminbe/
├── cmd/
│   ├── server/           # Main API server entry point
│   └── secret/           # JWT secret generator utility
├── configs/              # Configuration files
├── docs/                 # Documentation
├── internal/
│   ├── app/
│   │   ├── handlers/     # HTTP request handlers
│   │   ├── middleware/   # Custom middleware
│   │   └── models/       # Data models
│   ├── migrations/       # Database migrations
│   └── pkg/
│       ├── database/     # Database connection setup
│       └── utils/        # Utility functions
├── pkg/                  # Shared packages
└── scripts/              # Build and deployment scripts
```

### Code Quality

- Run tests:
```bash
go test ./...
```

- Format code:
```bash
go fmt ./...
```

- Run linter (if available):
```bash
golangci-lint run
```

## Deployment

### Docker (if applicable)
```bash
# Build Docker image
docker build -t adminbe .

# Run with Docker
docker run -p 8080:8080 \
  -e DB_HOST=your_db_host \
  -e JWT_SECRET=your_secret \
  adminbe
```

### Production Considerations
- Change JWT secret to a secure random value
- Use environment variables instead of config file
- Set up proper database connection pooling
- Configure reverse proxy (nginx) for production
- Implement proper logging and monitoring
- Set up SSL/TLS certificates

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Contact

Mas Priyambodo - [GitHub](https://github.com/maspriyambodo)

Project Link: [https://github.com/maspriyambodo/backend-go](https://github.com/maspriyambodo/backend-go)
