FROM ubuntu:bionic


RUN apt-get update && apt-get install wget -y
RUN wget https://dl.google.com/go/go1.11.1.linux-amd64.tar.gz
RUN tar -xzvf go1.11.1.linux-arm64.tar.gz
#RUN mv go1.11.1.linux-arm64 go
ENV GOROOT /usr/local/go
RUN echo $GOROOT
ENV GOPATH=/go
RUN echo $GOPATH
ENV PATH=$GOPATH/bin:$GOROOT/bin:$PATH
RUN echo $PATH
# Set our workdir to our current service in the gopath
WORKDIR /go/src/antelope/
# Copy the current code into our workdir
COPY . .

RUN go get github.com/astaxie/beego
RUN go get github.com/beego/bee
RUN go get -u github.com/golang/dep/cmd/dep

# Create a dep project, and run `ensure`, which will pull in all
# of the dependencies within this directory.
RUN dep ensure

EXPOSE 9081

CMD bee run -downdoc=true -gendoc=true