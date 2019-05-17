# clustertest
\[WIP\] Clustertest is an automated testing system for clustered system.

## Requirement
* Linux Server
* Proxmox VE nodes

## Installation
```bash
$ go get -u github.com/yuuki0xff/clustertest/cmd/clustertest
$ go get -u github.com/yuuki0xff/clustertest/cmd/clustertestd
```

## Command Usage
* `clustertest job create`
* `clustertest job inspect [ID-or-Name]`
* `clustertest job wait [ID-or-Name]`
* `clustertest job cancel [ID-or-Name]`
* `clustertest job output [ID-or-Name]`
* `clustertest job delete [ID-or-Name]`

## Example
```bash
$ clustertest job create hello-world.clustertest.conf
$ clustertest job wait hello-world
$ clustertest job output hello-world
hello world
```
