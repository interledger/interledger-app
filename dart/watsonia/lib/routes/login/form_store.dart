import 'package:mobx/mobx.dart';
import 'package:watsonia/stores/auth_store.dart';

part 'form_store.g.dart';

// ignore: library_private_types_in_public_api
class FormStore = _FormStore with _$FormStore;

abstract class _FormStore with Store {
  // TODO Need to pass auth store in so that we can access it from here
  final FormErrorState error = FormErrorState();

  @observable
  String email = '';

  @observable
  String password = '';

  @computed
  bool get canLogin => !error.hasErrors;

  @action
  void setError(String email, String password) {
    error.email = email.isEmpty ? email : null;
    error.password = password.isEmpty ? password : null;
  }

  void submit(AuthStore store) {
    // TODO: implement submit
    store.login(email, password);
    // Should send a login request via the AuthStore
    // If the request returns an error we should show the error
    // otherwise route to /

    // setError(email, password);
  }
}

// ignore: library_private_types_in_public_api
class FormErrorState = _FormErrorState with _$FormErrorState;

abstract class _FormErrorState with Store {
  @observable
  String? email;

  @observable
  String? password;

  @computed
  bool get hasErrors => email != null || password != null;
}
