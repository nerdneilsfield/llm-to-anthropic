FROM golang:1.23.2-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /out/llm-to-anthropic ./

FROM alpine:3.20

RUN addgroup -S app && adduser -S app -G app
USER app

WORKDIR /app
COPY --from=build /out/llm-to-anthropic /app/llm-to-anthropic

# Copy config file as example
COPY config.toml /app/config.toml
COPY .env.example /app/.env.example

# Expose default port
EXPOSE 8082

ENTRYPOINT ["/app/llm-to-anthropic"]
