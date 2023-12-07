// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'auth_store.dart';

// **************************************************************************
// StoreGenerator
// **************************************************************************

// ignore_for_file: non_constant_identifier_names, unnecessary_brace_in_string_interps, unnecessary_lambdas, prefer_expression_function_bodies, lines_longer_than_80_chars, avoid_as, avoid_annotating_with_dynamic, no_leading_underscores_for_local_identifiers

mixin _$AuthStore on _AuthStore, Store {
  late final _$flowIdAtom = Atom(name: '_AuthStore.flowId', context: context);

  @override
  String get flowId {
    _$flowIdAtom.reportRead();
    return super.flowId;
  }

  @override
  set flowId(String value) {
    _$flowIdAtom.reportWrite(value, super.flowId, () {
      super.flowId = value;
    });
  }

  late final _$sessionTokenAtom =
      Atom(name: '_AuthStore.sessionToken', context: context);

  @override
  String get sessionToken {
    _$sessionTokenAtom.reportRead();
    return super.sessionToken;
  }

  @override
  set sessionToken(String value) {
    _$sessionTokenAtom.reportWrite(value, super.sessionToken, () {
      super.sessionToken = value;
    });
  }

  late final _$statusAtom = Atom(name: '_AuthStore.status', context: context);

  @override
  AuthStatus get status {
    _$statusAtom.reportRead();
    return super.status;
  }

  @override
  set status(AuthStatus value) {
    _$statusAtom.reportWrite(value, super.status, () {
      super.status = value;
    });
  }

  late final _$initAsyncAction =
      AsyncAction('_AuthStore.init', context: context);

  @override
  Future<void> init() {
    return _$initAsyncAction.run(() => super.init());
  }

  late final _$initLoginFlowAsyncAction =
      AsyncAction('_AuthStore.initLoginFlow', context: context);

  @override
  Future<void> initLoginFlow() {
    return _$initLoginFlowAsyncAction.run(() => super.initLoginFlow());
  }

  late final _$loginAsyncAction =
      AsyncAction('_AuthStore.login', context: context);

  @override
  Future<void> login(String email, String password) {
    return _$loginAsyncAction.run(() => super.login(email, password));
  }

  late final _$_AuthStoreActionController =
      ActionController(name: '_AuthStore', context: context);

  @override
  void refreshRouter(AuthStatus status) {
    final _$actionInfo = _$_AuthStoreActionController.startAction(
        name: '_AuthStore.refreshRouter');
    try {
      return super.refreshRouter(status);
    } finally {
      _$_AuthStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  String toString() {
    return '''
flowId: ${flowId},
sessionToken: ${sessionToken},
status: ${status}
    ''';
  }
}
