# Porting to FreeBSD
Major milestones for porting docker on FreeBSD are:

* make it compile (DONE)
* make it start as a daemon (DONE)
* load an image and create the container (DONE)
* run the container 
* working top\start\stop\kill
* working networking aka NAT
* working port forward
* working volumes and links
* major code cleanup and steps to push code to docker project

# Running
We dont have working docker image on freebsd, and cross-compile doesn't work wery well, so now we need to compile on FreeBSD directly

    export GOPATH=`pwd`
    go get github.com/kvasdopil/docker
    cp -rp src/github.com/kvasdopil/docker/vendor/* .
    
    go build -tags daemon github.com/kvasdopil/docker/docker/docker
    
