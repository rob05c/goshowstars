# Dockerfile to build a goservestars container on Ubuntu

FROM ubuntu
MAINTAINER Robert Butts
#RUN apt-get update

RUN mkdir -p /usr/bin/
RUN mkdir -p /data/

# port to serve on. Change this if you like
EXPOSE 9090

ADD goshowstars /usr/bin/goshowstars
ADD index.html /data/index.html
ADD startemplate.html /data/startemplate.html

ENTRYPOINT ["/usr/bin/goshowstars", "-d", "robert-butts.me:8081"] 
