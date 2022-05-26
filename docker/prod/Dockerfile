FROM golang:alpine as builder
RUN apk --no-cache add gcc g++ make ca-certificates git
WORKDIR /app
ENV GO111MODULE=on

COPY app/ .

ARG VERSION
ARG USER
ARG DATE
ARG GIT_HASH

# RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-X 'edupay-go/build.Version=$VERSION' -X 'edupay-go/build.User=$USER' -X 'edupay-go/build.Time=$DATE'" -a -installsuffix cgo -o app .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app .
EXPOSE 8080
EXPOSE 9090
CMD ["./app"]
