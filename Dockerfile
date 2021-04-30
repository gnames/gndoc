FROM alpine:3.13

LABEL maintainer="Dmitry Mozzherin"

ENV LAST_FULL_REBUILD 2021-06-07

WORKDIR /bin

COPY ./gnfinder-doc/gnfinder-doc /bin

ENTRYPOINT [ "gnfinder-doc" ]

CMD ["-p", "8777"]
