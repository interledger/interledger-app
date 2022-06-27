This folder deploys cockroachdb into the Dev Cluster. The certs behind it are backed by vault in the shared cluster.
Once running this the first time it is important to initialize the cluster in one of the pods by running

```shell
cockroach init --certs-dir=cockroach-client-certs/
```

This will initialize the whole cluster using the root certs. This could be moved to a job later but for now it is
easier to run on creation.