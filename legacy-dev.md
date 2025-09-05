# Legcy development environment

### Create local dev environment

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

### SSH into dev environment

To ssh into the vagrant VM, run
```shell
make devssh
```

> [!WARNING]
> Sometimes the files on your host are not mounted into the VM. This oftwn appears as a file not found error.
> Rerun `make devdeploy`.

> [!WARNING]
> Sometimes the Nomad jobs might fail because of a placement error: "Constraint `${attr.consul.version} semver >= 1.8.0` filtered 1 node".
> Run `make devnomad` to fix this issue.
