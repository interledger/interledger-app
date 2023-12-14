# watsonia

## Getting Started

This project is a starting point for a Flutter app.

A few resources to get you started if this is your first Flutter project:

For help getting started with Flutter development, view the
[online documentation](https://docs.flutter.dev/).

## Commands to get endpoints working

We can probably integrate these into tilt at a later stage.

```shell
# Port forward the kratos - public
k9s # shift-f on kratos-0 and make sure both ports are 4433
# Port forward the backend - grpc
kubectl port-forward --namespace backend deployment/backend 8443:8443

# To generate the mobx stores run
flutter pub run build_runner watch --delete-conflicting-outputs
```

## Testing deep linking

https://docs.flutter.dev/development/ui/navigation/deep-linking

```shell
adb shell 'am start -a android.intent.action.VIEW \
    -c android.intent.category.BROWSABLE \
    -d "http://fynbos.app/transactions"' \
    dev.fynbos.watsonia

```

# Docs

## Auth

- Session token and wallet ID stored in secure storage.

## Data

Mobx

## Routing

Push goes nested, go is for replacing root pages
