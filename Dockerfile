FROM centurylink/ca-certs

ADD ./tmpls /tmpls
ADD pw /

EXPOSE 8080
ENTRYPOINT ["/pw"]
