FROM golang:1.22.5-alpine3.20 AS builder

COPY seclink/ /build/
WORKDIR /build
RUN go env -w GOPROXY='https://nexus.int.keepclm.com/repository/go-proxy' && \
    go env -w 'GOSUMDB=sum.golang.org https://nexus.int.keepclm.com/repository/go-sum'
RUN --mount=type=cache,id=gocache,target=/gocache,rw GOCACHE=/gocache go build -v .

FROM alpine:3.20 AS runner

RUN addgroup seclink --gid 599 && adduser --ingroup seclink --uid 599 --shell /bin/bash --disabled-password seclink

COPY --from=builder --chown=seclink:seclink /build/seclink /seclink/seclink
COPY seclink/seclink.yaml seclink/
RUN chmod +x /seclink/seclink
WORKDIR /seclink

ENTRYPOINT ["/seclink/seclink"]