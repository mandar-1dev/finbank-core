# --- build stage -----------------------------------------------------
FROM golang:1.22-alpine AS build
WORKDIR /src

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/corebank-server ./cmd/server

# --- runtime stage -----------------------------------------------------
FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /out/corebank-server ./corebank-server

EXPOSE 8080
ENTRYPOINT ["./corebank-server"]
