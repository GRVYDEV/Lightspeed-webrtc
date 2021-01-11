FROM golang:1.14 as builder

WORKDIR /go/src/app
COPY . .
ENV GO111MODULE=on
RUN go mod download
RUN go build -ldflags "-linkmode external -extldflags -static" -o lightspeed-webrtc .


FROM scratch
COPY --from=builder /go/src/app/lightspeed-webrtc /usr/local/bin/
EXPOSE 8080

#CMD ["lightspeed-webrtc --addr=XXX.XXX.XXX.XXX", "run"]
# defaults to localhost:8080, then up to docker compose to bind ports
CMD ["lightspeed-webrtc", "--addr=localhost"]
