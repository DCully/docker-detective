FROM ubuntu:20.04 as ubuntu

RUN mkdir -p /newroot/one && \
    mkdir -p /newroot/aa-two/a/b/c/d/e && \
    mkdir -p /newroot/aa-two/aa-three/four && \
    echo "hello!"  > /newroot/hello.txt && \
    echo "woohoo!" > /newroot/aa-two/a/b/c/d/e/hello-two.txt && \
    echo "whoa"    > /newroot/aa-two/aa-three/aa-hello-three.txt && \
    echo "four!!!" > /newroot/aa-two/aa-three/four/x.txt && \
    echo "howdy"   > /newroot/one/howdy.txt && \
    echo "again"   > /newroot/one/again.txt && \
    mkdir -p /newroot/five

FROM scratch

COPY --from=ubuntu /newroot/ /

CMD ["/hello.txt"]
