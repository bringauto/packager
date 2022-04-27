FROM debian:11.2

USER root
RUN echo root:1234 | chpasswd

#
# Install Base dependencies into the docker container
#
RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y \
      xz-utils sed git libssl-dev openssh-server wget && \
    rm -rf /var/lib/apt/lists/*

RUN wget "https://github.com/Kitware/CMake/releases/download/v3.22.2/cmake-3.22.2-linux-x86_64.sh" -O cmake.sh && \
    chmod +x cmake.sh && \
    ./cmake.sh --skip-license --prefix=/usr/local && \
    rm ./cmake.sh

RUN apt-get update && \
    apt-get purge -y \
      wget && \
      rm -rf /var/lib/apt/lists/*

RUN git clone https://github.com/cmakelib/cmakelib.git /cmakelib
RUN echo "export CMLIB_DIR=/cmakelib" >> /environment.sh

RUN sed -ri 's/#?PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config
RUN mkdir -p /run/sshd

#
# Install our autonomy-host
#
COPY oecore.sh /
RUN chmod +x oecore.sh && ./oecore.sh -S -y -d /root/toolchain || exit 0
RUN echo ". /root/toolchain/environment-setup-cortexa72-oe-linux" >> /environment.sh
RUN rm oecore.sh

COPY lsb_release lsb_release.json uname uname.json /root/
RUN chmod +x /root/lsb_release
RUN chmod +x /root/uname
RUN echo 'PATH=/root/:$PATH' >> /root/.bashrc

ENTRYPOINT ["/usr/sbin/sshd", "-D", "-o", "ListenAddress=0.0.0.0"]
