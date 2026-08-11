FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 go build -o /out/segton .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /out/segton /usr/local/bin/segton
ENTRYPOINT ["/usr/local/bin/segton"]
