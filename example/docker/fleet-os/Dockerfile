FROM debian:11.2

USER root
RUN echo root:1234 | chpasswd

#
# Install Base dependencies into the docker container
#
RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y \
      xz-utils sed git libssl-dev openssh-server wget unzip python3 build-essential && \
    rm -rf /var/lib/apt/lists/*

RUN wget "https://github.com/Kitware/CMake/releases/download/v3.22.2/cmake-3.22.2-linux-x86_64.sh" -O cmake.sh && \
    chmod +x cmake.sh && \
    ./cmake.sh --skip-license --prefix=/usr/local && \
    rm ./cmake.sh

RUN git clone https://github.com/cmakelib/cmakelib.git /cmakelib
RUN echo "export CMLIB_DIR=/cmakelib" >> /environment.sh

RUN sed -ri 's/#?PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config
RUN mkdir -p /run/sshd

#
# Install our BringAuto Fleet
#
RUN mkdir -p /root/toolchain-install && \
    wget https://gitlab.bringauto.com/bringauto-public/fleet-os-toolchain/-/raw/master/pre-release/toolchain.bringauto-image-compile.ba-raspberrypi4-64.deploy_development.zip \
    -O /root/toolchain-install/oecore.sh.zip
RUN cd /root/toolchain-install/ && \
    unzip oecore.sh.zip && \
    rm oecore.sh.zip && \
    chmod +x *
RUN /root/toolchain-install/* -S -y -d /root/toolchain/ || exit 0
RUN echo ". /root/toolchain/environment-setup-cortexa72-oe-linux" >> /environment.sh
RUN rm -r /root/toolchain-install/

RUN wget https://github.com/bringauto/packager/releases/download/v0.2.0-rc/bringauto-packager-tools_v0.2.0_x86-64-linux.zip \
    -O /root/bringauto-packager-tools_v0.2.0_x86-64-linux.zip

RUN cd /root && \
    unzip bringauto-packager-tools_v0.2.0_x86-64-linux.zip && \
    mv bap_tools_0.2.0_x86-64-linux/* /root/ && \
    rm -r bap_tools_0.2.0_x86-64-linux

COPY lsb_release.txt /root/
COPY uname.txt /root/
RUN chmod +x /root/lsb_release
RUN chmod +x /root/uname
RUN echo 'PATH=/root/:$PATH' >> /root/.bashrc

RUN apt-get update && \
    apt-get purge -y \
      wget unzip && \
      rm -rf /var/lib/apt/lists/*

ENTRYPOINT ["/usr/sbin/sshd", "-D", "-o", "ListenAddress=0.0.0.0"]
