# Interledger Wallet

## Local Development Environment

We provide a `docker compose` managed environment that will attempt to automatically rebuild the sources upon any change. 

The development environment will serve the various services under the following URLs.

| URL                              | Description                                 |
|-----------------------------------|---------------------------------------------|
| http://interledger.test                  | Wallet frontend (previously called Protea)  |
| http://admin.mgnt.interledger.test       | Wallet admin portal (previously Botanist)   |
| http://temporal.mgnt.interledger.test    | Temporal admin portal                       |
| http://local.fynbos.me                   | Backend endpoint                            |
| http://local.ilp.link                    | Wallet address                              |

Prerequisites:
- Docker locally installed

Start environment
1. Add host entries
Edit your `/etc/hosts` file with an appropriate text editor and add the following line
```
127.0.0.1 interledger.test admin.mgnt.interledger.test temporal.mgnt.interledger.test local.fynbos.me  local.ilp.link rafiki.mgnt.interledger.test auth.interledger.test
```
All the domains used for local development will now point to your local host from where it will be served.

2. Start all services using Docker Compose
```sh
# From repository folder

cd ./local
docker compose up -d
```

All services should start asyncronously and automatically rebuild sources whenever code changes have been made.

## Troubleshooting

### View log for a service
The example below demonstrates how to view the last 100 logs for the protea service.
```sh
docker compose logs -f --tail=100 protea
```

### Recreating environment
Removes the whole environment to start fresh
```sh
# Remove containers and volumes
docker compose down -v
```

## Legacy Development environment

It is also possible to start up the project using Nomad and Vagrant VMs. See [this](./legacy-dev.md) document for more information.
