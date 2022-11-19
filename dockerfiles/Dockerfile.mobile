FROM golang:1.17.1-buster

LABEL maintainer mzki
LABEL image-tag  gomobile-bind

ENV HOME /home/gopher
ENV SHELL /bin/bash

# install dependency tools
RUN apt-get update \
  && apt-get install --no-install-recommends -y unzip openjdk-11-jdk \
  && apt-get -y clean \
  && rm -rf /var/lib/apt/lists/*

# start user space task...
#USER ${USER}
RUN mkdir -p ${HOME}
WORKDIR ${HOME}

# Android sections reference:
# https://github.com/OpenPriv/android-go-mobile/blob/662f10bde613d9f6794792752438d80cfe8742d1/Dockerfile
# https://medium.com/@elye.project/intro-to-docker-building-android-app-cb7fb1b97602
#
# This Dockerfile differs from the references
# - Use Go based docker image.
# - Use more latest sdk versions.

# ANDROID_VERSION=31 --> Android 12.
ENV SDK_URL="https://dl.google.com/android/repository/commandlinetools-linux-8512546_latest.zip" \
    ANDROID_HOME="/usr/local/android-sdk" \
    ANDROID_VERSION=31 \
    ANDROID_BUILD_TOOLS_VERSION=31.0.0
ENV ANDROID_TOOLS_HOME=$ANDROID_HOME/cmdline-tools \
    ANDROID_SDK=$ANDROID_HOME

## Download Android SDK
# Extracted SDK path is modified by:
# https://www.google.com/url?sa=t&rct=j&q=&esrc=s&source=web&cd=&cad=rja&uact=8&ved=2ahUKEwiG1J2A3bD7AhXsl1YBHc8aArYQFnoECA0QAQ&url=https%3A%2F%2Fstackoverflow.com%2Fquestions%2F65262340%2Fcmdline-tools-could-not-determine-sdk-root&usg=AOvVaw0Oqw-glr4hAU7DIQiYe66q
RUN mkdir -p "$ANDROID_TOOLS_HOME" .android \
    && touch .android/repositories.cfg \
    && cd "$ANDROID_HOME" \
    && curl -o sdk.zip $SDK_URL \
    && unzip -d ${ANDROID_TOOLS_HOME} sdk.zip \
    && mv ${ANDROID_TOOLS_HOME}/cmdline-tools ${ANDROID_TOOLS_HOME}/tools \
    && rm sdk.zip \
    && yes | $ANDROID_TOOLS_HOME/tools/bin/sdkmanager --licenses \
    && cd -

## Install Android Build Tool and Libraries
RUN $ANDROID_TOOLS_HOME/tools/bin/sdkmanager --update
RUN $ANDROID_TOOLS_HOME/tools/bin/sdkmanager "build-tools;${ANDROID_BUILD_TOOLS_VERSION}" \
    "platforms;android-${ANDROID_VERSION}" \
    "platform-tools"

# Install NDK
ENV NDK_VER="25.1.8937393"
RUN $ANDROID_TOOLS_HOME/tools/bin/sdkmanager "ndk;$NDK_VER"
# gomobile adapt new ndk path, so this is no longer needed.
# RUN ln -sf $ANDROID_HOME/ndk/$NDK_VER $ANDROID_HOME/ndk-bundle

# Get gomobile and initialize with android NDK
RUN git clone https://github.com/golang/mobile.git mobile \
  && cd mobile \
  && git checkout 43a0384520996c8376bfb8637390f12b44773e65 \
  && cd cmd/gomobile \
  && go install . \
  && cd - \
  && gomobile init 

CMD ["gomobile", "version"]