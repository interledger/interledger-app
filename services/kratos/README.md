# Kratos

The following service has the definition for setting up Kratos in our kubernetes cluster. Initially this
helm chart only works for local deployment

Dependencies:
* [Helm](https://helm.sh/docs/intro/install/)
* [cockroachdb](https://www.cockroachlabs.com/docs/stable/install-cockroachdb-linux.html)


## Local Dev with Fynbos KIND

In order to get Kratos running locally with KIND follow these steps
* Create certs
* Deploy to cluster 

### Create Certs
Kratos requires certs to operate. The below command creates these for you and deploys the required ones to the
cluster. Ensure the certs in the `services/cockroach` have been created first!

```shell
./create-certs.sh
```

### Download Dependencies
Before deploying the chart, ensure you have the dependencies locally
```shell
helm dep update
```

### Installation
To install the ingress within the cluster, run the follow command
```shell
helm install kratos . -f values.dev.yaml
```

### Update
If you make any updates to the helm chart, you must run the following to update the deployed service.
```shell
helm upgrade kratos . -f values.dev.yaml
```