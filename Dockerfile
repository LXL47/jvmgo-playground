FROM golang:1.26.5-bookworm AS app-builder
ENV GOPROXY=https://goproxy.cn,direct \
    GOSUMDB=sum.golang.google.cn
WORKDIR /src/apps
COPY apps/go.mod apps/go.sum ./
RUN go mod download
COPY apps/ ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api \
    && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/runner ./cmd/runner \
    && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/healthcheck ./cmd/healthcheck

FROM golang:1.26.5-bookworm AS jvm-builder
ENV GOPROXY=https://goproxy.cn,direct \
    GOSUMDB=sum.golang.google.cn
WORKDIR /src/jvm
COPY jvm/runtime/go.mod ./
COPY jvm/runtime/ ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/jvmgo ./ch11

FROM gcr.io/distroless/static-debian12:nonroot AS api
WORKDIR /app
COPY --from=app-builder /out/api /app/api
COPY --from=app-builder /out/healthcheck /app/healthcheck
EXPOSE 8080
ENTRYPOINT ["/app/api"]
CMD ["-config", "/app/config/api.yaml"]

FROM eclipse-temurin:8-jdk-jammy AS runner
WORKDIR /app
COPY --from=app-builder --chmod=0555 /out/runner /app/runner
COPY --from=app-builder --chmod=0555 /out/healthcheck /app/healthcheck
COPY --from=jvm-builder --chmod=0555 /out/jvmgo /app/jvmgo
EXPOSE 8081
ENTRYPOINT ["/app/runner"]
CMD ["-config", "/app/config/runner.yaml"]
