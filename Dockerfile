FROM golang:1.12.5
COPY . /secrets
WORKDIR /secrets
RUN CC=$(which musl-gcc) go build --ldflags '-w -linkmode external -extldflags "-static"' secrets.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /secrets
COPY --from=0 /secrets /secrets
CMD ["./secrets"]

