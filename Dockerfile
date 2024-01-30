FROM golang:latest as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN GOARCH=amd64 GOOS=linux go build cmd/main/main.go

FROM scratch
COPY --from=builder /app/main .
COPY --from=builder /app/mig/000000.sql ./mig/
EXPOSE 80
CMD ["./main"]

