FROM --platform=linux/amd64 docker.io/library/golang:1.20-alpine AS builder

RUN apk add --no-cache upx 

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

# tags come from: https://github.com/containers/podman/issues/12548#issuecomment-989053364
RUN CGO_ENABLED=0 go build -tags "remote exclude_graphdriver_btrfs btrfs_noversion exclude_graphdriver_devicemapper containers_image_openpgp" -o studentbox ./cmd/studentbox/main.go && \
    upx --lzma studentbox && \
    # to copy in final image
    mkdir /podman


FROM scratch

VOLUME [ "/podman/" ]

# mkdir don't work in scratch container
COPY --from=builder /podman /

COPY --from=builder /app/studentbox /studentbox

ENTRYPOINT ["./studentbox", "--socket", "unix:/podman/podman.sock"]

CMD ["list"]
