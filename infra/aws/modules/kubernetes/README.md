## Shared Responsibility
AWS subscribes to the shared responsibility model. i.e. they are responsible for security of the cloud
whilst we are responsible for the security in the cloud. This extends to the k8s cluster as well.
As mentioned above, they are responsible for securing the control plane in its own VPC and we are
responsible for securing cluster configuration such as network policies, role bindings etc.

In light of this, this module configurations taken from
[eks best practices](https://aws.github.io/aws-eks-best-practices/security/docs/iam/#kubernetes-service-accounts)

### Updating the aws-node daemonset to use IRSA
Service accounts used in conjunction with K8s RBAC allow the assigning of roles to a pod.
Every namespace has a default service account that allows authenticated and unauthenticated 
requests to read K8s api resources. It is best practice to create and assign a service account
to your pods and grant it least privilidge access to k8s resources.

This can be taken one step further by assigning AWS IAM roles to service accounts (IRSA). This
way one can create an AWS role that has access to specific AWS resources and assign it to the 
service account. Technically, this means pods with service accounts attached call out to a
public OIDC discovery endpoint for AWS IAM when it starts. This endpoint signs a token issued
by k8s and the token gets exchanged for AWS IAM credentials. The credentials and AWS role ARN
are injected into the pods environment variables by EKS.

Practically, this means that you need to create an AWS role and assign it to the service account
using the following annotation
```sh
apiVersion: v1
kind: ServiceAccount
metadata:
  annotations:
    eks.amazonaws.com/role-arn: arn:aws:iam::<ACCOUNT_ID>:role/<IAM_ROLE_NAME>
```

In the context of the aws-node daemonset, we create an AWS role that uses the AWS EKS_CNI managed policy
and assign it to the aws-node service account in the aforementioned manner. We then enable IRSA on the aws-node
daemonset.

### Restrictive Pod Security Policy
By default, the EKS cluster has a privileged PSP. The `ConfigureClusterRolesAndPsp` created a restricted PSP
and applies it to all service accounts in the kube-system namespace as well as the automation group.
**TODO:** PSP are being depracated in favour of Pod Security Standards. As the feature matures we must
migrate over.

### Log aggregration and cluster metrics
We use fluentbit as a log aggregator to send logs to cloudwatch. Fluentbit is more performant than Fluentd and is 
recommended by AWS.

We also use cloudwatch agent to send cluster metrics to cloudwatch.

### Networking
The k8s network model is as follows
- Every pod gets its own IP address.
- Containers in a pod share the pod IP address and can communicate freely with eachother.
- Pods can communicate with all other pods in the cluster using IP addresses.
- Network policies are used to restrict traffic

The above essentially creates a flat network which is dangerous as an attacker just needs to be able to
breach the cluster network. As such we adopt the *Zero Trust Networking* approach where we assume the
network is always hostile. We apply a global deny all network policy and then explicitly open up communication
between our services. We codify these network policies and manage them using Calico (recommended by AWS for EKS).

Technically, Calico works at the host network level by setting up a network that uses BGP to route packets between hosts.
This offers performance gains over overlay networks.

Our default network policies are to globally deny all ingress and egress but allow CoreDns queries.

**Encyrption in Transit**
TODO: Decide on a service mesh to encrypt traffic between services. Istio is an option as it integrates with Calico
which would allow us to manage host and service mesh network policies in one place.
