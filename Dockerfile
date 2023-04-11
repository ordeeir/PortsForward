FROM golang:1.17-alpine
RUN apk add build-base

WORKDIR /porforward

COPY go.mod ./
COPY go.sum ./

COPY . .

CMD ["/bin/sh" ,"-c" ,"go mod download"]

RUN go build -o ./porforward

EXPOSE 443

#CMD tail -f /dev/null

ENTRYPOINT [ "./porforward" ]

