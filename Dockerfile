# Multi-stage: compila estático y empaca solo los binarios (~20 MB final).
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/api ./cmd/api \
 && CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/seed ./cmd/seed

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=build /out/ /app/
EXPOSE 4051
CMD ["/app/api"]
