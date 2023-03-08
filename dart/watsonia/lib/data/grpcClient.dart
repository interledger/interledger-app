import 'package:grpc/grpc.dart';
import 'package:watsonia/generated/protobuf-dart/backend/v1/backend.pbgrpc.dart';
// https://github.com/grpc/grpc-dart/blob/master/example/route_guide/lib/src/client.dart
Future<void> main(List<String> args) async {
  final channel = ClientChannel(
    'localhost',
    port: 50051,
    options: ChannelOptions(
      credentials: ChannelCredentials.insecure(),
      codecRegistry:
      CodecRegistry(codecs: const [GzipCodec(), IdentityCodec()]),
    ),
  );
  final stub = BackendServiceClient(channel);

  final name = args.isNotEmpty ? args[0] : 'world';

  try {
    final response = await stub.canSignup(
      CanSignupRequest()..id = name,
      options: CallOptions(compression: const GzipCodec()),
    );
    print('Greeter client received: ${response.canSignup}');
  } catch (e) {
    print('Caught error: $e');
  }
  await channel.shutdown();
}

class GrpcClient {
  static final channel = ClientChannel(
    'rhode-myrtle-desktops-sl.trycloudflare.com',
    port: 8443,
    options: ChannelOptions(
      credentials: ChannelCredentials.insecure(),
      codecRegistry:
      CodecRegistry(codecs: const [GzipCodec(), IdentityCodec()]),
    ),
  );
  final backend = BackendServiceClient(channel);
  final openPayments = OpenPaymentServiceClient(channel);

  @override
  
}