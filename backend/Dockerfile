FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod ./
COPY go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Update build timestamps in version files
RUN if [ -f VERSION.json ]; then \
      BUILD_DATE=$(TZ='Asia/Bangkok' date '+%Y-%m-%d %H:%M:%S') && \
      BUILD_TIME="$BUILD_DATE" && \
      if command -v jq >/dev/null 2>&1; then \
        jq --arg date "$BUILD_DATE" --arg time "$BUILD_TIME" '.buildDate = $date | .buildTime = $time' VERSION.json > VERSION.json.tmp && mv VERSION.json.tmp VERSION.json; \
      elif command -v python3 >/dev/null 2>&1; then \
        python3 -c "import json, sys; d=json.load(open('VERSION.json')); d['buildDate']='$BUILD_DATE'; d['buildTime']='$BUILD_TIME'; json.dump(d, open('VERSION.json','w'), indent=2); print()"; \
      fi; \
    fi && \
    if [ -f DATABASE_VERSION.json ]; then \
      BUILD_DATE=$(TZ='Asia/Bangkok' date '+%Y-%m-%d %H:%M:%S') && \
      BUILD_TIME="$BUILD_DATE" && \
      if command -v jq >/dev/null 2>&1; then \
        jq --arg date "$BUILD_DATE" --arg time "$BUILD_TIME" '.buildDate = $date | .buildTime = $time' DATABASE_VERSION.json > DATABASE_VERSION.json.tmp && mv DATABASE_VERSION.json.tmp DATABASE_VERSION.json; \
      elif command -v python3 >/dev/null 2>&1; then \
        python3 -c "import json, sys; d=json.load(open('DATABASE_VERSION.json')); d['buildDate']='$BUILD_DATE'; d['buildTime']='$BUILD_TIME'; json.dump(d, open('DATABASE_VERSION.json','w'), indent=2); print()"; \
      fi; \
    fi

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -mod=mod -a -installsuffix cgo -o main ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/main .

# Copy version files
COPY --from=builder /app/VERSION.json .
COPY --from=builder /app/DATABASE_VERSION.json .

# Copy timezone data for Thailand
RUN cp /usr/share/zoneinfo/Asia/Bangkok /etc/localtime

EXPOSE 8080

CMD ["./main"]

