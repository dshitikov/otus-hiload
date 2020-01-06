#FROM golang:1.13.1-alpine3.10 AS build-env
## Compile Delve
#RUN apk add --no-cache git
#RUN go get github.com/derekparker/delve/cmd/dlv

FROM alpine:3.10
WORKDIR /opt/app

COPY ["bin/build", "bin/build"]
COPY ["migrations", "migrations"]
COPY ["templates", "templates"]
COPY ["scripts", "scripts"]

#COPY --from=build-env /go/bin/dlv /

ENV SERVICE_PORT=8080
ENV STORAGE_DIR='/opt/app/storage'

RUN mkdir -p ${STORAGE_DIR} && chown 65535 ${STORAGE_DIR}

EXPOSE $SERVICE_PORT
#40000

# Allow delve to run on Alpine based containers.
#RUN apk add --no-cache libc6-compat

USER 65535:65535

CMD ["./scripts/startup.sh"]

# Run delve
#CMD ["/dlv", "--listen=:40000", "--headless=true", "--api-version=2", "exec", "/opt/app/bin/build"]
