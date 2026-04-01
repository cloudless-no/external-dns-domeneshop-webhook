# Dockerfile used for local test builds.
FROM --platform=${TARGETARCH} gcr.io/distroless/static-debian11:nonroot
ARG TARGETARCH
LABEL org.opencontainers.image.description="TEST Domeneshop DNS webhook for external-dns"
USER 20000:20000
ADD --chmod=555 build/bin/external-dns-domeneshop-webhook-${TARGETARCH} /opt/external-dns-domeneshop-webhook/bin/webhook
ENTRYPOINT ["/opt/external-dns-domeneshop-webhook/bin/webhook"]
