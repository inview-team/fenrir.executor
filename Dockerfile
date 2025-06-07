FROM golang:1.23 AS builder
ENV PROJECT_PATH=/build
ENV CGO_ENABLED=0
ENV GOOS=linux
COPY . ${PROJECT_PATH}
WORKDIR ${PROJECT_PATH}
RUN go build cmd/server/main.go

FROM golang:alpine
WORKDIR /etc/gorynych
COPY --from=builder /build/main .
EXPOSE 30000
CMD ["./main"]