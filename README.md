
# Sentinel

🚀 **Sentinel** is a modular, multi-tenant ready user management system written in Go. It supports:

**To watch over as a guard!**

- User registration
- Password hashing (bcrypt)
- JWT-based auth (coming soon)
- PostgreSQL-ready (migrations included)

## Getting Started

```bash
cd cmd/server
go run main.go
```


## API Endpoints

Here are the available endpoints for the Sentinel user management system:

### Health Check
```bash
curl -X GET http://localhost:8080/health
```
- **Description**: Check the health status of the server.
- **Method**: GET
- **Response**: `200 OK` if the server is running.

### User Registration
```bash
curl -X POST http://localhost:8080/register -H "Content-Type: application/json" -d '{"username":"example","password":"password123"}'
```
- **Description**: Register a new user.
- **Method**: POST
- **Request Body**:
  ```json
  {
    "username": "example",
    "password": "password123"
  }
  ```
- **Response**: `201 Created` on successful registration.

### User Login (Coming Soon)
- **Description**: Authenticate a user and return a JWT token.
- **Method**: POST
- **Endpoint**: `/login`

### User Info (Coming Soon)
- **Description**: Retrieve user information.
- **Method**: GET
- **Endpoint**: `/user/{id}`

