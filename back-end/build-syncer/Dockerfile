FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY main.go ./
RUN CGO_ENABLED=0 go build -o /out/build-syncer .

FROM alpine:3.20
COPY --from=build /out/build-syncer /usr/local/bin/build-syncer
ENTRYPOINT ["/usr/local/bin/build-syncer"]
