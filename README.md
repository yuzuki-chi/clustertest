# clustertest
\[WIP\] Clustertest is an automated testing system for clustered system.

NOTE: Clustertest is developed for test the elton. Currently I'm focusing on elton development. So it is not actively developed.

## Requirement
* Linux Server
* Proxmox VE nodes

## Installation
```bash
$ go install -u github.com/yuuki0xff/clustertest/cmd/clustertest
$ go install -u github.com/yuuki0xff/clustertest/cmd/clustertestd
```

## Command Usage
* `clustertest task run`
* `clustertest task start`
* `clustertest task list`
* `clustertest task wait [ID-or-Name]`
* `clustertest task output [ID-or-Name]`

## Example
Example config: See `clustertest.yaml`.

Example of running task synchronously:

```bash
$ clustertestd &
$ clustertest task run clustertest.yaml
Status: finished
-------------------- Before --------------------
ExitCode: 0
Host: root@192.168.189.77
Start: 2019-06-12 19:26:27.442614574 +0900 JST
End: 2019-06-12 19:26:38.259986893 +0900 JST
Output:
root@192.168.189.77$ echo OK
Warning: Permanently added '192.168.189.77' (ECDSA) to the list of known hosts.
OK
root@192.168.189.77$ hostname
Warning: Permanently added '192.168.189.77' (ECDSA) to the list of known hosts.
proxmox-provisioner-test-test-cluster-test-vm-0
================================
root@192.168.189.78$ echo OK
Warning: Permanently added '192.168.189.78' (ECDSA) to the list of known hosts.
OK
root@192.168.189.78$ hostname
Warning: Permanently added '192.168.189.78' (ECDSA) to the list of known hosts.
proxmox-provisioner-test-test-cluster-test-vm-1
-------------------- Main --------------------
ExitCode: 0
Output:
-------------------- After --------------------
ExitCode: 0
Output:
```

Example of running task asynchronously:

```bash
$ clustertest task start clustertest.yaml
0
$ clustertest task list
ID   Status
-- --------
 0 running
$ clustertest task wait 0
$ clustertest task output 0
Status: finished
-------------------- Before --------------------
(omitted)
-------------------- Main --------------------
ExitCode: 0
Output:
-------------------- After --------------------
ExitCode: 0
Output:
```

## How to use ProxmoxVE provisioner
1. Create templates with cloud-init support.
2. Copy templates to all nodes.  
   Naming conventions for templates: "<original_template_name>-<node_name>"
3. Create clustertest configuration.  See `clustertest.yaml`.
4. `clustertest task start <file_name>`
