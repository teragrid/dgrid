FROM alpine:3.7
MAINTAINER Greg Szabo <greg@teragrid.com>

RUN apk update && \
    apk upgrade && \
    apk --no-cache add curl jq file

VOLUME [ /teragrid ]
WORKDIR /teragrid
EXPOSE 46656 46657
ENTRYPOINT ["/usr/bin/wrapper.sh"]
CMD ["node", "--proxy_app dummy"]
STOPSIGNAL SIGTERM

COPY wrapper.sh /usr/bin/wrapper.sh

