# GoBank API

## Overview

GoBank API is an educational project to help me learn and improve my coding skills in Go and API development. This RESTful API simulates basic banking operations.

### Features

The GoBank API is capable of performing the following operations:

1. **Account Creation**: Create a new bank account.
2. **Account Retrieval**: Retrieve details of a specific account.
3. **Account Deletion**: Delete an existing account.
4. **Funds Transfer**: Transfer funds between two accounts.
5. **User Authentication**: Authenticate a user and generate a JWT token.

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

### Prerequisites

You need to have the following installed on your system:

- Docker
- Docker Compose

### Installation

1. **Clone the Repository**

   ```bash
   git clone https://github.com/beshoyabdelmalak/gobank.git
   cd gobank
   ```

2. **Set Up Environment Variables**

   Create a .env file in the root directory of the project and update the necessary environment variables:

   ```bash
    # .env file content
    POSTGRES_DB=your_database_name
    POSTGRES_USER=your_database_user
    POSTGRES_PASSWORD=your_database_password
    JWT_SECRET=your_jwt_secret_key
   ```

3. **Build and Run the Application**

   Use Docker Compose to build and run the application:

   ```bash
   docker-compose up --build
   ```

   This command will start the API server and the PostgreSQL database.

### Usage

Once the application is running, you can interact with the API through HTTP requests. The API endpoints include:

- POST /accounts: Create a new account.
- GET /accounts/{id}: Retrieve an account by its ID.
- DELETE /accounts/{id}: Delete an account by its ID.
- POST /login: Authenticate and receive a JWT token.
- POST /transfer: Transfer funds between accounts (requires JWT authentication).

### Testing

Run the automated tests for this system using the following command:

```bash
docker-compose exec gobank-api go test -v ./...
```

### Built With

- [Go](https://golang.org/) - The Go Programming Language
- [Docker](https://www.docker.com/) - Containerization Platform
- [PostgreSQL](https://www.postgresql.org/) - The World's Most Advanced Open Source Relational Database
- [Gorilla Mux](https://github.com/gorilla/mux) - A powerful URL router and dispatcher for Go
- [JWT](https://jwt.io/) - JSON Web Tokens for Go
