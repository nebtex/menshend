FROM ubuntu:16.04

# Installation:
# Import MongoDB public GPG key AND create a MongoDB list file
RUN apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv EA312927
RUN apt-get update -y && apt-get install -y --no-install-recommends software-properties-common
RUN echo "deb http://repo.mongodb.org/apt/ubuntu $(cat /etc/lsb-release | grep DISTRIB_CODENAME | cut -d= -f2)/mongodb-org/3.2 multiverse" | tee /etc/apt/sources.list.d/mongodb-org-3.2.list

RUN apt-get update -y && apt-get install -y wget curl unzip\
 mongodb-org python-pip python-dev libffi-dev libssl-dev libxml2-dev libxslt1-dev libjpeg8-dev zlib1g-dev\
 nodejs npm
RUN mkdir -p /data/db
RUN pip install --upgrade pip

# install consul
RUN wget https://releases.hashicorp.com/consul/0.7.5/consul_0.7.5_linux_amd64.zip
RUN unzip consul_0.7.5_linux_amd64.zip && cp consul /bin/consul && chmod +x /bin/consul

# install butterfly
RUN pip install butterfly

# install supervisor
RUN pip install supervisor

# install vault
RUN wget https://releases.hashicorp.com/vault/0.7.0/vault_0.7.0_linux_amd64.zip
RUN unzip vault_0.7.0_linux_amd64.zip && cp vault /bin/vault && chmod +x /bin/vault

# install npm

# instlall json server
RUN npm install -g json-server

# add web ui (for test crf)
