FROM golang:1.12.3-stretch

RUN apt-get update --yes && \
    apt-get install --yes \
    libssl-dev \
    vim && \
    apt-get clean --yes

RUN go get gopkg.in/square/go-jose.v2
RUN go get github.com/google/go-tpm/tpm2
RUN go get github.com/google/go-tpm-tools/simulator

# set bash as the default command in the new container
CMD ["bash"]
