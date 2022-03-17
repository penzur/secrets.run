FROM golang:latest AS builder
ADD . /secrets
WORKDIR /secrets
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o secrets .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder secrets/secrets ./
COPY --from=builder secrets/tmpl ./tmpl
RUN chmod +x ./secrets
ENTRYPOINT ["./secrets"]
