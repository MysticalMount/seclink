FROM golang:1.22.1-alpine3.19 AS builder

COPY seclink/ /build/
WORKDIR /build
RUN go build -v .

FROM alpine:3.19 AS runner

RUN addgroup seclink --gid 599 && adduser --ingroup seclink --uid 599 --shell /bin/bash --disabled-password seclink

COPY --from=builder --chown=seclink:seclink /build/seclink /seclink/seclink
COPY seclink/seclink.yaml seclink/
RUN chmod +x /seclink/seclink
WORKDIR /seclink

ENTRYPOINT ["/seclink/seclink"]