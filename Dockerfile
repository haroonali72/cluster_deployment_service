FROM golang:1.11.3 as builder

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