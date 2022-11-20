FROM alpine:3.12 as install-tool

LABEL maintainer mzki
LABEL image-tag  release-cmd

# install download tools
RUN apk update && \
  apk add --no-cache tar curl
# download release tools
RUN curl -L https://github.com/tcnksm/ghr/releases/download/v0.13.0/ghr_v0.13.0_linux_amd64.tar.gz \
  > ghr_v0.13.0_linux_amd64.tar.gz
RUN tar zxvf ghr_v0.13.0_linux_amd64.tar.gz

# actual running image
FROM alpine:3.12 as final
# copy github release tool 
COPY --from=install-tool /ghr_v0.13.0_linux_amd64/ghr /usr/local/bin/ghr
RUN mkdir -p /home/gopher/erago
WORKDIR /home/gopher/erago
CMD ["ghr", "--version"]