FROM alpine:3.10
WORKDIR /opt/app

COPY ["bin/build", "bin/build"]
COPY ["migrations", "migrations"]
COPY ["templates", "templates"]
COPY ["scripts", "scripts"]

ENV SERVICE_PORT=8080
ENV STORAGE_DIR='/opt/app/storage'

RUN mkdir -p ${STORAGE_DIR} && chown 65535 ${STORAGE_DIR}

EXPOSE $SERVICE_PORT

USER 65535:65535

CMD ["./scripts/startup.sh"]
