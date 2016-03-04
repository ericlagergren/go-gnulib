### Notice:

Because Fadvise was submitted to Go's new sys repository, the Fadvise64
functions in *this* repository are now depreciated. From now on out, all
Go-coreutils functions using Fadvise will be using Go's sys/unix package.

Here's how to install the new package:

- go get golang.org/x/sys/unix
- cd $GOPATH/src/golang.org/x/sys/unix
- export $GOOS # YOUR GOOS HERE... e.g., linux
- export $GOARCH # YOUR GOARCH HERE... e.g., amd64
- ./mkall.sh
- go install
