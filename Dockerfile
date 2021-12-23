FROM golang AS builder

WORKDIR /build
COPY . /build
RUN go get ./...
RUN go build /build/cmd/main.go

FROM golang:alpine
RUN apk add libc6-compat
COPY --from=builder /build/main /opt/main
ENTRYPOINT [ "/opt/main" ]
