FROM golang:1.20-alpine AS build

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 go build -o merchant .

FROM alpine:latest

WORKDIR /app

COPY --from=build /app/merchant /app/

EXPOSE 8080

CMD ["/app/merchant"]
