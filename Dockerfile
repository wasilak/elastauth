FROM quay.io/wasilak/golang:1.23 AS builder

COPY . /app
WORKDIR /app/
RUN mkdir -p ./dist

RUN CGO_ENABLED=0 go build -o /elastauth

FROM scratch

LABEL org.opencontainers.image.source="https://github.com/wasilak/elastauth"

COPY --from=builder /elastauth .

ENV USER=root

CMD ["/elastauth"]
