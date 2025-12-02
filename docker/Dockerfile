ARG ARCH=amd64
FROM golang:1.23.4 AS builder

ARG ARCH

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod tidy && go mod download
COPY . .

RUN set -ex \
  && export CGO_ENABLED=0 \
  && if [ "${ARCH}" = "amd64" ]; then export GOARCH=amd64; fi \
  && if [ "${ARCH}" = "arm32v7" ]; then export GOARCH=arm; fi \
  && if [ "${ARCH}" = "arm64v8" ]; then export GOARCH=arm64; fi \
  && echo "Compiling for $GOARCH" \
  && go install -v . ./cmd/...

### second stage

FROM $ARCH/alpine:3.16

COPY --from=builder /go/bin/go-flokicoin /bin/

RUN mv /bin/go-flokicoin /bin/lokid

EXPOSE 15212

CMD ["/bin/lokid"]