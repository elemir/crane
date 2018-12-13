FROM golang:1.10-alpine3.8

RUN echo @testing http://nl.alpinelinux.org/alpine/edge/testing >> /etc/apk/repositories
RUN apk add --no-cache strace gdb vim bind-tools ethtool iproute2 iptables mariadb-client redis git percona-toolkit@testing
RUN go get -u github.com/derekparker/delve/cmd/dlv

