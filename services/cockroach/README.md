
# Cockroach DB

Cockroach DB (CRBD) is used as the underlying dataplane for Fynbos. It ensures we only have to manage one cluster for multiple
databases across our services. Currently, it is only setup to be used with the Dev Cluster.

This setup of cockroach is using the CRDB official [StatefulSet configuration](https://github.com/cockroachdb/cockroach/blob/master/cloud/kubernetes/bring-your-own-certs/cockroachdb-statefulset.yaml).
This has been converted into a kustomize components for later overlay use.


Dependencies required
* [cockroachdb](https://www.cockroachlabs.com/docs/stable/install-cockroachdb-linux.html)
* [Kubectl](https://kubernetes.io/docs/tasks/tools/)

## Local Dev with Fynbos KIND

In order to get CRDB running locally with KIND follow these steps steps
* Create certs
* Deploy to cluster
* Init node on cluster
* Configure admin User
  * Deploy a client into cluster
  * Run shell against it
  * Create the admin users


### Create Certs
Cockroach requires a few certs to operate. The below command creates these for you and deploys the required ones to the
cluster.

```shell
./create-certs.sh
```

### Deploy

Deploy cockroach to the dev cluster using kustomize dev overlay

```shell
kubectl apply -k deploy/overlays/dev
```

### Init

The cockroachdb cluster needs to be initialized. Wait for the nodes to be in running state (~60s) before running the 
below command. If it fails, just wait for pods to be ready before running again
```shell
./init.sh
```

### Configure
To configure the cluster you need to run some commands against it. Ensure you are port-forwarding the cluster
first

```shell
kubectl port-forward service/cockroachdb 26257
```

Then run
```shell
cockroach sql --certs-dir=certs --host=localhost:26257 -f dbinit.sql -u root
```

### Access UI
To access the cockroach ui you need to port-forward to the service and use the admin user created above to access it

`user: roach`
`password: roach`

