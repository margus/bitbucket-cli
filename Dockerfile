# Goreleaser pre-builds the `bb` binary; this Dockerfile just packages
# it into a minimal scratch image. Run via `goreleaser release`.
FROM alpine:3.21 AS certs
RUN apk add --no-cache ca-certificates

FROM scratch
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY bb /usr/local/bin/bb
ENV HOME=/tmp
WORKDIR /workdir
ENTRYPOINT ["/usr/local/bin/bb"]
CMD ["--help"]
