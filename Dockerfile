# syntax=docker/dockerfile:1
FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/poros ./cmd/poros

FROM alpine:3.20
RUN apk --no-cache add ca-certificates
COPY --from=build /out/poros /usr/local/bin/poros
EXPOSE 8080
ENTRYPOINT ["poros"]
CMD ["serve", "--addr", ":8080", "--data", "/data", "--db", "postgres://poros:poros@db:5432/poros?sslmode=disable"]
