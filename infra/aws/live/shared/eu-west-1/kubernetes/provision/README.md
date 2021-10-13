## Architecture
The control plane is deployed in an AWS managed VPC. We then deploy managed node groups in 
our own VPC.

## Deployment steps
The end state we want our EKS cluster in is to have the cluster endpoint only accessible from within
our VPC. However, pulumi needs access to the cluster to configure the `aws-auth` configmap and VPC
CNI add on. We therefore need to:

- 1. provision the cluster with public access to the cluster endpoint.
This is done by setting the `ALLOW_PUBLIC_CONFIGURATION=true` environment variable and running `pulumi up`.
- 2. configure the cluster so that it meets EKS best practices. 
This is done by running the `configure` project.
- 3. close off public access to the EKS cluster. To disable public access, make sure the `ALLOW_PUBLIC_CONFIGURATION`
environment variable is not set. `unset ALLOW_PUBLIC_CONFIGURATION`. Then run `pulumi up`.
