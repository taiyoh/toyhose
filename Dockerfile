FROM golang:1.13.8-buster AS toyhose-builder

ADD . /app
WORKDIR /app

RUN make docker

FROM scratch

COPY --from=toyhose-builder /bin/toyhose ./toyhose

CMD [ "./toyhose" ]
