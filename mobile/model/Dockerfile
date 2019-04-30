FROM golang:1.12.1

LABEL maintainer mzki
LABEL image-tag  gomobile-bind

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

# install dependency tools
RUN apt-get update \
  && apt-get install --no-install-recommends -y unzip openjdk-8-jdk \
  && rm -rf /var/lib/apt/lists/*

# start user space task...
USER ${USER}
WORKDIR ${HOME}

# Get android sdk, ndk, and rest of the stuff needed to build the android app.
ENV ANDROID_HOME $HOME/android-sdk
ENV ANDROID_SDK_TOOL_DL sdk-tools-linux-4333796.zip

RUN mkdir -p $ANDROID_HOME \
  && curl -O "https://dl.google.com/android/repository/${ANDROID_SDK_TOOL_DL}" \
  && unzip -d $ANDROID_HOME $ANDROID_SDK_TOOL_DL \
  && rm -rf $ANDROID_SDK_TOOL_DL

# agree sdkmanager licenses
RUN mkdir ~/.android \
  && touch ~/.android/repositories.cfg
RUN yes | $ANDROID_HOME/tools/bin/sdkmanager --licenses
# install sdk packages
RUN $ANDROID_HOME/tools/bin/sdkmanager \
  'build-tools;28.0.3' \
  'platforms;android-28' \
  'extras;android;m2repository' \
  'ndk-bundle'

# Get gomobile
RUN go get -u golang.org/x/mobile/cmd/gomobile \
  && gomobile init 
