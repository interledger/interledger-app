## Autoscaling Gitlab Runners

This project deploys our Gitlab Runner infrastructure into the shared account.
This [guide](https://docs.gitlab.com/runner/configuration/runner_autoscale_aws/)
was followed mostly to get it up and running.

## Deploying

It's important to note that when deploying this project you need to set the GITLAB_TOKEN

To get the gitlab token a curl request needs to be made to

```shell
curl --request POST "https://gitlab.com/api/v4/runners" --form "token="
```

Where the token in the above command can be gotten from https://gitlab.com/fynbos/fynbos/-/settings/ci_cd under the
`Runners` section. Its called the `registration token`

Once you have the token from curl add it to the follow

```shell
GITLAB_TOKEN={{REPLACE_TOKEN}} aws-vault exec {{YOUR_PROFILE_FOR_SHARED}} -- pulumi up
```

## Configuration

Configuration of the GL Runner Manager is done through the `cloudinit.yaml` file. This is templated through Go's
templating to in place required variables.

## Architecture

The two main components of this setup are the `Gitlab Runner Manager` and the `Gitlab Runner Machine`. Both these
components are deployed into the `shared` AWS account in the `private` subnet.

### Gitlab Runner Manager

The manager is responsible for listening for jobs from Gitlab and spinning up new EC2 instances as required to run the
jobs. It also shuts down these EC2 instances if they are not being used anymore.

#### Role & Permissions

The roles and permissions for the Manager are defined in `role.go`. This role needs the ability for ec2 to assume it as
well as add a policy that allows it to manage `EC2` resources. This is so that it can spin up and down instances as
required.

#### Security Groups

Inbound

| Source | Protocol | Port Range | Comments |
|--------|-------------|----------|----------|
|       |           |         |         |

Outbound

| Destination | Protocol | Port Range | Comments |
|--------|-------------|----------|----------|
|   0.0.0.0/0         |    TCP       |      443   | Allow communication to Github.org to pull jobs. Also used for cloudint.yaml       |
|   Private Subnet    |    TCP       |      2376  | Talk over docker socket to the machines       |
|   Private Subnet    |    TCP       |      22    | SSH into machines it has provisioned       |


### Gitlab Runner Machine

The `Gitlab Runner Machine` is the actual machine where jobs are run on. This instance is spun up and down by
the `Manager`
according to the config setup in `cloudint.yaml`

#### Security Groups

Inbound

| Source | Protocol | Port Range | Comments |
|--------|-------------|----------|----------|
|   Private Subnet    |    TCP       |      2376  | Allow manager to talk to docker socket       |
|   Private Subnet    |    TCP       |      22    | Allow manager to SSH into machine      |

Outbound

| Destination | Protocol | Port Range | Comments |
|--------|-------------|----------|----------|
|   0.0.0.0/0         |    ALL       |      ALL   | There is some protocol this machine is using that I cant identify yet. Needs to be tightened up, but it wont work without this       |

## Gotchas

### Docker-machine not maintained

Docker-machine is no longer actively maintained by docker. Gitlab are maintaining an
[active fork](https://gitlab.com/gitlab-org/ci-cd/docker-machine) for critical bug fixes. This version is used in
`cloudinit.yaml` when pulling the docker-machine binary onto the machine.

### Docker version issues

There appears to be a [bug](https://github.com/docker/machine/issues/4858) with the latest docker version installed by
docker-machine. As a workaround we specify an older version in the config in `MachineOptions` with
`engine-install-url=https://releases.rancher.com/install-docker/19.03.15.sh`
