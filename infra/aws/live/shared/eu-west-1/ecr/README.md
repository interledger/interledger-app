## AWS ECR

This provisions the Fynbos container registry. It's important to ensure when provisioning this you ensure you
have installed the (Amazon ECR Docker Credential Helper)[https://github.com/awslabs/amazon-ecr-credential-helper]. Make 
sure to also follow the (configuration)[https://github.com/awslabs/amazon-ecr-credential-helper#configuration] of your
Docker client.

### Run

```shell
aws-vault exec {SHARED_ACC_PROFILE} -- pulumi up -s fynbos/main 
```