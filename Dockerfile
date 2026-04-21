FROM node:20-alpine AS dashboard-builder

WORKDIR /app

RUN corepack enable && corepack prepare pnpm@10.33.0 --activate

COPY dashboard/package.json dashboard/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY dashboard/ .
RUN NODE_OPTIONS=--max-old-space-size=4096 pnpm build

FROM golang:1.24-alpine AS go-builder

WORKDIR /app

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ .

COPY --from=dashboard-builder /app/ui/dist ./internal/handler/admin/ui/dist

ENV CGO_ENABLED=0

ARG VERSION=dev
ARG BUILD_TIME=unknown
ARG VCS_REF=unknown

RUN go build -p 1 \
    -ldflags "-X github.com/Culturae-org/culturae/internal/version.Version=${VERSION} -X github.com/Culturae-org/culturae/internal/version.BuildTime=${BUILD_TIME} -X github.com/Culturae-org/culturae/internal/version.VcsRef=${VCS_REF}" \
    -o culturae ./cmd/main.go

FROM alpine:3.20

ARG VERSION=dev
ARG BUILD_TIME=unknown
ARG VCS_REF=unknown

LABEL org.opencontainers.image.title="Culturae Platform" \
      org.opencontainers.image.description="Culturae quiz platform - API and Admin Dashboard" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.created="${BUILD_TIME}" \
      org.opencontainers.image.source="https://github.com/culturae-org/culturae" \
      org.opencontainers.image.revision="${VCS_REF}" \
      org.opencontainers.image.vendor="Culturae" \
      org.opencontainers.image.licenses="MIT"

RUN apk add --no-cache ca-certificates wget

RUN adduser -D -s /sbin/nologin appuser

WORKDIR /app

COPY --from=go-builder /app/culturae ./culturae
COPY --from=go-builder /app/docs ./docs

USER appuser

EXPOSE 8080

CMD ["./culturae"]
