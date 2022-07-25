FROM golang:1.17

WORKDIR /workdir

ENV CGO_ENABLED=0

# fetch dependencies
COPY go.mod go.sum ./
RUN go mod download

COPY . .
