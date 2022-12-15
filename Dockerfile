FROM scratch

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

COPY dist/token2go-server-${TARGETOS}-${TARGETARCH} /app/token2go-server

EXPOSE 8080

ENTRYPOINT ["/app/token2go-server"]
