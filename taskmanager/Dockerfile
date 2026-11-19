# syntax=docker/dockerfile:1

# Match the `go` line in your go.mod.
ARG GO_VERSION=1.25

FROM golang:${GO_VERSION}-alpine AS build
WORKDIR /src

# Dependencies first: this layer stays cached until go.mod/go.sum change,
# so rebuilds after code edits are fast.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# CGO_ENABLED=0 gives a fully static binary that runs on a bare image.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

# Runtime: no shell, no package manager, runs as non-root by default.
FROM gcr.io/distroless/static:nonroot
COPY --from=build /out/api /api
USER nonroot:nonroot
EXPOSE 8080
# Exec form, so the process receives SIGTERM directly and our graceful shutdown works.
ENTRYPOINT ["/api"]