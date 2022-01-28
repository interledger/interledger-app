# Fynbos

## Local Dev

Below is roughly how local dev works. This document will be updated as the process evolves. Generally
you will only need to run the Create Cluster and Initialize Services once or when new configs are added.


### Update Host File

Ensure you have updated your host (`/etc/hosts`) to point `fynbos.test` to `127.0.0.1`

```shell
127.0.0.1      fynbos.test
```

### Create local cluster

To create the local Kind cluster and registry run the following command
```shell
make kindup
```

### Delete local cluster

To delete the local Kind cluster and registry run the following command
```shell
make kinddown
```

### Initialize services 

To initialize all the services within the cluster run 
```shell
make kindpulumiup
```

This will run a local pulumi instance to deploy all services to the cluster

### Run Tilt

To run Tilt for your local inner dev loop run

```shell
make tiltup
```

