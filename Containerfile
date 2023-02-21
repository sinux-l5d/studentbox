# build go image multistage
# FROM docker.io/library/golang:1.20-alpine3.17 AS builder
FROM --platform=linux/amd64 docker.io/library/golang:1.20-alpine AS builder

# add necessary packages to BUILD against glibc
RUN apk add --no-cache make gcc musl-dev lvm2-dev gpgme-dev btrfs-progs-dev

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
# RUN go build -ldflags '-linkmode external -w -extldflags "-static" ' -o studentbox ./cmd/cli/main.go

# working
RUN CGO_ENABLED=1 go build -o studentbox ./cmd/cli/main.go

# final stage
# FROM docker.io/library/alpine:latest
FROM docker.io/library/alpine:3.17

VOLUME [ "/podman/" ]

RUN apk add --no-cache gcompat gpgme device-mapper

# RUN apk add --no-cache libc6-compat && mkdir -p /var/run/podman/
RUN mkdir -p /podman

# copy the binary from the builder stage
COPY --from=builder /app/studentbox /studentbox

# run the binary
ENTRYPOINT ["./studentbox", "-socket", "unix:/podman/podman.sock"]

CMD ["-list"]
