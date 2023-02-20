FROM golang:1.20.1-alpine3.16 as go

RUN mkdir /app

WORKDIR /app
COPY . /app

RUN CGO_ENABLED=0 go build -o main .

FROM scratch

WORKDIR /app
COPY --from=go /etc/passwd /etc/passwd
COPY --from=go /app /app

USER nobody

ENTRYPOINT ["/app/main"]
