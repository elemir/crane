FROM golang:1.10-alpine3.8

RUN apk add --no-cache strace gdb vim bind-tools ethtool iproute2 iptables mariadb-client redis git
RUN go get -u github.com/derekparker/delve/cmd/dlv

