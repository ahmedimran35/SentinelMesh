# Stage 1: Build frontend
FROM node:20-alpine AS frontend
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# Stage 2: Build Go binary
FROM golang:1.22-alpine AS backend
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/frontend/dist ./frontend/dist
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o sentinelmesh .

# Stage 3: Final minimal image
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=backend /app/sentinelmesh .

ENV PORT=8090
ENV HOST=0.0.0.0
ENV DB_PATH=/data/sentinelmesh.db
ENV API_RATE_LIMIT=60
ENV MAX_CONCURRENT_SCANS=5

EXPOSE 8090
VOLUME ["/data"]

ENTRYPOINT ["./sentinelmesh"]
