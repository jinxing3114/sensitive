FROM alpine:latest

WORKDIR /sensitive

ADD sensitive sensitive

EXPOSE 8888

ENTRYPOINT [ "sh", "-c", "./sensitive "]
