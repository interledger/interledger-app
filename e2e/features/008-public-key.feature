Feature: Web Monetization - Public key management
    As a wallet user wanting to set up Web Monetizaton through the plugin
    I want to be able to manage mutiple public keys linked to my wallet
    So that I can control which browsers or other clients have access to my wallet through the frontend

    Background:
        Given a random test identifier is generated
        And the frontend is running at "https://interledger.test"
        And mockgatehub is running at "https://mockgatehub.interledger.test"
        And Rafiki assets are seeded

        Given the details of 'public-key-user' are
        | field           | value                        |
        | emailSuffix     | bob@example.com              |
        | password        | InterlEdger2025!TestPassword |
        | country         | Germany                      |
        | firstName       | Public                       |
        | lastName        | Key                          |
        | dateOfBirth     | 1990-03-15                   |
        And I complete the minimal KYC flow `public-key-user`

   @publicKey
    Scenario: Successfully link a new public key to the wallet
        When I navigate to "/settings/keys/add-public"
        And I fill in "Nickname" with "testPublicKey1"
        And I fill in "Public key" with "eyJrdHkiOiJPS1AiLCJjcnYiOiJFZDI1NTE5IiwieCI6IkFLdUFZY0RQa0ltdnlZZFAzQkZTT0w3UUlrc21tOENNWE5jMmhMNXRwcDgiLCJraWQiOiJiZjQwODcyMi1kY2MzLTQ1ZmItOGM3Zi0wMjBkYTVmYzU0NGQifQ=="
        And I click the "Save" button
        Then I should see the snackbar "Public key added successfully."

        # add the same key with a different Nickname
        When I navigate to "/settings/keys/add-public"
        And I fill in "Nickname" with "testPublicKey2"
        And I fill in "Public key" with "eyJrdHkiOiJPS1AiLCJjcnYiOiJFZDI1NTE5IiwieCI6IkFLdUFZY0RQa0ltdnlZZFAzQkZTT0w3UUlrc21tOENNWE5jMmhMNXRwcDgiLCJraWQiOiJiZjQwODcyMi1kY2MzLTQ1ZmItOGM3Zi0wMjBkYTVmYzU0NGQifQ=="
        And I click the "Save" button 
        Then I should see the snackbar "This public key is already linked to your wallet."

        # remove the public key 
        When I navigate to "/settings/keys"
        And I press on "testPublicKey1" 
        And I click the "Delete" button
        Then I should be redirected to "/settings/keys"
        And I should see the snackbar "Public key was deleted."

        # add a public key that is not valid
        When I navigate to "/settings/keys/add-public"
        And I fill in "Nickname" with "invalidKey"
        And I fill in "Public key" with "eyJrdHkiOiJPS1AiLCJjcnYiOiJFZDI1NTE5IiwieCI6IkJWY1JZY0RQa0ltdnlZZFA0QkZTT0w3UUlrc21tOENNWE5jMmhMNXRwcDgiLCJraWQiOiJiZjQwODcyNC1kY2M1LTQ3ZmItOGM2Zi0wMjFkYTVmYzUzNGQifQ==!@#"
        And I click the "Save" button 
        Then I should see the snackbar "There was an error with your request. Please retry. If the error continues, contact support."

