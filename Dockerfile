# Dockerfile used by goreleaser
FROM gcr.io/distroless/static-debian11:nonroot
ARG TARGETPLATFORM
USER 20000:20000
ADD --chmod=555 ${TARGETPLATFORM}/external-dns-domeneshop-webhook /opt/external-dns-domeneshop-webhook/bin/webhook
ENTRYPOINT ["/opt/external-dns-domeneshop-webhook/bin/webhook"]
