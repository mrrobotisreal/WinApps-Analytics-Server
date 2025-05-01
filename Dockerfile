# ---- Build stage ----
FROM golang:1.24.2-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /app/server ./cmd/server
RUN CGO_ENABLED=0 go build -o /out/ingest   ./cmd/ingest
RUN CGO_ENABLED=0 go build -o /out/consumer ./cmd/consumer

# ---- Final stage ----
FROM gcr.io/distroless/static
WORKDIR /app
COPY --from=build /out/* /app/

CMD ["/app/server","-h"]
#Trying something new, commenting out the rest for now...
#FROM alpine:latest
#RUN apk add --no-cache postgresql redis supervisor
#COPY --from=build /app/server /app/server
#COPY deploy/supervisord.conf /etc/supervisord.conf
#
#CMD ["/usr/bin/supervisord", "-c", "/etc/supervisord.conf"]