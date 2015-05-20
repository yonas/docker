# Porting to FreeBSD
Major milestones for porting docker on FreeBSD are:

* make it compile (DONE)
* make it start as a daemon (DONE)
* load an image and create the container (aka working graphdriver) (DONE)
* run the container (DONE)
* working top\start\stop\kill (aka working execdriver) (IN PROGRESS)
* working networking aka NAT
* working port forward (aka working networkdriver)
* working volumes and links
* working limits
* major code cleanup and steps to push code to docker project

# Running
We dont have working docker image on freebsd, and cross-compile doesn't work wery well, so now we need to compile on FreeBSD directly

First we get the sources

    export GOPATH=`pwd`
    go get github.com/docker/docker
    cd src/docker/docker
    git remote set-url origin https://github.com/kvasdopil/docker.git
    git pull
    
Now build the docker

    sh hack/make/.go-autogen
    cd ../../..
    cp -rp src/github.com/docker/docker/vendor/* .
    go build -tags daemon github.com/docker/docker/docker/docker

This should build the docker executable in current directory. You can run docker with command:
    
    zfs create -o mountpoint=/dk zroot/docker # mounpoint should be short
    ./docker -d -b none -e jail -s zfs -g /dk -D

After the daemon is started we can pull the image and start the container

    ./docker pull kazuyoshi/freebsd-minimal
    ./docker run kazuyoshi/freebsd-minimal echo hello world
   
Interactive mode works too

    ./docker run --it kazuyoshi/freebsd-minimal csh

# List of working commands and features

* attach    - ok
* build
* commit    - bug
* cp
* create    - ok
* diff      - ok (on stopped containers)
* events    - ok
* exec
* export
* history   - ok
* images    - ok
* import
* info      - bug
* inspect   - ok
* kill      - ok
* load
* login     - ok 
* logout    - ok 
* logs      - ok
* pause     - not working (not supported on freebsd)
* port      - not working
* ps        - ok
* pull      - ok
* push
* rename    - ok
* restart   - ok
* rm        - ok
* rmi       - ok
* run       - ok
* save
* search    - ok
* start     - ok
* stats     - not working
* stop      - ok
* tag       - ok
* unpause   - not working (not supported on freebsd)
* version   - ok
* wait      - ok