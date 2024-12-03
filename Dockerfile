FROM golang:1 AS builder

RUN mkdir /build
ADD . /build/
WORKDIR /build

RUN CGO_ENABLED=1 GOOS=linux go build -a -buildvcs=false -installsuffix cgo -ldflags "-extldflags '-static'" -o main git.tdpain.net/codemicro/palmatum/palmatum

RUN go install github.com/caddyserver/xcaddy/cmd/xcaddy@latest

RUN xcaddy build \
    --output caddy \
    --with git.tdpain.net/codemicro/palmatum/caddyZipFs \
    --replace git.tdpain.net/codemicro/palmatum=/build

FROM alpine
COPY --from=builder /build/main /
COPY --from=builder /build/caddy /
WORKDIR /run

CMD ["../main"]
