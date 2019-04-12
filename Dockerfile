FROM ubuntu:bionic

WORKDIR /go/src/antelope/
#RUN apt-get install wget
RUN wget https://dl.google.com/go/go1.11.1.linux-arm64.tar.gz
RUN tar -xzvf go1.11.1.linux-arm64.tar.gz
RUN mv go1.11.1.linux-arm64 go
ENV GOROOT=/usr/local/go
ENV GOPATH=$HOME/go
RUN PATH=$GOPATH/bin:$GOROOT/bin:$PATH

# Set our workdir to our current service in the gopath

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