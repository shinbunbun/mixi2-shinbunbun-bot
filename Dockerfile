FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bot ./cmd/bot

FROM gcr.io/distroless/static-debian12:nonroot
LABEL org.opencontainers.image.source=https://github.com/shinbunbun/mixi2-shinbunbun-bot
COPY --from=builder /bot /bot
ENTRYPOINT ["/bot"]
