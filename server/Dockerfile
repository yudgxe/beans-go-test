FROM ubuntu:latest

RUN apt update && apt install openssh-server sudo zip -y
RUN useradd -rm -d /home/dev -s /bin/bash -g root -G sudo -u 1000 dev
RUN echo 'dev:123' | chpasswd
RUN service ssh start

EXPOSE 22

CMD ["/usr/sbin/sshd","-D"]