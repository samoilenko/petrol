FROM golang:alpine as builder
RUN apk --no-cache add gcc g++ make ca-certificates git
WORKDIR /app
ENV GO111MODULE=on

COPY app/ .

ARG VERSION
ARG USER
ARG DATE
ARG GIT_HASH

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-X 'petrol/build.Version=$VERSION' -X 'petrol/build.User=$USER' -X 'petrol/build.Time=$DATE' -X 'petrol/build.Hash=$GIT_HASH'" -a -installsuffix cgo -o app .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app .

CMD ["./app"]
