FROM gcr.io/distroless/static-debian11:nonroot

COPY token2go-server /token2go-server/token2go-server

EXPOSE 8080

ENTRYPOINT ["/token2go-server/token2go-server"]
