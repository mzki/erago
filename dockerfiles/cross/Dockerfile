ARG GOVARIANTS=1.19-bullseye
# no used due to drop latest go version
# FROM dockercore/golang-cross:1.13.15
FROM golang:${GOVARIANTS}

LABEL maintainer mzki
LABEL image-tag  erago-cross

ENV USER gopher
ENV HOME /home/${USER}
ENV SHELL /bin/bash

# install archiver
RUN apt update && \
  apt-get install --no-install-recommends -y zip unzip && \
  rm -rf /var/lib/apt/lists/*

# add user
# It's required when build results are used on local linux machine.
# However the volume mount has root permission and denied writing build cache.
# No problem for releasing executable's. 
#
# RUN echo "Adding ${USER} user and group" \
#  && groupadd --system --gid 1000 ${USER} \
#  && useradd --system --gid ${USER} --uid 1000 --shell ${SHELL} --create-home ${USER} \
#  && chown --recursive ${USER}:${USER} ${HOME}
# RUN echo "${USER}:${USER}"| chpasswd
# RUN usermod -aG sudo gopher

# start user space task...
# USER ${USER}
RUN mkdir -p ${HOME}
WORKDIR ${HOME}

