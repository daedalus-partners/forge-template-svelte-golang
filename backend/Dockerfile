FROM golang:1.25.1-alpine AS build
WORKDIR /src
COPY backend/ ./backend/
WORKDIR /src/backend
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/app ./cmd/app

FROM alpine:3.20
WORKDIR /app
COPY --from=build /out/app ./app
ENV PORT=8080
EXPOSE 8080
ENTRYPOINT ["/app/app"]


