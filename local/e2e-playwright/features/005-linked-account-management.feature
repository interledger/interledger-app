Feature: Linked Account Management
  As a wallet user
  I want to be able to manage my linked Bank Accounts associated with my wallet
  So that I can easily configure the details of where withdrawals should go

Background: 
  Given a random test identifier is generated
  And the frontend is running at "https://interledger.test"
  And mockxago is running at "https://mockxago.interledger.test"
  And Rafiki assets are seeded

@xago @linked-accounts
Scenario: Link SA Fnb Account and make sure it is visible
  Given the details of 'fanie-zuma' are
    | field           | value                        |
    | emailSuffix     | link-fnb@example.com         |
    | password        | InterlEdger2025!TestPassword |
    | country         | South Africa                 |
    | firstName       | Fanie                        |
    | lastName        | Zuma                         |
    | dateOfBirth     | 1990-03-15                   |
  And I complete the minimal KYC flow `fanie-zuma`
  When I navigate to the dashboard Home
  And I wait "3" seconds for the page to load
  And I press on "Connect a bank account"
  Then I should see the "Connect bank account" form
  When I fill in "Account number" with "6208889997"
  And select Bank option "First National Bank"
  And press on "Continue"
  Then I should be navigated to dashboard "Accounts"
  And the linked account should be shown as "6******997"
  And the "Receive only" label should be shown for the account
  When I give the linked account the nickname "myfnb"
  Then the linked account should be shown as "myfnb"

  