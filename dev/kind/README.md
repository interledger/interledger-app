# Kind for Local Dev

At Fynbos we use [ctlptl](https://github.com/tilt-dev/ctlptl) for locally development of our Kubernetes Cluster. In order to install 
kind on your machine follow the [docs](https://github.com/tilt-dev/ctlptl). 

The following dependencies are required: 
* [Docker](https://docs.docker.com/get-started/overview/)
* [Kubectl](https://kubernetes.io/docs/tasks/tools/)

## Create

To create a local cluster run the following

```shell
./run.sh
```

This will create a local Kind cluster called `fynbos-dev`

## Destroy

To delete the cluster run the following

```shell
./nuke.sh
```