import 'package:flutter/material.dart';
import 'package:watsonia/data/grpcClient.dart';

import '../generated/protobuf-dart/backend/v1/backend.pbgrpc.dart';

class Wallet extends ChangeNotifier {
  BackendServiceClient backend = GrpcClient.backend;

  String _paymentPointer = '';

  /// The user's formatted payment pointer
  String get paymentPointer => _paymentPointer;

  Future<void> getPaymentPointer() async {
    // await Future<void>.delayed(const Duration(milliseconds: 200));
    try {
      print('Initiate grpc call');
      final response = await backend.getPublicWalletDetails(
        GetPublicWalletDetailsRequest()..id = "a3a6989b-732d-497d-930d-192b3982f35e",
        // options: CallOptions(compression: const GzipCodec()),
      );
      print('Greeter client received: ${response.publicName}');
    } catch (e) {
      print('Caught error: $e');
    }
    // Sign out.
    _paymentPointer = "\$fynbos.me/cairin";
    notifyListeners();
  }
}
