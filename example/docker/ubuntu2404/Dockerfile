FROM ubuntu:24.04

USER root
RUN echo root:1234 | chpasswd

RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y \
      	coreutils lsb-release build-essential  openssh-server git libssl-dev wget patchelf && \
    rm -rf /var/lib/apt/lists/*

RUN wget "https://github.com/Kitware/CMake/releases/download/v3.30.3/cmake-3.30.3-linux-x86_64.sh" -O cmake.sh && \
    chmod +x cmake.sh && \
    ./cmake.sh --skip-license --prefix=/usr/local && \
    rm ./cmake.sh 

RUN apt-get purge -y \
    wget && \
    rm -rf /var/lib/apt/lists/*
    
RUN git clone https://github.com/cmakelib/cmakelib.git /cmakelib
RUN echo "export CMLIB_DIR=/cmakelib" >> /root/.bashrc

RUN sed -ri 's/#?PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config
RUN mkdir -p /run/sshd

ENTRYPOINT ["/usr/sbin/sshd", "-D", "-o", "ListenAddress=0.0.0.0"]