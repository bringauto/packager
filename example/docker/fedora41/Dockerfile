FROM fedora:41

USER root
RUN echo root:1234 | chpasswd

RUN dnf -y update && \
    dnf -y install  \
    automake binutils gcc gcc-c++ git kernel-devel lsb-release make openssh-server openssl-devel patchelf wget \
    && dnf clean all

RUN wget "https://github.com/Kitware/CMake/releases/download/v3.30.3/cmake-3.30.3-linux-x86_64.sh" -O cmake.sh && \
    chmod +x cmake.sh && \
    ./cmake.sh --skip-license --prefix=/usr/local && \
    rm ./cmake.sh

RUN dnf -y update && \
    dnf -y remove \
      wget

RUN git clone https://github.com/cmakelib/cmakelib.git /cmakelib
RUN echo "export CMLIB_DIR=/cmakelib" >> /root/.bashrc

RUN sed -ri 's/#?PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config
RUN mkdir -p /run/sshd

RUN ssh-keygen -A

ENTRYPOINT ["/usr/sbin/sshd", "-D", "-o", "ListenAddress=0.0.0.0"]
