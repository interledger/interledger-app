# Interledger-app

### Create local dev environment

Add a value to  COOKIE_KEY = "" in /dev/nomad/rafiki.hcl
This spins up a Nomad and Consul deployment in a VM using VirtualBox and Vagrant.
```shell
make devup
```

### If running first time 
 Generate a asset in rafiki (https://rafiki.mgnt.fynbos.test/assets)
 copy the ID to /go/backend/rafiki/external/client.go <replace-me>

### Delete local dev environment

To delete the local environment run the following command
```shell
make devdown
```

### Rerunning deployment 

To deploy all the services or when deployment config has changed, run
```shell
make devdeploy
```
*NB!* Sometimes the files on your host are not mounted into the VM. This often appears
as a file not found error. Rerun `make devdeploy` or `(cd dev/vagrant && vagrant reload)`.

### SSH into dev environment

To ssh into the vagrant VM, run
```shell
make devssh
```
