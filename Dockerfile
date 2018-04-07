FROM alpine:edge
RUN apk --no-cache add ca-certificates
RUN update-ca-certificates

ADD ./tmpls /tmpls
ADD pw /

EXPOSE 8080
ENTRYPOINT ["/pw"]
