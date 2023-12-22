FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.20-alpine as build

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

COPY --from=watchdog /fwatchdog /usr/bin/fwatchdog
RUN chmod +x /usr/bin/fwatchdog

ENV CGO_ENABLED=0

RUN mkdir -p /app
WORKDIR /app
COPY . .


ARG GO111MODULE="on"
ARG GOPROXY="goproxy.cn"
ARG GOFLAGS=""
ARG DEBUG=0

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go get

RUN CGO_ENABLED=${CGO_ENABLED} GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build --ldflags "-s -w" -a -installsuffix cgo -o scaler .

FROM --platform=${TARGETPLATFORM:-linux/amd64} alpine:3.17
# Add non root user and certs
RUN addgroup -S app && adduser -S -g app app
# Split instructions so that buildkit can run & cache
# the previous command ahead of time.
RUN mkdir -p /home/app \
    && chown app /home/app

WORKDIR /home/app

COPY --from=build --chown=app /app/scaler   .


USER app

CMD ["./scaler"]
