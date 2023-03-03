ARG PLATFORM_OS=linux
ARG PLATFORM_ARCH=amd64
ARG PLATFORM=${PLATFORM_OS}/${PLATFORM_ARCH}
FROM --platform=${PLATFORM} docker.io/library/golang:1.20-alpine AS builder

RUN apk add --no-cache upx make

WORKDIR /app

COPY go.mod go.sum Makefile ./

RUN make download

COPY . .

ARG VERSION=dev
ARG BINARY_NAME=studentbox

RUN make build && \
    upx --lzma bin/studentbox && \
    # to copy in final image
    mkdir /podman


FROM scratch

VOLUME [ "/podman/" ]

# mkdir don't work in scratch container
COPY --from=builder /podman /

COPY --from=builder /app/bin/studentbox /studentbox

ENTRYPOINT ["./studentbox", "--socket", "unix:/podman/podman.sock"]

CMD ["list"]
