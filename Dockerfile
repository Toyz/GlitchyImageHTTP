FROM scratch    
ADD /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ADD ./tmpls /tmpls
ADD pw /

EXPOSE 8080
ENTRYPOINT ["/pw"]
