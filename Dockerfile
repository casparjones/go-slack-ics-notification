# Starte von einem Golang base image
FROM golang:1.19.3-alpine

# Setze das aktuelle Arbeitsverzeichnis in das Container image
WORKDIR /app

# Kopiere go mod und sum files
COPY go.mod go.sum ./

# Lade die Go dependencies
RUN go mod download

# Kopiere den Quellcode in das Container image
COPY . .

# Baue die Go Anwendung
RUN go build -o main .

# Exponiere Port 8080
EXPOSE 8080

# Führe den kompilierten binären Code aus
CMD ["./main"]
