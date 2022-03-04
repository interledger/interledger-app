## AWS ECR

This provisions the Fynbos container registry. It's important to ensure when provisioning this you ensure you
have installed the (Amazon ECR Docker Credential Helper)[https://github.com/awslabs/amazon-ecr-credential-helper]. Make 
sure to also follow the (configuration)[https://github.com/awslabs/amazon-ecr-credential-helper#configuration] of your
Docker client.

### Run

```shell
aws-vault exec {SHARED_ACC_PROFILE} -- pulumi up -s fynbos/main 
```

## Syncing Dependencies

We require some images that we host our selves by mirror them. We use a tool called [skopeo](https://github.com/containers/skopeo) to facilitate this. For 
now, this process is manual. Edit the `sync.yml` file to ensure which images are being tracked and then run the following
command locally to sync it.

First login to skopeo

```shell
make login account={SHARED_ACC_PROFILE}
```

And then run the following to do the sync. 

```shell
make sync
```