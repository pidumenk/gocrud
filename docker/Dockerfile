# Multi-stage build
FROM golang:alpine as builder

# Set working dir to copy files in there
WORKDIR /app

#  Download necessary dependencies
COPY ../go.mod ../go.sum ./
RUN go mod download

# Copy the source code into working dir
COPY ../ ./

# Build Go application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /gocrud

#------------------------------------------------------------#

# Use extra-minimal image to optimize the final size (https://hub.docker.com/_/scratch)
FROM scratch

# Copy data from the first stage
COPY --from=builder /gocrud /

# Specify the port to be exposed
EXPOSE 8080

# Set entrypoint
ENTRYPOINT ["/gocrud"]