FROM ubuntu:22.04

RUN echo "hello!"

RUN apt-get update && apt-get install -y iputils-ping

RUN rm -rf /var/lib/apt/lists/*

CMD ["echo"]
