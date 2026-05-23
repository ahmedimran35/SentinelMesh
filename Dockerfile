# Stage 1: Build frontend
FROM node:20-alpine AS frontend
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

# Stage 2: Build backend
FROM golang:1.22-alpine AS backend
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/frontend/dist ./frontend/dist
RUN CGO_ENABLED=1 go build -o sentinelmesh .

# Stage 3: Runtime
FROM alpine:3.19
RUN apk add --no-cache ca-certificates sqlite-libs
COPY --from=backend /app/sentinelmesh /usr/local/bin/sentinelmesh
EXPOSE 8090
ENTRYPOINT ["sentinelmesh"]
