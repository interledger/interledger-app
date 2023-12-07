import 'package:mobx/mobx.dart';
import 'package:watsonia/components.dart';
import 'package:watsonia/main.dart';
import 'package:watsonia/router.dart';
import 'package:watsonia/stores/wallet_store.dart';

part 'form_store.g.dart';

// ignore: library_private_types_in_public_api
class FormStore = _FormStore with _$FormStore;

abstract class _FormStore with Store {
  final FormErrorState error = FormErrorState();

  @observable
  String name = '';

  @action
  void setError(String name) {
    error.name = name.isEmpty ? name : null;
  }

  Future<void> submit(WalletStore store) async {
    try {
      await store.setPublicName(name);
      const snackBar = SnackBar(
        showCloseIcon: true,
        behavior: SnackBarBehavior.floating,
        margin: EdgeInsets.all(16),
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.all(Radius.circular(12)),
        ),
        content: Text('Your public name was updated.'),
      );

      snackbarKey.currentState?.showSnackBar(snackBar);
      rootNavigatorKey.currentState!.pop();
    } catch (e) {
      print('Error setting wallet name: $e');
      setError('There was an error setting the name.');
    }
  }
}

// ignore: library_private_types_in_public_api
class FormErrorState = _FormErrorState with _$FormErrorState;

abstract class _FormErrorState with Store {
  @observable
  String? name;

  @computed
  bool get hasErrors => name != null;
}
