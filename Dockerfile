ARG BASE_IMAGE=indiefan/nanit

FROM --platform=$BUILDPLATFORM golang:1.24.0 AS build
ADD cmd /app/cmd
ADD pkg /app/pkg
ADD go.mod /app/
ADD go.sum /app/
ADD scripts /app/scripts
WORKDIR /app
ARG CI_COMMIT_SHORT_SHA
ARG TARGETOS TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags "-X main.GitCommit=$CI_COMMIT_SHORT_SHA" -o ./bin/nanit ./cmd/nanit/*.go


FROM $BASE_IMAGE
COPY --from=build /app/bin/nanit /app/bin/nanit
COPY --from=build /app/scripts /app/scripts

RUN apt-get update && apt-get install -y \
    bash \
    curl \
    jq

RUN mkdir -p /data && \
    chmod +x /app/scripts/*.sh

WORKDIR /app
ENTRYPOINT ["/app/bin/nanit"]