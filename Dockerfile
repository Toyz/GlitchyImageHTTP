#FROM alpine:edge
#RUN apk --no-cache add ca-certificates
#RUN update-ca-certificates
FROM scratch
ADD ca-certificates.crt /etc/ssl/certs/

ADD ./assets /assets

ADD pw /

EXPOSE 8080
ENTRYPOINT ["/pw"]
