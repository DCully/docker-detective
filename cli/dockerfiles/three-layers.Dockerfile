FROM ubuntu:20.04 as ubuntu

RUN mkdir /newroot

RUN mkdir /newroot/one && \
    echo "hello!"  > /newroot/one/hello.txt \
    && mkdir /newroot/one/more && \
    touch /newroot/one/more/time.txt

RUN mkdir /newroot/two && echo "old!"  > /newroot/two/old.txt

FROM scratch

COPY --from=ubuntu /newroot/one/ /one
COPY --from=ubuntu /newroot/two/ /two

CMD ["/one/hello.txt"]
