import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;

class Auth extends ChangeNotifier {
  static const String _basePath = r'10.0.2.2:3030';

  // final api = OryKratosClient(basePathOverride: basePath).getFrontendApi();
  // try {
  //   final response = await api.createNativeLoginFlow();
  //   print(response);
  // } catch (e) {
  //   print("Exception when calling CourierApi->getCourierMessage: $e\n");
  // }
  // var url = Uri.http('cove-athletic-reed-scanning.trycloudflare.com', 'sessions/whoami');
  // var response = await http.get(url);
  // print('Response status: ${response.statusCode}');
  // print('Response body: ${response.body}');

  bool _isUser = false; // Should persist this value and the logged in token

  //TODO flow ids
  String _flowId = "";

  /// Whether user has logged in.
  bool get isUser => _isUser;

  Future<void> whoami() async {
    var url = Uri.https(_basePath, 'sessions/whoami');
    var response = await http.get(url);
    print('Response status: ${response.statusCode}');
    print('Response body: ${response.body}');
  }

  /// Signs out the current user.
  Future<void> signOut() async {
    await Future<void>.delayed(const Duration(milliseconds: 200));
    // Sign out.
    _isUser = false;
    notifyListeners();
  }

  /// Signs in a user.
  Future<bool> signIn(String username, String password) async {
    await Future<void>.delayed(const Duration(milliseconds: 200));

    // Sign in. Allow any password.
    _isUser = true;
    notifyListeners();
    return _isUser;
  }
}
