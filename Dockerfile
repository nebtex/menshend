FROM ubuntu:16.04

RUN apt update -y && apt install ca-certificates -y
# add menshend
ADD menshend /bin/menshend
RUN chmod +x /bin/menshend

ENV VAULT_ADDR=http://localhost:8200

# copy menshend config
ADD menshend.yml /etc/menshend.yml

# run entrypoint
ENTRYPOINT ["menshend", "server", "-config", "/etc/menshend.yml"]
