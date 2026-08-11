FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 go build -o /out/build-syncer .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /out/build-syncer /usr/local/bin/build-syncer
ENTRYPOINT ["/usr/local/bin/build-syncer"]
