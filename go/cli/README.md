# Fynbos CLI

The command-line interface to make payments from your Fynbos wallet.

This cli makes use of Fynbos' [Open-Payments](https://docs.openpayments.guide/) API to access your
wallet. See how to grant it access [here](#granting-the-cli-acess-to-your-Fynbos-wallet).

## Installation

**Building from source**
1. Install [golang](https://go.dev/doc/install) (at least v.1.18)
2. Clone the repo
```sh
git clone https://gitlab.com/fynbos/cli (TODO: public access)
```
3. Build
```sh
cd cli && go build -o fynbos
```

## Granting the cli access to your Fynbos wallet
**Make the cli your own**
```sh
fynbos config set wallet <your payment-pointer>
```

**Generate a key pair**

```sh
fynbos keys create
```
Provide an alias for the key when prompted. Take note of the public key it has generated.

**Grant your cli access to your wallet**

Navigate to https://fynbos.app/clients and paste your public key. From here, you will be able to
specify limits your cli is able to spend.

**Make a payment**

You're all set. Try
```sh
fynbos pay https://ilp.link/adrian
```
and follow the prompts to make your first payment.
