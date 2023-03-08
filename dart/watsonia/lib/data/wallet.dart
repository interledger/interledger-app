import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;

class Wallet extends ChangeNotifier {

  String _paymentPointer = '';

  /// The user's formatted payment pointer
  String get paymentPointer => _paymentPointer;
  

  Future<void> getPaymentPointer() async {
    await Future<void>.delayed(const Duration(milliseconds: 200));
    // Sign out.
    _paymentPointer = "\$fynbos.me/cairin";
    notifyListeners();
  }
}
