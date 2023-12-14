// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'inquiry_store.dart';

// **************************************************************************
// StoreGenerator
// **************************************************************************

// ignore_for_file: non_constant_identifier_names, unnecessary_brace_in_string_interps, unnecessary_lambdas, prefer_expression_function_bodies, lines_longer_than_80_chars, avoid_as, avoid_annotating_with_dynamic, no_leading_underscores_for_local_identifiers

mixin _$InquiryStore on _InquiryStore, Store {
  late final _$isInitializingAtom =
      Atom(name: '_InquiryStore.isInitializing', context: context);

  @override
  bool get isInitializing {
    _$isInitializingAtom.reportRead();
    return super.isInitializing;
  }

  @override
  set isInitializing(bool value) {
    _$isInitializingAtom.reportWrite(value, super.isInitializing, () {
      super.isInitializing = value;
    });
  }

  late final _$isLoadingAtom =
      Atom(name: '_InquiryStore.isLoading', context: context);

  @override
  bool get isLoading {
    _$isLoadingAtom.reportRead();
    return super.isLoading;
  }

  @override
  set isLoading(bool value) {
    _$isLoadingAtom.reportWrite(value, super.isLoading, () {
      super.isLoading = value;
    });
  }

  late final _$idempotencyKeyAtom =
      Atom(name: '_InquiryStore.idempotencyKey', context: context);

  @override
  String get idempotencyKey {
    _$idempotencyKeyAtom.reportRead();
    return super.idempotencyKey;
  }

  @override
  set idempotencyKey(String value) {
    _$idempotencyKeyAtom.reportWrite(value, super.idempotencyKey, () {
      super.idempotencyKey = value;
    });
  }

  late final _$canceledAtom =
      Atom(name: '_InquiryStore.canceled', context: context);

  @override
  InquiryCanceled? get canceled {
    _$canceledAtom.reportRead();
    return super.canceled;
  }

  @override
  set canceled(InquiryCanceled? value) {
    _$canceledAtom.reportWrite(value, super.canceled, () {
      super.canceled = value;
    });
  }

  late final _$errorAtom = Atom(name: '_InquiryStore.error', context: context);

  @override
  InquiryError? get error {
    _$errorAtom.reportRead();
    return super.error;
  }

  @override
  set error(InquiryError? value) {
    _$errorAtom.reportWrite(value, super.error, () {
      super.error = value;
    });
  }

  late final _$completedAtom =
      Atom(name: '_InquiryStore.completed', context: context);

  @override
  InquiryComplete? get completed {
    _$completedAtom.reportRead();
    return super.completed;
  }

  @override
  set completed(InquiryComplete? value) {
    _$completedAtom.reportWrite(value, super.completed, () {
      super.completed = value;
    });
  }

  late final _$initializeInquiryAsyncAction =
      AsyncAction('_InquiryStore.initializeInquiry', context: context);

  @override
  Future<void> initializeInquiry() {
    return _$initializeInquiryAsyncAction.run(() => super.initializeInquiry());
  }

  late final _$startInquiryAsyncAction =
      AsyncAction('_InquiryStore.startInquiry', context: context);

  @override
  Future<void> startInquiry() {
    return _$startInquiryAsyncAction.run(() => super.startInquiry());
  }

  late final _$_InquiryStoreActionController =
      ActionController(name: '_InquiryStore', context: context);

  @override
  void dispose() {
    final _$actionInfo = _$_InquiryStoreActionController.startAction(
        name: '_InquiryStore.dispose');
    try {
      return super.dispose();
    } finally {
      _$_InquiryStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  String toString() {
    return '''
isInitializing: ${isInitializing},
isLoading: ${isLoading},
idempotencyKey: ${idempotencyKey},
canceled: ${canceled},
error: ${error},
completed: ${completed}
    ''';
  }
}
