FROM golang:1.22-bullseye AS builder

COPY ./ /app
WORKDIR /app

# Adjust the SQL file to an executable sql-migration file.
RUN echo "" > /app/build/support-files/sql/0001_iam_20200327-1442_mysql.sql
RUN sed -i "1 i -- +migrate Up" /app/build/support-files/sql/*
RUN sed -i 's/`bkiam`.//g' /app/build/support-files/sql/*

# Go build
ARG BINARY=iam
RUN make build && chmod +x ${BINARY}

FROM debian:bullseye-slim

ARG BINARY=iam
RUN mkdir -p /app/logs
COPY --from=builder /app/${BINARY} /app/${BINARY}
COPY --from=builder /app/build/support-files/sql /app/sql
COPY --from=builder /app/build/sql-migrate /app/sql-migrate

CMD ["/app/iam", "-c", "/app/config.yaml"]
