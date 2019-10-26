FROM golang:stretch

RUN apt update && apt install git python3 python3-pip -y
RUN pip3 install scrapy setuptools
RUN go get github.com/alash3al/scrapyd-go

WORKDIR /root/

ENTRYPOINT [ "scrapyd-go" ]
