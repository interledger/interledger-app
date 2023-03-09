# watsonia

The etter way to pay

## Getting Started

This project is a starting point for a Flutter application.

A few resources to get you started if this is your first Flutter project:

- [Lab: Write your first Flutter app](https://docs.flutter.dev/get-started/codelab)
- [Cookbook: Useful Flutter samples](https://docs.flutter.dev/cookbook)

For help getting started with Flutter development, view the
[online documentation](https://docs.flutter.dev/), which offers tutorials,
samples, guidance on mobile development, and a full API reference.

# Commands to get endpoints working

```shell
# Tunnel to access protea and kratos
cloudflared tunnel --http-host-header fynbos.test --url http://localhost:80/
# Port forward protea (currently doesn't work because tls)
# kubectl port-forward --namespace protea deployment/protea 3030:3000
# Port forward the backend grpc
kubectl port-forward --namespace backend deployment/backend 8443:8443
```

# Testing deep linking

https://docs.flutter.dev/development/ui/navigation/deep-linking

```shell
adb shell 'am start -a android.intent.action.VIEW \
    -c android.intent.category.BROWSABLE \
    -d "http://fynbos.app/transactions"' \
    dev.fynbos.watsonia

```