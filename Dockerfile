FROM debian:buster-slim

RUN apt-get update && apt-get install -y \
  curl \
  git \
 && apt-get clean \
 && rm -rf /var/lib/apt/lists/*

ENTRYPOINT ["/entrypoint.sh"]

COPY scripts/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

COPY pr-revert_*.deb /tmp/
RUN dpkg -i /tmp/pr-revert_*.deb
