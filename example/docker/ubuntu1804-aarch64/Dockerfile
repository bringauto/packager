FROM ubuntu:18.04

USER root
RUN echo root:1234 | chpasswd

RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y \
      build-essential make libssl-dev coreutils lsb-release wget g++-8 gcc-8 gcc-8-aarch64-linux-gnu g++-8-aarch64-linux-gnu patchelf unzip openssh-server git && \
    rm -rf /var/lib/apt/lists/*

RUN update-alternatives --install /usr/bin/gcc gcc /usr/bin/gcc-8 60 --slave /usr/bin/g++ g++ /usr/bin/g++-8

RUN wget "https://github.com/Kitware/CMake/releases/download/v3.28.0/cmake-3.28.0-linux-x86_64.sh" -O cmake.sh && \
    chmod +x cmake.sh && \
    ./cmake.sh --skip-license --prefix=/usr/local && \
    rm ./cmake.sh

RUN git clone https://github.com/cmakelib/cmakelib.git /cmakelib
RUN echo "export CMLIB_DIR=/cmakelib" >> /root/.bashrc

RUN sed -ri 's/#?PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config
RUN mkdir -p /run/sshd

RUN sed -ri 's/#?PermitUserEnvironment no/PermitUserEnvironment yes/' /etc/ssh/sshd_config
RUN mkdir /root/.ssh && \
    echo "CXX=${CXX}" > /root/.ssh/environment && \
    echo "CC=${CC}" >> /root/.ssh/environment

COPY init_toolchain.sh /root/
RUN chmod +x /root/init_toolchain.sh && \
    /root/init_toolchain.sh /root/tools
COPY lsb_release.txt /root/tools/
COPY uname.txt /root/tools/

ENV CXX=aarch64-linux-gnu-g++-8
ENV CC=aarch64-linux-gnu-gcc-8

ENTRYPOINT ["/usr/sbin/sshd", "-D", "-o", "ListenAddress=0.0.0.0"]

