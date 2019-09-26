# https://stackoverflow.com/questions/55645868/can-not-use-conf-in-golang-beego-framework-when-docker-multi-stage-build
# build stage
FROM golang:1.11.3  AS build-env

#RUN apk add bash ca-certificates git gcc g++ libc-dev
WORKDIR /go/src/antelope

RUN go get -u github.com/golang/dep/cmd/dep

ARG SSH_PRIVATE_KEY
RUN mkdir -p ~/.ssh && umask 0077 && echo "${SSH_PRIVATE_KEY}" > ~/.ssh/id_rsa \
    && git config --global url."git@bitbucket.org:".insteadOf https://bitbucket.org/ \
    && ssh-keyscan bitbucket.org >> ~/.ssh/known_hosts 

COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure -vendor-only

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/src/antelope/antelope

# final stage
FROM ubuntu:bionic
RUN apt update && apt install  ca-certificates -y && apt install -y openssl && apt install ssh -y

WORKDIR /app
COPY --from=build-env /go/src/antelope/swagger/ /app/swagger/
COPY --from=build-env /go/src/antelope/keys/ /app/keys/
COPY --from=build-env /go/src/antelope/scripts/ /app/scripts/
COPY --from=build-env /go/src/antelope/antelope /app/

EXPOSE 9081

CMD ./antelope