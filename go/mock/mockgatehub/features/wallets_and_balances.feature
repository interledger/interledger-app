Feature: Wallet lifecycle and balances
  As a wallet developer
  I want managed users to automatically receive XRPL wallets upon creation
  So that client apps can query wallets and balances without explicit wallet provisioning

  Background:
    Given an authenticated request using HMAC headers
    And a managed user id produced by /auth/v1/users/managed

  # In real GateHub, managed users are provisioned with a primary wallet
  # automatically. There is no separate "create wallet" API call.
  # The interledger-app expects GET /core/v1/users/{userID} to return
  # wallets already present on the user object.
  Scenario: Managed user has a wallet provisioned automatically
    When I GET /core/v1/users/{userId}
    Then a wallets array is returned with at least one wallet
    And the first wallet address starts with "r"
