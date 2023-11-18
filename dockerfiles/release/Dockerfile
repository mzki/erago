ARG GOVARIANTS=1.19-bullseye
FROM alpine:3.12 as install-tool

# install download tools
RUN apk update && \
  apk add --no-cache tar curl
# download release tools
RUN curl -L https://github.com/tcnksm/ghr/releases/download/v0.13.0/ghr_v0.13.0_linux_amd64.tar.gz \
  > ghr_v0.13.0_linux_amd64.tar.gz
RUN tar zxvf ghr_v0.13.0_linux_amd64.tar.gz

#==================================================
ARG GOVARIANTS
FROM golang:${GOVARIANTS} AS install-go-tool

# build as static binary to run it on alpine image.
RUN CGO_ENABLED=0 go install -ldflags="-s -w -extldflags \"-static\"" github.com/google/go-licenses@v1.5.0

#==================================================
# actual running image.
ARG GOVARIANTS
FROM golang:${GOVARIANTS} as final

LABEL maintainer mzki
LABEL image-tag  release-cmd

# copy github release tool 
COPY --from=install-tool /ghr_v0.13.0_linux_amd64/ghr /usr/local/bin/ghr
COPY --from=install-go-tool /go/bin/go-licenses /usr/local/bin/go-licenses
RUN mkdir -p /home/gopher/erago
WORKDIR /home/gopher/erago
CMD ["ghr", "--version"]