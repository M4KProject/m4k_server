FROM caddy:2-builder AS builder

ENV GOPROXY=https://goproxy.io,direct

# First build caddy without plugins for download deps
RUN xcaddy build

# Build with plugins
RUN xcaddy build \
    --with github.com/caddyserver/replace-response \
    --with github.com/lindenlab/caddy-s3-proxy \
    --with github.com/ueffel/caddy-imagefilter/v2/defaults

FROM caddy:2

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
