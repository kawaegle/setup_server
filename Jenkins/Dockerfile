FROM jenkins/jenkins:2.440
USER root
WORKDIR /tmp
RUN curl -O https://download.docker.com/linux/static/stable/x86_64/docker-24.0.7.tgz && tar xzvf docker-24.0.7.tgz && mv docker/* /usr/bin/ && rm -rf docker*
USER jenkins

