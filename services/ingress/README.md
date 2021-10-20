# Emissary Ingress

The following service has the definition for setting up the Emissary Ingress in our kubernetes cluster. Initially this
helm chart only works for an insecure local deployment. 

Dependencies:
* [Helm](https://helm.sh/docs/intro/install/)


## Local

When using locally, our Kind cluster is setup to automatically mount the ports of `8080` for http and `8443` for 
https for the listeners defined in `values.dev.yaml`. Follow the instructions below to deploy to the local Kind cluster. 

### Download Dependencies
Before deploying the chart, ensure you have the dependencies locally
```shell
helm dep update
```

### Installation
To install the ingress within the cluster, run the follow command
```shell
helm install ingress . -f values.dev.yaml
```

### Update
If you make any updates to the helm chart, you must run the following to update the deployed service.
```shell
helm upgrade ingress . -f values.dev.yaml
```