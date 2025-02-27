# # Step 1: Builder
# FROM golang:1.20-alpine3.16 as permify-builder
# WORKDIR /go/src/app
# RUN apk update && apk add --no-cache git
# COPY . .
# RUN CGO_ENABLED=0 go build -v ./cmd/permify/


# # Step 2: Final
# FROM cgr.dev/chainguard/static:latest
# COPY --from=ghcr.io/grpc-ecosystem/grpc-health-probe:v0.4.12 /ko-app/grpc-health-probe /usr/local/bin/grpc_health_probe
# COPY --from=permify-builder /go/src/app/permify /usr/local/bin/permify
# ENTRYPOINT ["permify"]
# CMD [""]

FROM golang:1.19-buster as builder

# Create and change to the app directory.
WORKDIR /app

# Retrieve application dependencies.
# This allows the container build to reuse cached dependencies.
# Expecting to copy go.mod and if present go.sum.
COPY go.* ./
RUN go mod download

# Copy local code to the container image.
COPY . ./

# Build the binary.
RUN go build -v -o permify ./cmd/permify/

# Use the official Debian slim image for a lean production container.
# https://hub.docker.com/_/debian
# https://docs.docker.com/develop/develop-images/multistage-build/#use-multi-stage-builds
FROM debian:buster-slim
RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates &&  \
    apt install dumb-init && \
    rm -rf /var/lib/apt/lists/*
    
# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/ /app/

# Run the web service on container startup.
# migrate up --database-engine postgres --database-uri postgres://postgres:postgres@%s/cog-analytics-backend:us-central1:permify/postgres
# or RUN apt install tini
ENTRYPOINT ["/app/permify", "serve", "--database-engine", "postgres", "--database-uri", "user=postgres password=postgres database=postgres host=%s/cog-analytics-backend:us-central1:permify"]
# ENTRYPOINT ["/app/permify", "migrate", "up", "--database-engine", "postgres", "--database-uri", "user=postgres password=postgres database=postgres host=%s/cog-analytics-backend:us-central1:permify"]
# or ENTRYPOINT ["tini", "--"]
# CMD ["./permify serve --database-engine postgres --database-uri postgres://postgres:postgres@%s/cog-analytics-backend:us-central1:permify/postgres"]


# [END run_helloworld_dockerfile]
# [END cloudrun_helloworld_dockerfile]