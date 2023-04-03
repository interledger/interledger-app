import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;

// Implement a class that implements calls to https://www.ory.sh/docs/kratos/reference/api#tag/frontend
// The class should extend ChangeNotifier so that it can be used with the Provider package.
// Use the Auth class below as a reference.

class Auth extends ChangeNotifier {
  static const String _basePath =
      r'stated-pattern-exists-milan.trycloudflare.com';

  bool _isUser = false; // Should persist this value and the logged in token
  String _sessioNToken = '';

  //TODO flow ids
  String _currentFlow = '';

  /// Whether user has logged in.
  bool get isUser => _isUser;

  Future<void> whoami() async {
    var url = Uri.https(_basePath, 'sessions/whoami');
    var response = await http.get(url);
    print('Response status: ${response.statusCode}');
    print('Response body: ${response.body}');
  }

  Future<void> createLoginFlow() async {
    var url = Uri.https(_basePath, 'self-service/login/api');
    var response = await http.get(url);
    print('Response status: ${response.statusCode}');
    print('Response body: ${response.body}');

    // Set flow
    _currentFlow = response.body;
  }

  Future<bool> login(String email, String password) async {
    var url = Uri.https(_basePath, 'self-service/login/api?flow=$_currentFlow');
    var response = await http.post(url,
        headers: <String, String>{
          'Content-Type': 'application/json; charset=UTF-8',
        },
        body: jsonEncode(<String, String>{
          'identifier': email,
          'password': password,
          'method': 'password'
        }));
    print('Response status: ${response.statusCode}');
    print('Response body: ${response.body}');

    var json = jsonDecode(response.body);

    if (response.statusCode == 200) {
      _sessioNToken = json['session_token'];
      _isUser = true;
      _currentFlow = '';
    }

    notifyListeners();
    return _isUser;
  }

  /// Signs out the current user.
  Future<void> signOut() async {
    await Future<void>.delayed(const Duration(milliseconds: 200));
    // Sign out.
    _isUser = false;
    notifyListeners();
  }

  Future<bool> logout() async {
    var url = Uri.https(_basePath, 'sessions/logout');
    var response = await http.post(url);
    print('Response status: ${response.statusCode}');
    print('Response body: ${response.body}');

    _isUser = false;
    return true;
  }

  Future<void> createRegisterFlow() async {
    var url = Uri.https(_basePath, 'self-service/registration/api');
    var response = await http.get(url);
    print('Response status: ${response.statusCode}');
    print('Response body: ${response.body}');

    // Set flow
    _currentFlow = response.body;
  }

  Future<bool> register(String email, String password) async {
    var url = Uri.https(
        _basePath, 'self-service/registration/api?flow=$_currentFlow');
    var response = await http.post(url,
        headers: <String, String>{
          'Content-Type': 'application/json; charset=UTF-8',
        },
        body: jsonEncode(<String, String>{
          'email': email,
          'password': password,
          'traits': '{"name": "Milan"}'
        }));
    print('Response status: ${response.statusCode}');
    print('Response body: ${response.body}');

    // TODO jsonDecode the response body and store the token in secure storage.

    _isUser = true;
    _currentFlow = '';
    return true;
  }
}
