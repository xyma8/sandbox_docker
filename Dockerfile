FROM golang:1.24.9 AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -x -v -o /out/myapp main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /out/myapp ./myapp
COPY ui/ ./ui/
ENTRYPOINT ["./myapp"]
