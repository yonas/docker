#!/usr/bin/env bash
#set -e

cd "$(dirname "$BASH_SOURCE")/.."

# Downloads dependencies into vendor/ directory
mkdir -p vendor
cd vendor

clone() {
	vcs=$1
	pkg=$2
	rev=$3

	pkg_url=https://$pkg
	target_dir=src/$pkg

	echo -n "$pkg @ $rev: "

	if [ -d $target_dir ]; then
		echo -n 'rm old, '
		rm -fr $target_dir
	fi

	echo -n 'clone, '
	case $vcs in
		git)
			git clone --quiet --no-checkout $pkg_url $target_dir
			( cd $target_dir && git reset --quiet --hard $rev )
			;;
		hg)
			hg clone --quiet --updaterev $rev $pkg_url $target_dir
			;;
	esac

	echo -n 'rm VCS, '
	( cd $target_dir && rm -rf .{git,hg} )

	echo -n 'rm vendor, '
	( cd $target_dir && rm -rf vendor Godeps/_workspace )

	echo done
}

# the following lines are in sorted order, FYI
clone git github.com/Sirupsen/logrus v0.8.2 # logrus is a common dependency among multiple deps
clone git github.com/docker/libtrust 230dfd18c232
clone git github.com/go-check/check 64131543e7896d5bcc6bd5a76287eb75ea96c673
clone git github.com/gorilla/context 14f550f51a
clone git github.com/gorilla/mux e444e69cbd
clone git github.com/kr/pty 5cf931ef8f
clone git github.com/mistifyio/go-zfs v2.1.1
clone git github.com/tchap/go-patricia v2.1.0
clone hg code.google.com/p/go.net 84a4013f96e0
clone hg code.google.com/p/gosqlite 74691fb6f837

#get libnetwork packages
clone git github.com/docker/libnetwork a09a60a123b4106954e5b42ed5ec41f264eeb8bf
clone git github.com/vishvananda/netns 008d17ae001344769b031375bdb38a86219154c6
clone git github.com/vishvananda/netlink 8eb64238879fed52fd51c5b30ad20b928fb4c36c

clone git github.com/BurntSushi/toml f706d00e3de6abe700c994cdd545a1a4915af060
clone git github.com/deckarep/golang-set ef32fa3046d9f249d399f98ebaf9be944430fd1d

#libnetwork uses swarm discovery and store libraries
clone git github.com/docker/swarm 54dfabd2521314de1c5b036f6c609efbe09df4ea
mv src/github.com/docker/swarm/discovery tmp-discovery
mv src/github.com/docker/swarm/pkg/store tmp-store
rm -rf src/github.com/docker/swarm
mkdir -p src/github.com/docker/swarm
mv tmp-discovery src/github.com/docker/swarm/discovery
mkdir -p src/github.com/docker/swarm/pkg
mv tmp-store src/github.com/docker/swarm/pkg/store

clone git github.com/hashicorp/consul 954aec66231b79c161a4122b023fbcad13047f79
mv src/github.com/hashicorp/consul/api tmp-api
rm -rf src/github.com/hashicorp/consul
mkdir -p src/github.com/hashicorp/consul
mv tmp-api src/github.com/hashicorp/consul/api

clone git github.com/coreos/go-etcd 73a8ef737e8ea002281a28b4cb92a1de121ad4c6
mv src/github.com/coreos/go-etcd/etcd tmp-etcd
rm -rf src/github.com/coreos/go-etcd
mkdir -p src/github.com/coreos/go-etcd
mv tmp-etcd src/github.com/coreos/go-etcd/etcd

clone git github.com/samuel/go-zookeeper d0e0d8e11f318e000a8cc434616d69e329edc374
mv src/github.com/samuel/go-zookeeper/zk tmp-zk
rm -rf src/github.com/samuel/go-zookeeper
mkdir -p src/github.com/samuel/go-zookeeper
mv tmp-zk src/github.com/samuel/go-zookeeper/zk

# get distribution packages
clone git github.com/docker/distribution b9eeb328080d367dbde850ec6e94f1e4ac2b5efe
mv src/github.com/docker/distribution/digest tmp-digest
mv src/github.com/docker/distribution/registry/api tmp-api
rm -rf src/github.com/docker/distribution
mkdir -p src/github.com/docker/distribution
mv tmp-digest src/github.com/docker/distribution/digest
mkdir -p src/github.com/docker/distribution/registry
mv tmp-api src/github.com/docker/distribution/registry/api

clone git github.com/docker/libcontainer 57a50dd378e66d234faef0e61fa98718371156ff
# libcontainer deps (see src/github.com/docker/libcontainer/update-vendor.sh)
clone git github.com/coreos/go-systemd v2
clone git github.com/godbus/dbus v2
clone git github.com/syndtr/gocapability 66ef2aa7a23ba682594e2b6f74cf40c0692b49fb
clone git github.com/golang/protobuf 655cdfa588ea
