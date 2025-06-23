FROM dockerhub.timeweb.cloud/golang:1.24-alpine3.21 AS builder
RUN apk update && apk add git make
WORKDIR /build
COPY go.mod go.sum .
COPY . .
RUN make build

FROM dockerhub.timeweb.cloud/alpine:3.21 AS stager
RUN apk update && apk add --no-cache upx
WORKDIR /root/
COPY --from=builder /build/bin/server /bin/server
RUN upx /bin/server

FROM scratch
COPY --from=stager /bin/server /bin/server
COPY --from=builder /build/configs/ /configs/
EXPOSE 8008
ENTRYPOINT ["/bin/server"]

