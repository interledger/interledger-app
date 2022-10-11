## Architecture
The control plane is deployed in an AWS managed VPC. We then deploy managed node groups in 
our own VPC.

## After initial deploy

### Export kubeconfig and store locally
```shell
pulumi stack -s fynbos/main output kubeconfig | jq > kubeconfig.yaml
mv kubeconfig.yaml ~/.fynbos/aws/prod-cluster.yaml
```
### Delete `aws-node` daemonset
We need to remove the `aws-node` daemonset to ensure we can deploy our Cilium service mesh.


