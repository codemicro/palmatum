FROM golang:1 as builder

RUN mkdir /build
ADD . /build/
WORKDIR /build

RUN CGO_ENABLED=1 GOOS=linux go build -a -buildvcs=false -installsuffix cgo -ldflags "-extldflags '-static'" -o main github.com/codemicro/palmatum/palmatum

RUN go install github.com/caddyserver/xcaddy/cmd/xcaddy@latest

RUN xcaddy build \
    --output caddy \
    --with github.com/codemicro/palmatum/caddyZipFs \
    --replace github.com/codemicro/palmatum=/build

FROM alpine
COPY --from=builder /build/main /
COPY --from=builder /build/caddy /
WORKDIR /run

CMD ["../main"]
