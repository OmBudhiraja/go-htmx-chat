FROM golang:1.21-alpine as base
WORKDIR /app
COPY . .
RUN go get -d -v ./...
ENV CGO_ENABLED=0
RUN go build -o ./tmp/main .


FROM golang:1.21-alpine as dev
WORKDIR /app
RUN go install github.com/cosmtrek/air@latest
COPY go.mod go.sum ./
RUN go mod download
CMD ["air", "-c", ".air.toml"]

FROM alpine:latest as prod
COPY --from=base /app /app
WORKDIR /app
EXPOSE 5000
CMD [ "./tmp/main" ]