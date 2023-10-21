FROM golang:1.18

ARG UID
ARG GID

RUN apt update && \
    apt -y install sudo &&\
    addgroup --gid $GID nonroot && \
    adduser --uid $UID --gid $GID --disabled-password --gecos "" nonroot && \
    echo 'nonroot ALL=(ALL) NOPASSWD: ALL' >> /etc/sudoers

RUN mkdir -p /app/migration
RUN chown nonroot:nonroot /app/migration

USER nonroot

RUN go install github.com/rubenv/sql-migrate/...@latest
WORKDIR /app/migration
