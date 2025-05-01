# ---- Build stage ----
FROM golang:1.24.2-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .

FROM builder AS build-server
RUN CGO_ENABLED=0 go build -o /app/server ./cmd/server

FROM builder AS build-ingest
RUN CGO_ENABLED=0 go build -o /app/ingest   ./cmd/ingest

FROM builder AS build-consumer
RUN CGO_ENABLED=0 go build -o /app/consumer ./cmd/consumer

# ---- Final stage ----
FROM alpine:latest AS runtime
RUN apk add --no-cache postgresql redis supervisor ca-certificates
COPY --from=build-server   /app/server   /app/server
COPY --from=build-ingest   /app/ingest   /app/ingest
COPY --from=build-consumer /app/consumer /app/consumer
COPY deploy/supervisord.conf /etc/supervisord.conf

CMD ["/app/server"]

#Trying something new, commenting out the rest for now...
#FROM alpine:latest
#RUN apk add --no-cache postgresql redis supervisor
#COPY --from=build /app/server /app/server
#COPY deploy/supervisord.conf /etc/supervisord.conf
#
#CMD ["/usr/bin/supervisord", "-c", "/etc/supervisord.conf"]