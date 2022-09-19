FROM ubuntu:22.04

# This results in the contents of /var/lib/apt/lists/
# remaining in the layers but absent from the image.
RUN apt-get update && apt-get install -y \
    apt-transport-https \
    ca-certificates \
    curl \
    iputils-ping \
    gnupg \
    lsb-release
RUN rm -rf /var/lib/apt/lists/*

# This results in two copies of docker-compose in the layers vs only one in the image.
RUN curl -L "https://github.com/docker/compose/releases/download/1.29.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
RUN chmod +x /usr/local/bin/docker-compose
