FROM golang:1.24-alpine AS build

WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN go build -o /out/lingshu ./cmd/server

FROM alpine:3.20 AS runtime

WORKDIR /app
COPY --from=build /out/lingshu /app/lingshu
COPY migrations /app/migrations
EXPOSE 8080
CMD ["/app/lingshu"]
