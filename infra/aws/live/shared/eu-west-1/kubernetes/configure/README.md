## Shared Responsibility
AWS subscribes to the shared responsibility model. i.e. they are responsible for security of the cloud
whilst we are responsible for the security in the cloud. This extends to the k8s cluster as well.
As mentioned above, they are responsible for securing the control plane in its own VPC and we are
responsible for securing cluster configuration such as network policies, role bindings etc.

Our Kubernetes module helps us to do this by following
[eks best practices](https://aws.github.io/aws-eks-best-practices/security/docs/iam/#kubernetes-service-accounts).
See the module readme for more information.


# 1. Export kubeconfig and cluster oidcId from provision stack
**TODO:** Figure out why we can't import kubeconfig from provision stack and create an eks provider without it erroring out.
First we need to export the kubeconfig from the `provision` project.
```sh
aws-vault exec <shared-account> -- pulumi stack output -C ../provision kubeconfig | jq > kubeconfig.yaml
export KUBECONFIG=./kubeconfig.yaml
```
Next, we need to look up the clusters oidc from the provision stack and set it in this projects config.
```sh
aws-vault exec <shared-account> -- pulumi stack output -C ../provision oidcProvider
# look up the id from the output above and set it in config
pulumi config set cluster:oidcId <id from oidcProvider output>
```

# 2. Import aws-node service account and daemonset
This only needs to be done when the project is being deployed for the first time.
```sh
aws-vault exec <shared-account> -- pulumi import kubernetes:core/v1:ServiceAccount aws-node-sa kube-system/aws-node
aws-vault exec <shared-account> -- pulumi import kubernetes:apps/v1:DaemonSet aws-node-ds kube-system/aws-node
```
**NB** Do not change the names `aws-node-sa` and `aws-node-ds` as the pulumi project is expecting it be named so.

# 3. Pulumi up
Run
```sh
aws-vault exec <shared-account> -- pulumi up
```