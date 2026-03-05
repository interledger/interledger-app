Feature: Fee configuration and application
  As an e2e test runner
  I want to configure transaction fees dynamically via an API
  So that I can test fee-related flows without restarting MockGatehub

  # The /admin/fees endpoint requires no authentication and allows
  # tests to set deposit and withdrawal fee percentages at runtime.
  # Fees default to 0% so existing test suites are unaffected.

  Background:
    Given a clean MockGatehub instance
    And HMAC headers using app id "local-test-app-id" and secret "local-test-app-secret"

  # ── Fee configuration endpoint ──────────────────────────────────────

  Scenario: Default fees are zero
    When I GET /admin/fees without authentication
    Then the response status is 200
    And the deposit fee percentage is 0
    And the withdrawal fee percentage is 0

  Scenario: Configure deposit fee percentage
    When I PUT /admin/fees without authentication and body {"deposit_fee_percentage": 1.5}
    Then the response status is 200
    And the deposit fee percentage is 1.5
    And the withdrawal fee percentage is 0

  Scenario: Configure withdrawal fee percentage
    When I PUT /admin/fees without authentication and body {"withdrawal_fee_percentage": 2.0}
    Then the response status is 200
    And the deposit fee percentage is 0
    And the withdrawal fee percentage is 2.0

  Scenario: Configure both fees at once
    When I PUT /admin/fees without authentication and body {"deposit_fee_percentage": 1.5, "withdrawal_fee_percentage": 2.5}
    Then the response status is 200
    And the deposit fee percentage is 1.5
    And the withdrawal fee percentage is 2.5

  Scenario: Reset fees back to zero
    Given deposit fee is configured to 3.0%
    When I PUT /admin/fees without authentication and body {"deposit_fee_percentage": 0, "withdrawal_fee_percentage": 0}
    Then the response status is 200
    And the deposit fee percentage is 0
    And the withdrawal fee percentage is 0

  Scenario: Fee configuration does not require authentication
    When I PUT /admin/fees without any HMAC headers and body {"deposit_fee_percentage": 5.0}
    Then the response status is 200
    And the deposit fee percentage is 5.0

  Scenario: Negative fee percentage is rejected
    When I PUT /admin/fees without authentication and body {"deposit_fee_percentage": -1.0}
    Then the response status is 400

  Scenario: Fee percentage over 100 is rejected
    When I PUT /admin/fees without authentication and body {"deposit_fee_percentage": 101}
    Then the response status is 400

  # ── Fee applied to deposits ─────────────────────────────────────────

  Scenario: Deposit with 0% fee credits full amount
    Given deposit fee is configured to 0%
    And a managed user with at least one wallet address
    When I POST /core/v1/transactions with type 1, deposit_type "external", amount 100.00, currency "EUR", and a valid vault_uuid
    Then the response status is 201
    And the transaction fee is "0.00"
    And the transaction total_amount is "100.00"
    And the transaction amount is "100.00"
    When I GET /core/v1/transactions/{txId}
    Then the transaction fee is "0.00"
    When I GET /core/v1/wallets/{walletAddress}/balances
    Then the EUR balance is "100.00"

  Scenario: Deposit with 1.5% fee deducts fee from amount
    Given deposit fee is configured to 1.5%
    And a managed user with at least one wallet address
    When I POST /core/v1/transactions with type 1, deposit_type "external", amount 100.00, currency "EUR", and a valid vault_uuid
    Then the response status is 201
    And the transaction fee is "1.50"
    And the transaction total_amount is "100.00"
    And the transaction amount is "100.00"
    When I GET /core/v1/transactions/{txId}
    Then the transaction fee is "1.50"
    And the transaction total_amount is "100.00"
    When I GET /core/v1/wallets/{walletAddress}/balances
    Then the EUR balance is "98.50"

  Scenario: Deposit fee is reflected in GetTransaction response
    Given deposit fee is configured to 2.5%
    And a managed user with at least one wallet address
    When I POST /core/v1/transactions with type 1, deposit_type "external", amount 200.00, currency "EUR", and a valid vault_uuid
    Then the response status is 201
    And the transaction fee is "5.00"
    When I GET /core/v1/transactions/{txId}
    Then the transaction fee is "5.00"
    And the transaction amount is "200.00"
    And the transaction total_amount is "200.00"
    When I GET /core/v1/wallets/{walletAddress}/balances
    Then the EUR balance is "195.00"

  Scenario: Deposit via iframe with fee configured
    Given deposit fee is configured to 1.0%
    And a managed user with at least one wallet address
    And an iframe token obtained with scope ["deposit"] and mapped to the managed user
    When I POST /transaction/complete?paymentType=deposit&bearer={iframeToken} with amount "50.00" and currency "EUR" and Authorization header "Bearer {iframeToken}"
    Then the response status is 200 with status "success"

  # ── Fee applied to withdrawals ──────────────────────────────────────

  Scenario: Withdrawal with 0% fee has no fee
    Given withdrawal fee is configured to 0%
    And a managed user with at least one wallet address
    When I create a withdrawal transaction for 50.00 EUR
    Then the transaction fee is "0.00"
    And the transaction total_amount is "50.00"

  Scenario: Withdrawal with 2% fee adds fee to total
    Given withdrawal fee is configured to 2.0%
    And a managed user with at least one wallet address
    When I create a withdrawal transaction for 50.00 EUR
    Then the transaction fee is "1.00"
    And the transaction total_amount is "51.00"
    And the transaction amount is "50.00"

  # ── Fee precision and edge cases ────────────────────────────────────

  Scenario: Fee is rounded to two decimal places
    Given deposit fee is configured to 3.0%
    And a managed user with at least one wallet address
    When I POST /core/v1/transactions with type 1, deposit_type "external", amount 33.33, currency "EUR", and a valid vault_uuid
    Then the response status is 201
    And the transaction fee is "1.00"
    And fields amount, total_amount, and fee are string formatted with two decimals
    When I GET /core/v1/wallets/{walletAddress}/balances
    Then the EUR balance is "32.33"

  Scenario: Fee on small amount rounds down to zero
    Given deposit fee is configured to 0.1%
    And a managed user with at least one wallet address
    When I POST /core/v1/transactions with type 1, deposit_type "external", amount 0.01, currency "EUR", and a valid vault_uuid
    Then the response status is 201
    And the transaction fee is "0.00"
    When I GET /core/v1/wallets/{walletAddress}/balances
    Then the EUR balance is "0.01"

  Scenario: Fee configuration persists across requests
    When I PUT /admin/fees without authentication and body {"deposit_fee_percentage": 4.0}
    Then the deposit fee percentage is 4.0
    When I GET /admin/fees without authentication
    Then the deposit fee percentage is 4.0

  # ── Fee does not affect hosted transfers ────────────────────────────

  Scenario: Hosted transfer ignores deposit fee
    Given deposit fee is configured to 5.0%
    And a managed user with at least one wallet address
    When I POST /core/v1/transactions with user_id, amount 100.00, currency "EUR", type 2, and deposit_type "hosted"
    Then the transaction fee is "0.00"
    And the transaction total_amount is "100.00"
    When I GET /core/v1/wallets/{walletAddress}/balances
    Then the EUR balance is "100.00"

  # ── User-specific fee configuration ─────────────────────────────────

  Scenario: Get user-specific fees without authentication
    Given a managed user with at least one wallet address
    When I GET /admin/users/{userId}/fees without authentication
    Then the response status is 200
    And the deposit fee percentage is 0
    And the withdrawal fee percentage is 0
    And the deposit fee source is "global"
    And the withdrawal fee source is "global"

  Scenario: Set user-specific deposit fee without authentication
    Given a managed user with at least one wallet address
    When I PUT /admin/users/{userId}/fees without authentication and body {"deposit_fee_percentage": 2.5}
    Then the response status is 200
    And the deposit fee percentage is 2.5
    And the deposit fee source is "user"
    And the withdrawal fee source is "global"

  Scenario: Set user-specific withdrawal fee without authentication
    Given a managed user with at least one wallet address
    When I PUT /admin/users/{userId}/fees without authentication and body {"withdrawal_fee_percentage": 3.0}
    Then the response status is 200
    And the withdrawal fee percentage is 3.0
    And the withdrawal fee source is "user"
    And the deposit fee source is "global"

  Scenario: Clear user-specific fees without authentication
    Given a managed user with at least one wallet address
    And user-specific deposit fee is set to 2.5%
    When I DELETE /admin/users/{userId}/fees without authentication
    Then the response status is 200
    And the deposit fee source is "global"
    And the withdrawal fee source is "global"
