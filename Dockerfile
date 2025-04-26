# ---- Build stage ----
FROM golang:1.24.2-alpine AS build
WORKDIR /src
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /app/server ./cmd/server

# ---- Final stage ----
FROM alpine:latest
RUN apk add --no-cache postgresql redis supervisor
COPY --from=build /app/server /app/server
COPY deploy/supervisord.conf /etc/supervisord.conf
CMD ["/usr/bin/supervisord", "-c", "/etc/supervisord.conf"]