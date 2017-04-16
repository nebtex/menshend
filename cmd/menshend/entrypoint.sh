#!/usr/local/bin/dumb-init /bin/ash

set -e

chown menshend:menshend ${MENSHEND_CONFIG_FILE}

if [ "$1" == "port-forward" ]; then
  su-exec menshend menshend port-forward --port 8787
else
  su-exec menshend /bin/menshend server --port 8787 --address 0.0.0.0
fi;
