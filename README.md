# Fynbos

### Create local dev environment

*NB!* Kill docker-desktop if you have it installed / runing

This spins up a Nomad and Consul deployment in a VM using VirtualBox and Vagrant.
```shell
make devup
```

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
