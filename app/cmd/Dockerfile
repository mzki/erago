FROM dockercore/golang-cross:1.12.5

LABEL maintainer mzki
LABEL image-tag  golang-cross-local-user

ENV USER gopher
ENV HOME /home/${USER}
ENV SHELL /bin/bash

# add user 
RUN echo "Adding ${USER} user and group" \
  && groupadd --system --gid 1000 ${USER} \
  && useradd --system --gid ${USER} --uid 1000 --shell ${SHELL} --create-home ${USER} \
  && chown --recursive ${USER}:${USER} ${HOME}
RUN echo "${USER}:${USER}"| chpasswd
RUN usermod -aG sudo gopher

# start user space task...
USER ${USER}
WORKDIR ${HOME}

