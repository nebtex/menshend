FROM alpine:3.5
MAINTAINER Cristian Lozano <criscalovis@gmail.com> (@criloz)
ENV MENSHEND_CONFIG /etc/menshend.yml
# Create a  user and group first so the IDs get set the same way, even as
# the rest of this may change over time.
RUN addgroup menshend && adduser -S -G menshend menshend

RUN  apk update \
  && apk add --no-cache ca-certificates wget su-exec\
  && update-ca-certificates
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

# install dump-init
RUN wget -O /usr/local/bin/dumb-init https://github.com/Yelp/dumb-init/releases/download/v1.2.0/dumb-init_1.2.0_amd64
RUN chmod +x /usr/local/bin/dumb-init

# add menshend
ADD dist/menshend_linux_amd64 /bin/menshend
RUN chmod +x /bin/menshend

# copy menshend config
ADD menshend.yml /etc/menshend.yml

# run entrypoint
ENTRYPOINT ["menshend", "server", "-config", ${MENSHEND_CONFIG}]
