FROM golang:1.21.1 as base

# create dev stage from base stage
FROM base as dev
# install air binary
RUN go install github.com/cosmtrek/air@latest

# run air and set work dir
WORKDIR /opt/app/api
RUN git config --global --add safe.directory /opt/app/api
CMD ["air"]