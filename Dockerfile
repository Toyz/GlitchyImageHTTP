FROM alpine:edge
RUN apk --no-cache add ca-certificates
RUN update-ca-certificates

ADD ./tmpls /tmpls
ADD pw /

#RUN export RUNNING_MODE=release && MEMCACHE_HOST="172.17.0.3:11211"
EXPOSE 8080
ENTRYPOINT ["/pw"]
