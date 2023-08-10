FROM --platform=$BUILDPLATFORM golang:1.20 AS builder

LABEL org.opencontainers.image.source=https://github.com/canonical/identity-platform-admin-ui

ARG SKAFFOLD_GO_GCFLAGS
ARG TARGETOS
ARG TARGETARCH
ARG app_name=reader

ENV GOOS=$TARGETOS
ENV GOARCH=$TARGETARCH
ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GO_BIN=/go/bin/app
ENV APP_NAME=$app_name

WORKDIR /var/app

COPY . .

RUN make build

FROM gcr.io/distroless/static:nonroot

LABEL org.opencontainers.image.source=https://github.com/canonical/identity-platform-admin-ui

COPY --from=builder /go/bin/app /app

CMD ["/app"]
