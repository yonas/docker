# Porting to FreeBSD
I'm trying to make docker work on freebsd

Major milestones for porting docker on FreeBSD are:

* make it compile (DONE)
* make it start as a daemon (DONE)
* load an image and create the container (aka working graphdriver) (DONE)
* run the container (DONE)
* working top\start\stop\kill (aka working execdriver) (DONE)
* working networking aka NAT (IN PROGRESS)
* working port forward (aka working networkdriver)
* working volumes and links
* working limits
* major code cleanup and steps to push code to docker project

(See the bigger list below)

# Running
We dont have working docker image on freebsd, and cross-compile doesn't work wery well, so now we need to compile on FreeBSD directly

Prereqesites

    # pkg install go
    # pkg install git
    # pkg install sqlite3
    # pkg install bash
    # pkg install ca_root_nss # use this if pull command is not working

First we get the sources
    
    # git clone https://github.com/kvasdopil/docker 
    # cd docker
    # git checkout freebsd-compat
    
Now build the binary    

    # setenv AUTO_GOPATH 1
    # ./hack/make.sh binary 

This should build the docker executable ./bundles/latest/binary/docker. Now run the daemon:

    # zfs create -o mountpoint=/dk zroot/docker # mounpoint should be short
    # ./bundles/latest/binary/docker -d -e jail -s zfs -g /dk -D

After the daemon is started we can pull the image and start the container

    # ./bundles/latest/binary/docker pull lexaguskov/bsd-minimal 
    # ./bundles/latest/binary/docker run lexaguskov/bsd-minimal echo hello world
   
Interactive mode works too

    # ./bundles/latest/binary/docker run -it lexaguskov/bsd-minimal csh

# Retrieving real FreeBSD image

Since "docker push" command is not working, we have to obtain the image somewhere else.

    # fetch http://download.a-real.ru/freebsd.10.1.amd64.img.txz
    # tar xf freebsd.10.1.amd64.img.txz
    # ./bundles/latest/binary/docker import - freebsd:10.1 < bsd.img

    Now we can test networking etc.

    # ./bundles/latest/binary/docker run -it freebsd:10.1 ifconfig lo1

# Networking

Now the docker can setup basic networking, but not nat

    # kldload pf.ko

    # echo "nat on {you-external-interface} from 172.17.0.0/16 to any -> ({your-external-interface})" > /etc/pf.conf
    # pfctl -f /etc/pf.conf
    # pfctl -e

    # ./bundles/latest/binary/docker run -it freebsd:10.1 ping ya.ru # this should work

# List of working commands and features

Commands:
* attach    - ok
* build
* commit    - ok
* cp        - ok
* create    - ok
* diff      - ok
* events    - ok
* exec      - ok
* export    - ok
* history   - ok
* images    - ok
* import    - ok
* info      - bug
* inspect   - ok
* kill      - ok
* load      - not working
* login     - ok
* logout    - ok
* logs      - ok
* pause     - not working (not supported on freebsd)
* port      - not working
* ps        - ok
* pull      - ok
* push      - not working (server 500 error)
* rename    - ok
* restart   - ok
* rm        - ok
* rmi       - ok
* run       - ok
* save      - ok
* search    - ok
* start     - ok
* stats     - not working
* stop      - ok
* tag       - ok
* top       - ok
* unpause   - not working (not supported on freebsd)
* version   - ok
* wait      - ok

Features:
* image loading         - ok
* container creation    - ok
* container stop\start  - ok
* build on FreeBSD 10.1 - ok
* NAT                   - partial support
* port forward          - not working
* volumes               - not working
* links                 - not working
* limits                - not working

# Participating

If you wish to help, you can join IRC channel #freebsd-docker on freenode.net. 

Now we have following issues:
* not working "docker load"
* on "docker push" the hub returns the error
* the codebase must be syncronized with docker master branch (they have replaced networkdriver with a library)
* netlink functions from libcontainer are not working
* docker can't load (pull, import or commit) an image if not started from build path

Current progress is focused on networking, NAT and port forwarding.
