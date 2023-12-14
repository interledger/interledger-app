import 'package:mobx/mobx.dart';
import 'package:watsonia/router.dart';
import 'package:watsonia/services/auth_service.dart';
import 'package:watsonia/services/grpc_client.dart';
import 'package:watsonia/storage.dart';

part 'auth_store.g.dart';

// ignore: library_private_types_in_public_api
class AuthStore = _AuthStore with _$AuthStore;

abstract class _AuthStore with Store {
  final SecureStorage secureStorage;
  final AuthService authService;

  _AuthStore(this.authService, this.secureStorage)
      : status = AuthStatus.uninitialized;

  @observable
  String flowId = '';

  @observable
  String sessionToken = '';

  @observable
  AuthStatus status;

  @action
  Future<void> init() async {
    bool hasToken = await secureStorage.hasToken();

    if (hasToken) {
      // For now we'll assume if you have a token you're authed.
      // The backend should handle bad tokens gracefully.
      status = AuthStatus.authenticated;
      String? token = await secureStorage.getToken();

      sessionToken = token!;
      GrpcClient.updateAuthToken(token);
      appRouter.refresh();
    } else {
      status = AuthStatus.uninitialized;
    }
  }

  @action
  Future<void> initLoginFlow() async {
    flowId = await authService.initiateLoginFlow();
  }

  @action
  Future<void> login(String email, String password) async {
    final sessionToken = await authService.signIn(flowId, email, password);
    // TODO figure out error state
    secureStorage.persistToken(sessionToken);
    GrpcClient.updateAuthToken(sessionToken);
    this.sessionToken = sessionToken;

    status = AuthStatus.authenticated;
    appRouter.refresh();
  }
// TODO Setup a reaction on status to refresh the router
}
