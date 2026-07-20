# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/migrate ./cmd/migrate
RUN CGO_ENABLED=0 go build -o /bin/api ./cmd/api
RUN CGO_ENABLED=0 go build -o /bin/worker ./cmd/worker

FROM alpine:3.21 AS migrate
RUN apk add --no-cache ca-certificates
COPY --from=build /bin/migrate /migrate
COPY migrations /migrations
ENTRYPOINT ["/migrate"]

FROM alpine:3.21 AS api
RUN apk add --no-cache ca-certificates
COPY --from=build /bin/api /api
EXPOSE 8080
ENTRYPOINT ["/api"]

FROM alpine:3.21 AS worker
RUN apk add --no-cache ca-certificates
COPY --from=build /bin/worker /worker
ENTRYPOINT ["/worker"]
