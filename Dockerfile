ARG BASE_IMAGE=registry.gitlab.com/adam.stanek/nanit/base:latest

FROM golang:1.15.5-buster AS build
ADD cmd /app/cmd
ADD pkg /app/pkg
ADD go.mod /app/
ADD go.sum /app/
WORKDIR /app
ARG CI_COMMIT_SHORT_SHA
RUN go build -ldflags "-X main.GitCommit=$CI_COMMIT_SHORT_SHA" -o ./bin/nanit ./cmd/nanit/*.go

FROM $BASE_IMAGE
RUN mkdir -p /app/data
COPY --from=build /app/bin/nanit /app/bin/nanit
WORKDIR /app
CMD ["/app/bin/nanit"]