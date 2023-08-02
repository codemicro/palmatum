FROM golang:1 as builder

RUN mkdir /build
ADD . /build/
WORKDIR /build

RUN CGO_ENABLED=0 GOOS=linux go build -a -buildvcs=false -installsuffix cgo -ldflags "-extldflags '-static'" -o main github.com/codemicro/palmatum/palmatum

FROM alpine
COPY --from=builder /build/main /
WORKDIR /run

CMD ["../main"]
