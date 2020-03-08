FROM scratch

ARG version

ADD ./pkg/${version}/toyhose_linux_amd64/toyhose .

CMD [ "./toyhose" ]
