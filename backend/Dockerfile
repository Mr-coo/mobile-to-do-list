FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /app/api ./cmd/api

FROM alpine:3.24
RUN addgroup -S app && adduser -S -G app app
COPY --from=build /app/api /app/api
USER app
EXPOSE 8080
CMD ["/app/api"]
