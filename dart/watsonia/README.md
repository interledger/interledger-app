# watsonia

The better way to pay

## Getting Started

Figma
sitemap - https://www.figma.com/file/vDjoPZIZhpAQ0WPBEwmvzi/Sitemap?type=whiteboard&node-id=0-1&t=zros6SSJx9HMGpqZ-0

This project is a starting point for a Flutter application.

A few resources to get you started if this is your first Flutter project:

- [Lab: Write your first Flutter app](https://docs.flutter.dev/get-started/codelab)
- [Cookbook: Useful Flutter samples](https://docs.flutter.dev/cookbook)

For help getting started with Flutter development, view the
[online documentation](https://docs.flutter.dev/), which offers tutorials,
samples, guidance on mobile development, and a full API reference.

# Commands to get endpoints working

```shell
# Port forward the kratos - public
k9s # shift-f on kratos-0 and make sure both ports are 4433
# Port forward the backend - grpc
kubectl port-forward --namespace backend deployment/backend 8443:8443

# To generate the mobx stores run
flutter pub run build_runner watch --delete-conflicting-outputs
```

# Testing deep linking

https://docs.flutter.dev/development/ui/navigation/deep-linking

```shell
adb shell 'am start -a android.intent.action.VIEW \
    -c android.intent.category.BROWSABLE \
    -d "http://fynbos.app/transactions"' \
    dev.fynbos.watsonia

```

# TODO:

- [ ] Finish kratos implementation
- [ ] Start implementing grpc clients to pull data.
- [ ] Local secure storage for auth credentials once user logged in
- [ ] Card component should just expose a Column. Cards should be placed in a ListView.
- [ ] Pusher integration - https://pusher.com/docs/channels/getting_started/flutter/
- [ ] Proper splash
- [ ] https://stackoverflow.com/questions/54464853/flutter-loading-an-iframe-from-webview
- [ ] Local auth biometrics - https://pub.dev/packages/local_auth

## Auth

- Session token and wallet ID stored in secure storage.

## Stores

- KYC
- Accounts
- Identities
- Transactions
- Features

### Copy to clipboard

https://stackoverflow.com/questions/55885433/flutter-dart-how-to-add-copy-to-clipboard-on-tap-to-a-app

### Share plugin

https://pub.dev/packages/share_plus/versions