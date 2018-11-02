FROM centos:7
LABEL maintainer "Devtools <devtools@redhat.com>"
LABEL author "Devtools <devtools@redhat.com>"
ENV LANG=en_US.utf8
ARG USE_GO_VERSION_FROM_WEBSITE=0

# Some packages might seem weird but they are required by the RVM installer.
RUN yum --enablerepo=centosplus install -y --quiet \
      findutils \
      git \
      $(test -z $USE_GO_VERSION_FROM_WEBSITE && echo "golang") \
      make \
      procps-ng \
      tar \
      wget \
      which \
      bc \
      postgresql \
    && yum clean all

RUN test -n $USE_GO_VERSION_FROM_WEBSITE \
    && cd /tmp \
    && wget --no-verbose https://dl.google.com/go/go1.10.4.linux-amd64.tar.gz \
    && echo "fa04efdb17a275a0c6e137f969a1c4eb878939e91e1da16060ce42f02c2ec5ec go1.10.4.linux-amd64.tar.gz" > checksum \
    && sha256sum -c checksum \
    && tar -C /usr/local -xzf go1.10.4.linux-amd64.tar.gz \
    && rm -f go1.10.4.linux-amd64.tar.gz
ENV PATH=$PATH:/usr/local/go/bin

ENTRYPOINT ["/bin/bash"]
