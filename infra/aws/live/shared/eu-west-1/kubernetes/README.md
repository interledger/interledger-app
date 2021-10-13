There are two steps to deploying the k8s cluster: provisioning and configuration. The reason
these are split into two is because our cluster is provisioned with private access only to the
cluster endpoint. This requires the use of Boundary to complete the configuration.

In light of this, the `provision` project needs to be run first followed by the `configure`
project. See the respective READMEs for more information.

## Best practice considerations
AWS has published best practices for EKS https://aws.github.io/aws-eks-best-practices/security/docs/iam/#kubernetes-service-accounts.
It is encouraged to read through them as they contain crucial information for secure deployments
of applications. Important considerations are highlighted below.

## Kubernetes Service Accounts

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