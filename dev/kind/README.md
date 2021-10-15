# Kind for Local Dev

At Fynbos we use [Kind](https://kind.sigs.k8s.io/) for locally development of our Kubernetes Cluster. In order to install 
kind on your machine follow the [docs](https://kind.sigs.k8s.io/docs/user/quick-start/). 

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