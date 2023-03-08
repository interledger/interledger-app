import 'package:grpc/grpc.dart';
import 'package:watsonia/generated/protobuf-dart/backend/v1/backend.pbgrpc.dart';

class GrpcClient {
  GrpcClient._();

  static final ClientChannel _channel = ClientChannel(
    '10.0.2.2',
    port: 8443,
    options: const ChannelOptions(
      credentials: ChannelCredentials.insecure(),
    ),
  );

  final _backend = BackendServiceClient(_channel);
  final _openPayments = OpenPaymentServiceClient(_channel);

  // Use a private constructor to ensure that we only have one instance.
  static final GrpcClient _instance = GrpcClient._();

  static BackendServiceClient get backend => _instance._backend;

  static OpenPaymentServiceClient get openPayments => _instance._openPayments;
}
