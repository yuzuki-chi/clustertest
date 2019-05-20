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
* `clustertest task create`
* `clustertest task inspect [ID-or-Name]`
* `clustertest task wait [ID-or-Name]`
* `clustertest task cancel [ID-or-Name]`
* `clustertest task output [ID-or-Name]`
* `clustertest task delete [ID-or-Name]`

## Example
```bash
$ clustertest job create hello-world.clustertest.yaml
$ clustertest job wait hello-world
$ clustertest job output hello-world
hello world
```
