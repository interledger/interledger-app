# Local development environment

## Prerequisites
* [Docker](https://www.docker.com/) installed and running

## Creating the environment

We provide a `docker compose` managed environment that will attempt to automatically rebuild the sources upon any change.

Service configuration can be overridden through a `.env` file in the `local/` directory. Start by copying [local/example.env](./example.env) to `.env` and editing only the values you want to change.

Environment variable names are prefixed by service to avoid collisions:
- `BACKEND_*` configures the wallet backend and shared mock credentials used by `mockgatehub`, `mockxago`, and `mockpti`
- `PROTEA_*` configures the Protea frontend

```sh
cp example.env .env
```

The environment creation is backed up by a ```Makefile``` in the ```local``` directory. 

**This is the recommended way to use it.** However, you can use plain ```docker compose```.

> **Note:** Make sure that your current working directory ```local``` before starting.

> **Note:** `docker compose` automatically reads `.env` from this directory, so you do not need to pass `--env-file` for the normal local workflow.

### Getting familiar with ```make``` interface
Running ```make``` or ```make help``` will output some available commands:
```sh
make <command>
e.g. make infrastructure

help                    Show this help message
hosts                   Add required entries to /etc/hosts (requires sudo)

infrastructure|infra    Traefik, Postgres, Redis
services|svc            Kratos, Temporal, Rafiki, Ngrok (requires infra)
application|app         Wallet application, backends + frontends (requires infra, svc)
backend|back            Just backends services
frontend|front          Just frontends services
all                     All services (infra, svc, app)

build                   Build all images
pull                    Pull all images

down                    Stop all running services
delete-volumes|reset    Stop all services and remove associated volumes

certs                   Generate self-signed TLS certificates for local development
print-certs             Print details of the self-signed TLS certificates

trust                   Trust the self-signed TLS certificates on macOS
print-trust             List all self-signed TLS certificates on macOS
untrust                 Remove the self-signed TLS certificates from macOS
```

### Step by step
The ```help``` in the ```Makefile``` should be enough to provide the support in building the envrionment, but let's break down the steps.

1. (Optional, recommended) Add the host records in ```/etc/hosts``` file. This is only needed if you did not add the records by now in a manual fashion. Even better, you can remove the manual additions and allow the command to manage them for you. This is idempotent.
    ```sh
    make hosts
    ```

2. Create the self-signed certificates. This will create a certificate for ```interledger.test``` with some required ```SANs```. You can look at the entire list of alternative names in ```local/config/san.cnf``` under the ```alt_names``` section.
    ```sh
    make certs
    ```

3. (Optional) Print the generated certificates for sanity.
    ```sh
    make print-certs
    ```

4. (Optional, recommended) Add the self-signed certificates to macOS trust store. This will ensure that the certificates are considered valid by your browser.
    > **Note:** If you are using Firefox, you need to add them manualy into it. Firefox has its trust store and does not follow the system-wide one.
    ```sh
    make trust
    ```

5. (Optional) Print the trust store on macOS just as a check that the certs were added in the trust store
    ```sh
    make print-trust
    ```
    If needed, you can remove the certificates from trust store
    ```sh
    make untrust
    ```
    > **Note:** Always check the certificate removal

6. Start the infrastructure services - Traefik, Postgresql, Redis
    ```sh
    make infra
    ````
    Open the browser and go to ```http://traefik.test```. First, you should be redirected to the ```https://traefik.test```. If all went well, then you should not have any security-related pop-ups and the certificate is valid. You should check the certificate also and look for ```interledger.test``` in the details.

7. Start the required third-party services - Temporal, Kratos, Rafiki(s)
    ```sh
    make svc
    ```

8. Start the wallet services - Backend, Protea, Botanist
    ```sh
    make app
    ```

> **TIP:** Start ```infra``` first, ```svc``` next followed by ```app```, use 3 terminal tabs or ```tmux``` or ```screen```.

### Building and pulling images
To build all images without starting any services:
```sh
make build
```

To pull all images without starting any services:
```sh
make pull
```

### Docker compose interface
If you prefer using plain docker compose here's some tips.

Consider managing ```/etc/hosts``` and ```certs``` via ```make``` it's not managed by compose.

```sh
# start all services backgroud
docker compose --profile "*" up -d

# start all services foreground
docker compose --profile "*" up

# start all services foreground and live reload
docker compose --profile "*" up --watch

# start in bacground and live reload
# not a supported feature in docker

# available profiles
# - infrastructure (traefik, postgres, redis)
# - services (kratos, temporal, rafiki) # note: requires infrastructure
# - application (backend+frontend) # note: requires infrastrucure and services
# - backend
# - frontend

# e.g infrastructure
docker compose --profile infrastructure up

# multiple profiles
docker compose --profile infrastructure --profile services up

# View last 100 lines of logs for protea service
docker compose logs -f --tail=100 protea

# Remove containers and volumes
docker compose down -v
```

> **Note:** Remember to start profiles in order.

### URLs

| URL                                       | Description                                 |
|-------------------------------------------|---------------------------------------------|
| https://interledger.test                  | Wallet frontend (previously called Protea)  |
| https://admin.mgnt.interledger.test       | Wallet admin portal (previously Botanist)   |
| https://temporal.mgnt.interledger.test    | Temporal admin portal                       |
| https://local.ilp.link                    | Wallet address                              |
| https://traefik.test                      | Traefik dashboard                           |
| https://rafiki.mgnt.interledger.test      | Rafiki                                      |
| https://ngrok.test                        | Ngrok ui                                    |
| https://mockgatehub.interledger.test      | MockGateHub API (local GateHub replacement) |
| https://mockxago.interledger.test         | MockXago API (local Xago replacement)       |
| https://mockpti.interledger.test          | MockPTI API and SDK stub                    |


### TODO
- [ ] Add instructions for Linux users regarding trusting the self-signed certificates.
