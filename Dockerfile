FROM scratch

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

COPY dist/*-${TARGETOS}-${TARGETARCH}/token2go-server* /app/token2go-server

EXPOSE 8080

ENTRYPOINT ["/app/token2go-server"]
