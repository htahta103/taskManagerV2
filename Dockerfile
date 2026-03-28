# Build: docker build -t taskmanagerv2-api .
# Run:   docker run -e DATABASE_URL=... -e JWT_SECRET=... -e CORS_ORIGIN=... -p 8080:8080 taskmanagerv2-api
FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/api ./cmd/api

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
COPY --from=build /out/api /app/api
ENV PORT=8080
EXPOSE 8080
USER nobody
ENTRYPOINT ["/app/api"]
