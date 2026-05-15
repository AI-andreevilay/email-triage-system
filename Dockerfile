FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/api-server ./cmd/api-server
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/classifier-worker ./cmd/classifier-worker
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/label-worker ./cmd/label-worker
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/migrator ./cmd/migrator

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=builder /out/api-server /app/api-server
COPY --from=builder /out/classifier-worker /app/classifier-worker
COPY --from=builder /out/label-worker /app/label-worker
COPY --from=builder /out/migrator /app/migrator
COPY --from=builder /src/migrations /app/migrations
