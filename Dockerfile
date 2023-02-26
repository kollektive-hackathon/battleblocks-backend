FROM golang:1.20 as build

WORKDIR /go/src/app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /go/bin/app cmd/main.go

FROM golang:1.20-alpine
COPY --from=build /go/bin/app /
CMD ["/app"]
