# Mimsrv
An image server and UI for a simple web photo viewer

## Quick Start

### Set up the environment

1. [Download](https://golang.org/dl/) and [install](https://golang.org/doc/install) Go
   * You may be able to install from your OS's system repository,
     if it is a new enough version; if not, use the
     download and install links on the above line
   * on Fedora run `sudo dnf install golang`
   * on debian run `sudo apt-get install golang`
1. Install git: `sudo dnf install git` or `sudo apt-get install git`
1. Install nodejs and npm using your package manager.
   * on debian, npm may not be in the apt repository, so:
       ```
       curl -sL https://deb.nodesource.com/setup_9.x | sudo -E bash -
       sudo apt-get install nodejs
       sudo apt-get install npm
       ```
1. Install typescript: `sudo npm install -g typescript`
1. Install bower: `sudo npm install -g bower`
1. Install polymer cli: `sudo npm install -g polymer-cli --unsafe-perm`

### Download mimsrv and dependencies

1. Set your GOPATH so `go get` knows where to put files:
   `export GOPATH=$HOME/go`
1. Download this repo and its go dependencies: `go get github.com/jimmc/mimsrv`
    * If you get an error, you may have an old version of `go`
1. cd into the mimsrv directory for the remaining work: `cd ~/go/src/github.com/jimmc/mimsrv`
1. Download polymer dependencies: `(cd _ui && bower install)`

### Build and test

1. Build the server: `go build`
1. Run the server tests: `go test ./...`
1. Compile the UI typescript to javascript: `(cd _ui && tsc)`
1. Build the UI polymer bundle: `(cd _ui && polymer build)`

### Try it out

1. Run the demo server: `./demo`
1. Open [localhost:8021](http://localhost:8021/), login in as `user1` with password `pw1`,
   or as `editor` with password `pwe` to be able to use the editing features.
1. Click on the right-arrow to expand or collapse a folder;
   click on the three little horizontal lines to open the menu
