# syntax = docker/dockerfile:experimental

#
# BUILD STAGE - Stage for building the app
#

FROM golang:1.15.2 as build_app

RUN mkdir -p /app/code

RUN mkdir -p -m 0600 ~/.ssh && ssh-keyscan github.com >> ~/.ssh/known_hosts
RUN git config --global url."git@github.com:".insteadOf "https://github.com/"

COPY Makefile go.mod go.sum /app/code/
WORKDIR /app/code
RUN make deps

COPY . /app/code
RUN make build


#
# BASE STAGE - Stage for building base runtime
#
FROM ubuntu:20.04 as base

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN apt update -yq && \
  DEBIAN_FRONTEND=noninteractive apt install --no-install-recommends -yq \
    make ca-certificates && \
  apt autoremove -y --purge  && \
  apt clean -y  && \
  apt autoclean -y && \
  rm -rf /var/lib/apt/lists/* && \
  useradd -r -u 999 release_dash && \
  mkdir -p /app/bin && \
  chown -R release_dash:release_dash /app/bin


#
# RUNTIME STAGE - Stage for running the service
#
FROM base as run

USER release_dash

COPY --chown=release_dash:release_dash --from=build_app /go/bin/release-dash /app/bin
EXPOSE 8080

CMD /app/bin/release-dash
