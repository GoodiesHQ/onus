# -------------------------------
# 1) Build webapp (SvelteKit SPA)
# -------------------------------
FROM node:22-alpine AS build-webapp
WORKDIR /app/webapp

# Enable pnpm via corepack
RUN corepack enable

# Copy manifests for caching
COPY webapp/package.json webapp/pnpm-lock.yaml webapp/.npmrc* ./
RUN pnpm install --frozen-lockfile

# Copy the rest and build
COPY webapp/ ./
RUN pnpm build

# ---------------------
# 2) Build backend (Go)
# ---------------------
FROM golang:1.25-alpine AS build-backend
WORKDIR /app

# Install git and update CA certificates
RUN apk add --no-cache ca-certificates git

# Cache Go deps
COPY go.mod go.sum ./
RUN go mod download

# Copy Go source 
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY sql/ ./sql/
COPY fs.go ./

ENV CGO_ENABLED=0
RUN go build -trimpath -ldflags="-s -w" -o /out/onus ./cmd

# ----------------------
# 3) Final runtime image
# ----------------------
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app

COPY --from=build-backend /out/onus /app/onus
COPY --from=build-webapp /app/webapp/build /app/static

ENV ONUS_SERVER_STATIC_DIR=/app/static

ENTRYPOINT ["/app/onus"]
CMD ["serve"]