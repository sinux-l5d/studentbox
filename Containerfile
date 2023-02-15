# build go image multistage
# FROM docker.io/library/golang:1.20-alpine3.17 AS builder
FROM --platform=linux/amd64 docker.io/library/golang:1.20-alpine AS builder

# add necessary packages to BUILD against glibc
RUN apk add --no-cache gcompat libstdc++ make gcc musl-dev linux-headers device-mapper lvm2-dev gpgme-dev btrfs-progs-dev

# RUN apt-get update && apt-get install -y golang-github-proglottis-gpgme-dev golang-github-containerd-btrfs-dev libdevmapper-dev
# RUN apt-get update &-a & \
#     apt-get install -y golang-github-proglottis-gpgme-dev golang-github-containerd-btrfs-dev libdevmapper1.02.1 libdevmapper-event1.02.1 libdevmapper-dev musl-dev && \
#     wget https://www.musl-libc.org/releases/musl-latest.tar.gz && \
#     tar -xzf musl-latest.tar.gz && \
#     cd musl-* && \
#     ./configure --enable-static --disable-shared && \
#     make && make install && \
#     ln -s /usr/local/musl/bin/musl-gcc /usr/local/bin/musl-gcc
# RUN apk update && apk add --update --no-cache gpgme gcc musl-dev gcompat make && \
#     wget https://www.musl-libc.org/releases/musl-latest.tar.gz && \
#     tar -xzf musl-latest.tar.gz && \
#     cd musl-* && \
#     ./configure --enable-static --disable-shared && \
#     make && make install && \
#     ln -s /usr/local/musl/bin/musl-gcc /usr/local/bin/musl-gcc

# set working directory
WORKDIR /app

# copy go.mod and go.sum
COPY go.mod go.sum ./

# download dependencies
RUN go mod download

# copy source code
COPY . .

# build the binary
# RUN GOOS=linux GOARCH=amd64 go build -o studentbox ./cmd/cli/main.go && upx studentbox
# RUN GOOS=linux GOARCH=amd64 go build -ldflags '-linkmode external -w -extldflags "-static" ' -o studentbox ./cmd/cli/main.go
RUN CGO_ENABLED=1 go build -o studentbox ./cmd/cli/main.go

# final stage
# FROM docker.io/library/alpine:latest
FROM docker.io/library/alpine:3.17

VOLUME [ "/podman/" ]

# RUN apk add --no-cache libc6-compat && mkdir -p /var/run/podman/
RUN mkdir -p /podman

# copy the binary from the builder stage
COPY --from=builder /app/studentbox /studentbox

# run the binary
ENTRYPOINT ["./studentbox", "-socket", "/podman/podman.sock"]
