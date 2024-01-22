FROM golang:1.21.5-alpine

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

# Build the application
RUN go build -o gobank .

EXPOSE 8000

# Command to run the executable
CMD ["./gobank"]
