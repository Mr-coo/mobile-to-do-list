FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /app/api ./cmd/api

FROM alpine
COPY --from=build /app/api /app/api
EXPOSE 8080
CMD ["/app/api"]
