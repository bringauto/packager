FROM debian:12.0

USER root
RUN echo root:1234 | chpasswd

RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y \
      build-essential binutils lsb-release make git libssl-dev openssh-server wget patchelf software-properties-common flex && \
    rm -rf /var/lib/apt/lists/*

RUN git clone --depth=1 --single-branch --branch=releases/gcc-13.3.0 git://gcc.gnu.org/git/gcc.git /gcc && \
    cd /gcc && \
    contrib/download_prerequisites && \
    mkdir objdir

RUN cd /gcc/objdir && \
    ../configure --enable-languages=c,c++ --disable-multilib && \
    make -j$(nproc) && \
    make install

RUN wget "https://github.com/Kitware/CMake/releases/download/v3.25.2/cmake-3.25.2-linux-x86_64.sh" -O cmake.sh && \
    chmod +x cmake.sh && \
    ./cmake.sh --skip-license --prefix=/usr/local && \
    rm ./cmake.sh

RUN apt-get update && \
    apt-get purge -y \
      wget && \
      rm -rf /var/lib/apt/lists/*

RUN git clone https://github.com/cmakelib/cmakelib.git /cmakelib
RUN echo "export CMLIB_DIR=/cmakelib" >> /root/.bashrc && \
    echo "export CC=/usr/local/bin/gcc" >> /root/.bashrc && \
    echo "export CXX=/usr/local/bin/g++" >> /root/.bashrc && \
    echo "export LD_LIBRARY_PATH=/usr/local/lib64" >> /root/.bashrc

RUN sed -ri 's/#?PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config
RUN mkdir -p /run/sshd

ENTRYPOINT ["/usr/sbin/sshd", "-D", "-o", "ListenAddress=0.0.0.0"]
