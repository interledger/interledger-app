import 'package:mobx/mobx.dart';

import '../generated/protobuf-dart/backend/v1/backend.pbgrpc.dart';
import '../services/grpc_client.dart';

part 'kyc_store.g.dart';

// ignore: library_private_types_in_public_api
class KYCStore = _KYCStore with _$KYCStore;

abstract class _KYCStore with Store {
  BackendServiceClient backend = GrpcClient.backend;

  @observable
  late int status = 0;

  @action
  Future<void> init() async {
    final response = await backend.kYCStatus(
      Empty(),
    );
    status = response.kycStatus;
  }
}
