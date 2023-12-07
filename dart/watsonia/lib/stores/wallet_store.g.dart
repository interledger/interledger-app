// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'wallet_store.dart';

// **************************************************************************
// StoreGenerator
// **************************************************************************

// ignore_for_file: non_constant_identifier_names, unnecessary_brace_in_string_interps, unnecessary_lambdas, prefer_expression_function_bodies, lines_longer_than_80_chars, avoid_as, avoid_annotating_with_dynamic, no_leading_underscores_for_local_identifiers

mixin _$WalletStore on _WalletStore, Store {
  late final _$publicNameAtom =
      Atom(name: '_WalletStore.publicName', context: context);

  @override
  String get publicName {
    _$publicNameAtom.reportRead();
    return super.publicName;
  }

  @override
  set publicName(String value) {
    _$publicNameAtom.reportWrite(value, super.publicName, () {
      super.publicName = value;
    });
  }

  late final _$setPublicNameAsyncAction =
      AsyncAction('_WalletStore.setPublicName', context: context);

  @override
  Future<void> setPublicName(String name) {
    return _$setPublicNameAsyncAction.run(() => super.setPublicName(name));
  }

  late final _$initAsyncAction =
      AsyncAction('_WalletStore.init', context: context);

  @override
  Future<void> init() {
    return _$initAsyncAction.run(() => super.init());
  }

  @override
  String toString() {
    return '''
publicName: ${publicName}
    ''';
  }
}
