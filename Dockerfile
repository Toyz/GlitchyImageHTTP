FROM scratch    
ADD ca-certificates.crt /etc/ssl/certs/

ADD ./tmpls /tmpls
#temp
ADD ./uploads /uploads

ADD pw /

EXPOSE 8080
ENTRYPOINT ["/pw"]
