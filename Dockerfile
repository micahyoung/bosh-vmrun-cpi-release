FROM ubuntu:xenial

RUN apt update
RUN apt install -y git curl ruby jq vim
RUN gem install bundler
