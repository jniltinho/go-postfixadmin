# Go-PostfixAdmin API Documentation

## Overview

The modern API is available under `/api/v1`.

All protected endpoints require a valid JWT access token sent as:
```
Authorization: Bearer <access_token>
```

## Authentication Flow

1. **Login**
   - `POST /api/v1/auth/login`
   - Body: `{ "username": "...", "password": "..." }`
   - Returns access token + sets `refresh_token` as httpOnly cookie

2. **Refresh Token**
   - `POST /api/v1/auth/refresh`
   - Uses the httpOnly refresh cookie automatically

3. **Current User**
   - `GET /api/v1/auth/me`

## OpenAPI Specification

The OpenAPI 3.0 specification is available at:

```
GET /api/v1/swagger.json
```

You can view it using:
- [Swagger Editor](https://editor.swagger.io/)
- [Redoc](https://github.com/Redocly/redoc)
- Any OpenAPI compatible tool

## Main Resource Endpoints

| Resource       | Base Path                    | Methods          | Notes |
|----------------|------------------------------|------------------|-------|
| Auth           | /auth                        | login, refresh, me | Public login |
| Domains        | /domains                     | CRUD             | Scoped |
| Mailboxes      | /mailboxes                   | CRUD             | Scoped |
| Aliases        | /aliases                     | CRUD             | Scoped |
| Alias Domains  | /alias-domains               | CRUD             | Scoped |
| Admins         | /admins                      | CRUD             | Superadmin mostly |
| Transports     | /transports                  | CRUD             | Superadmin only |
| Logs           | /logs                        | GET (paginated)  | Scoped |
| Maillog        | /maillog                     | GET (paginated)  | Scoped |

## Running the API

```bash
./postfixadmin server
```

Swagger spec: `http://localhost:8080/api/v1/swagger.json`
