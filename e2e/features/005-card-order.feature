Feature: Card Ordering
  As a wallet user
  I want to order a card from my wallet
  So that I can make card payments

  Background:
    Given a random test identifier is generated
    And the frontend is running at "https://interledger.test"
    And mockgatehub is running at "https://mockgatehub.interledger.test"
    And Rafiki assets are seeded

  @cards @gatehub @card-order
  Scenario: Cards are not accessible without KYC
    Given the details of 'no-kyc-user' are
      | field       | value                        |
      | emailSuffix | nokyc@example.com            |
      | password    | InterlEdger2025!TestPassword |
      | country     | Germany                      |
      | firstName   | No                           |
      | lastName    | KYC                          |
      | dateOfBirth | 1990-01-01                   |
    And I complete the signup workflow for 'no-kyc-user'

    # KYC not done — manageWalletCardsEnabled is false
    Then the "Cards" navigation item should not be visible
    When I navigate to "/cards"
    Then I should be redirected to "/"

  @cards @gatehub @card-order
  Scenario: Successfully order a physical card with existing address
    Given the details of 'card-user' are
      | field       | value                        |
      | emailSuffix | cardorder@example.com        |
      | password    | InterlEdger2025!TestPassword |
      | country     | Germany                      |
      | firstName   | Card                         |
      | lastName    | Tester                       |
      | dateOfBirth | 1988-06-15                   |
    And I complete the minimal KYC flow `card-user`

    # KYC complete — manageWalletCardsEnabled is true, KYC address available
    When I navigate to the cards page
    Then I should see the cards page with an "Order card" button
    And I take a screenshot "cards-page-no-cards"

    When I click the "Order card" button
    Then I should be on the card order page
    And I take a screenshot "card-order-product-select"

    # Step 1: select a card product and choose physical
    When I select the first available card product
    And I click the "Physical" button
    Then I should be on the delivery address step
    And I take a screenshot "card-order-delivery-address"

    # Step 2: use the existing KYC address
    When I select the existing delivery address
    And I click the "Confirm" button
    Then I should be on the card order confirmation step
    And I take a screenshot "card-order-confirmation"

    # Step 3: confirm the order
    When I confirm the card order
    Then I should be redirected to "/cards"
    And I should see the snackbar "Your card in the making! We'll notify you as soon as it's ready to go."
    And I take a screenshot "card-order-success"
