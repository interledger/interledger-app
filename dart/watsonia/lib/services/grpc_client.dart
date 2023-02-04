import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:grpc/grpc.dart';
import 'package:watsonia/generated/protobuf-dart/backend/v1/backend.pbgrpc.dart';

class GrpcClient {
  GrpcClient._();

  static final ClientChannel _channel = ClientChannel(
    dotenv.env['BACKEND_URL'] as String,
    port: int.parse(dotenv.env['BACKEND_PORT'] as String),
    options: const ChannelOptions(
      credentials: ChannelCredentials.insecure(),
    ),
  );

  String _xSessionToken = '';

  late BackendServiceClient _backend = BackendServiceClient(
    _channel,
    interceptors: [AuthInterceptor(_xSessionToken)],
  );

  // Use a private constructor to ensure that we only have one instance.
  static final GrpcClient _instance = GrpcClient._();

  static BackendServiceClient get backend => _instance._backend;

  static void updateAuthToken(String token) {
    _instance._xSessionToken = token;
    _instance._backend = BackendServiceClient(
      _channel,
      interceptors: [AuthInterceptor(token)],
    );
  }
}

class AuthInterceptor implements ClientInterceptor {
  final String authToken;

  AuthInterceptor(this.authToken);

  // NOTE: There is a race condition here where the token could not be initialised yet.

  @override
  ResponseFuture<R> interceptUnary<Q, R>(
      ClientMethod<Q, R> method, Q request, CallOptions options, invoker) {
    var newOptions = options.mergedWith(
      CallOptions(
        metadata: <String, String>{
          'token': authToken,
        },
      ),
    );

    return invoker(method, request, newOptions);
  }

  @override
  ResponseStream<R> interceptStreaming<Q, R>(
      ClientMethod<Q, R> method,
      Stream<Q> requests,
      CallOptions options,
      ClientStreamingInvoker<Q, R> invoker) {
    var newOptions = options.mergedWith(
      CallOptions(
        metadata: <String, String>{
          'token': authToken,
        },
      ),
    );

    return invoker(method, requests, newOptions);
  }
}
