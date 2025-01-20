FROM debian:11.2

USER root
RUN echo root:1234 | chpasswd

ENV DEBIAN_FRONTEND=noninteractive

#
# Install Base dependencies into the docker container
#
RUN apt-get update && \
    apt-get install -y \
      build-essential xz-utils sed git libssl-dev openssh-server unzip python3 make patchelf wget locales perl file \
      findutils && \
    rm -rf /var/lib/apt/lists/*

# Generate the locale for en_US.UTF-8
RUN locale-gen en_US.UTF-8

# Set environment variables for locale
ENV LANG en_US.UTF-8
ENV LANGUAGE en_US:en
ENV LC_ALL en_US.UTF-8

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
COPY init_toolchain.sh /root/
RUN chmod +x /root/init_toolchain.sh && \
    /root/init_toolchain.sh /root/toolchain /root/tools
COPY lsb_release.txt /root/tools/
COPY uname.txt /root/tools/

RUN apt-get update && \
    apt-get purge -y \
      wget unzip && \
      rm -rf /var/lib/apt/lists/*

ENTRYPOINT ["/usr/sbin/sshd", "-D", "-o", "ListenAddress=0.0.0.0"]
