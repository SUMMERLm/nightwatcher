FROM 121.40.102.76:30080/ci/golang:alpine AS builder

WORKDIR /build

COPY . .

RUN go env -w GOPROXY=https://goproxy.cn,direct
#RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-w -s" -o _output/bin/nightwatcher -a -installsuffix cgo ./
#RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o _output/bin/gslb -a -installsuffix cgo ./cmd/gslb/
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o _output/bin/cdn -a -installsuffix cgo ./


#FROM 121.40.102.76:30080/ci/ubuntu:18.04 as nightwatcher
#
#WORKDIR /
#COPY --from=builder /build/_output/bin/nightwatcher .
#
#ENTRYPOINT ["./nightwatcher"]

# Build the scheduler binary
#FROM 121.40.102.76:30080/ci/ubuntu:18.04 as gslb
#WORKDIR /
#COPY  --from=builder /build/_output/bin/gslb .
#ENTRYPOINT ["./gslb"]

# Build the scheduler binary
FROM 121.40.102.76:30080/ci/ubuntu:18.04 as cdn
WORKDIR /
COPY  --from=builder /build/_output/bin/cdn .
ENTRYPOINT ["./cdn"]
