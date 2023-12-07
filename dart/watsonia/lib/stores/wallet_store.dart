import 'package:mobx/mobx.dart';

import '../generated/protobuf-dart/backend/v1/backend.pbgrpc.dart';
import '../services/grpc_client.dart';

part 'wallet_store.g.dart';

// ignore: library_private_types_in_public_api
class WalletStore = _WalletStore with _$WalletStore;

abstract class _WalletStore with Store {
  BackendServiceClient backend = GrpcClient.backend;

  @observable
  late String publicName = '';

  @observable
  late String walletAddress = '';

  @action
  Future<void> setPublicName(String name) async {
    await backend.setWalletName(SetWalletNameRequest(name: name));
    // TODO handle error
    publicName = name;
  }

  @action
  Future<void> init() async {
    final walletInfo = await backend.getWalletInfo(
      Empty(),
    );

    final publicWalletInfo = await backend.getPublicWalletInfo(
      GetPublicWalletInfoRequest(walletAddress: walletInfo.url),
    );
    publicName = publicWalletInfo.publicName;
    walletAddress = walletInfo.formattedURL;
  }
}
