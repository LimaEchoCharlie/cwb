FROM golang:1.12.3-stretch

RUN apt-get update --yes && \
    apt-get install --yes \
    curl \
    build-essential \
    libtool \
    autoconf \
    automake \
    vim && \
    apt-get clean --yes

# add ZeroMQ library and Go wrapper
ADD https://github.com/zeromq/libzmq/releases/download/v4.2.5/zeromq-4.2.5.tar.gz .
RUN tar xf zeromq-4.2.5.tar.gz && \
    cd zeromq-4.2.5 && \
    ./autogen.sh && \
    ./configure --without-docs --enable-drafts=yes && \
    make install && \
    ldconfig /usr/local/lib && \
    go get github.com/pebbe/zmq4

RUN go get github.com/flynn/noise

# add examples
ADD src /go/src

# set bash as the default command in the new container
CMD ["bash"]
