FROM quay.io/wasilak/golang:1.20.3-alpine as builder

ADD . /app
WORKDIR /app/src
RUN mkdir -p ../dist
RUN go build -o ../dist/elastauth

FROM quay.io/wasilak/alpine:3

COPY --from=builder /app/dist/elastauth /elastauth

CMD ["/elastauth"]
