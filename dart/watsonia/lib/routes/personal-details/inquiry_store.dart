import 'dart:async';

import 'package:mobx/mobx.dart';
import 'package:persona_flutter/persona_flutter.dart';
import 'package:uuid/uuid.dart';
import 'package:watsonia/generated/protobuf-dart/backend/v1/backend.pbgrpc.dart';
import 'package:watsonia/router.dart';
import 'package:watsonia/services/grpc_client.dart';

part 'inquiry_store.g.dart';

// ignore: library_private_types_in_public_api
class InquiryStore = _InquiryStore with _$InquiryStore;

abstract class _InquiryStore with Store {
  BackendServiceClient backend = GrpcClient.backend;
  late StreamSubscription<InquiryCanceled> _canceledSubscription;
  late StreamSubscription<InquiryError> _errorSubscription;
  late StreamSubscription<InquiryComplete> _completedSubscription;

  @observable
  bool isInitializing = true;

  @observable
  bool isLoading = false;

  @observable
  String idempotencyKey = '';

  @observable
  InquiryCanceled? canceled;

  @observable
  InquiryError? error;

  @observable
  InquiryComplete? completed;

  @action
  Future<void> initializeInquiry() async {
    try {
      if (idempotencyKey == '') {
        var uuid = const Uuid();
        idempotencyKey = uuid.v4();
      }

      KYCPersonaInquiryResponse response = await backend.getPersonaInquiry(
        KYCPersonaInquiryRequest(idempotencyKey: idempotencyKey),
      );

      InquiryConfiguration config = InquiryIdConfiguration(
        inquiryId: response.id,
      );
      PersonaInquiry.init(
        configuration: config,
      );
      _canceledSubscription =
          PersonaInquiry.onCanceled.listen((event) => canceled = event);
      _errorSubscription =
          PersonaInquiry.onError.listen((event) => error = event);
      _completedSubscription = PersonaInquiry.onComplete.listen((event) {
        completed = event;
        appRouter.go('/');
      });
    } on InquiryError catch (e) {
      error = e;
      // TODO Show and error with a retry button?
      isInitializing = false;
    } finally {
      isInitializing = false;
    }
  }

  @action
  Future<void> startInquiry() async {
    if (!isInitializing) {
      // Button click should only trigger start after initialization
      isLoading = true;
      try {
        await PersonaInquiry.start();
      } on InquiryError catch (e) {
        error = e;
      } finally {
        isLoading = false;
      }
    }
  }

  @action
  void dispose() {
    _canceledSubscription.cancel();
    _errorSubscription.cancel();
    _completedSubscription.cancel();
  }
}
