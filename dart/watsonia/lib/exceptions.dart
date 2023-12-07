// TODO Look at how this is done with grpc client and possibly copy that rather

class UnauthorizedException implements Exception {} // session token is invalid

class InvalidCredentialsException implements Exception {
  // invalid credentials when sign in/sign up
  final Map<String, String> errors;

  InvalidCredentialsException(this.errors);
}

class UnknownException implements Exception {
  final String message;

  UnknownException(this.message);
}
