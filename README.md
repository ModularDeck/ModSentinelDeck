# Sentinel

🚀 **Sentinel** is a modular, multi-tenant ready user management system written in Go. It acts as a guardian for your user data.

*"To watch over as a guard!"*

---

## 🔑 Key Features

- User registration
- Password hashing (bcrypt)
- JWT-based authentication
- PostgreSQL support with database migrations

---

## 🚀 Getting Started

To get the Sentinel server running, follow these steps:

```bash
# Build the project
make build

# Or, to clean previous builds and build everything:
make clean

# To run the server (after building)
make all
```

**Note:** Ensure you have Go and PostgreSQL set up and configured as required by the project.

---

## 📡 API Endpoints

### 🔍 Health Check

```bash
curl -X GET http://localhost:8080/health
```

- **Description:** Checks the operational status of the Sentinel server.
- **Method:** `GET`
- **Endpoint:** `/health`
- **Response:** `200 OK` if the server is running.

---

### 📝 User Registration

```bash
curl -X POST http://localhost:8080/register \
-H "Content-Type: application/json" \
-d '{
  "name": "John Doe",
  "email": "user3@example.com",
  "password": "securepassword",
  "tenant_name": "TenantName",
  "tenant_desc": "Description of the firm",
  "team_name": "Example Team",
  "team_desc": "Description of the team",
  "user_role": "admin",
  "team_role": "admin"
}'
```

- **Description:** Registers a new user. Optionally creates a tenant and team if they don't exist.
- **Method:** `POST`
- **Endpoint:** `/register`
- **Headers:**
  - `Content-Type: application/json`
- **Request Body (JSON):**
  ```json
  {
    "name": "John Doe",
    "email": "user3@example.com",
    "password": "securepassword",
    "tenant_name": "TenantName",
    "tenant_desc": "Description of the firm",
    "team_name": "Example Team",
    "team_desc": "Description of the team",
    "user_role": "admin",
    "team_role": "admin"
  }
  ```
- **Response:** `201 Created`
- **Response Example:**
  ```json
  {
    "message": "User created successfully with tenant and team",
    "team_id": 1,
    "tenant_id": 1,
    "user_id": 1
  }
  ```

---

### 🔐 User Login

```bash
curl -X POST http://localhost:8080/login \
-H "Content-Type: application/json" \
-d '{
  "email": "user3@example.com",
  "password": "securepassword",
  "tenant_id": 1
}'
```

- **Description:** Authenticates a user and returns a JWT.
- **Method:** `POST`
- **Endpoint:** `/login`
- **Headers:**
  - `Content-Type: application/json`
- **Request Body (JSON):**
  ```json
  {
    "email": "user3@example.com",
    "password": "securepassword",
    "tenant_id": 1
  }
  ```
- **Response:** `200 OK`
- **Response Example:**
  ```json
  {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InVzZXIzQGV4YW1wbGUuY29tIiwidGVuYW50X2lkIjoxLCJyb2xlIjoiYWRtaW4iLCJleHAiOjE3NDc0Nzc4NzB9.UD96ERyBSyjzHHMK9eUtmyaFSyvlFe1xxAzOjwbfUus"
  }
  ```
  *(Note: Token will differ for each login.)*

---



### 🚪 User Logout

```bash
curl -X POST http://localhost:8080/logout \
-H "Authorization: Bearer <your_jwt_token>" \
-H "Content-Type: application/json"
```

- **Description:** Logs out a user by invalidating their JWT.
- **Method:** `POST`
- **Endpoint:** `/logout`
- **Headers:**
  - `Authorization: Bearer <your_jwt_token>`
  - `Content-Type: application/json`
- **Response:** `200 OK`
- **Response Example:**
  ```json
  {
    "message": "User logged out successfully"
  }
  ```
  *(Returns an appropriate error: `401 Unauthorized` if the JWT is missing, `500 Internal Server Error` for any other issue.)*

### 👤 Get User Info

```bash
curl -X GET http://localhost:8080/api/user/{id} \
-H "Authorization: Bearer <your_jwt_token>" \
-H "Content-Type: application/json"
```

- **Description:** Retrieves detailed information about a specific user by ID. Requires JWT.
- **Method:** `GET`
- **Endpoint:** `/api/user/{id}`
- **Headers:**
  - `Authorization: Bearer <your_jwt_token>`
  - `Content-Type: application/json`
- **Response:** `200 OK`
- **Response Example:**
  ```json
  {
    "team_id": 1,
    "team_name": "Example Team",
    "tenant_id": 1,
    "tenant_name": "TenantName",
    "user_email": "user3@example.com",
    "user_id": 1,
    "user_name": "John Doe"
  }
  ```
  *(Returns an appropriate error: `401 Unauthorized`, `404 Not Found` if the request fails.)*

---

