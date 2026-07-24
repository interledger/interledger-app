Feature: Add Public key
    As a verified user with KYC approved 
    I want to link a public key to my wallet
    So that I have external applications connected to the wallet

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
    Scenario: Successfully link a public key to the wallet
        When I navigate to "/settings/keys/add-public"
        And I fill in "Nickname" with "testPublicKey1"
        And I fill in "Public key" with "eyJrdHkiOiJPS1AiLCJjcnYiOiJFZDI1NTE5IiwieCI6IkFLdUFZY0RQa0ltdnlZZFAzQkZTT0w3UUlrc21tOENNWE5jMmhMNXRwcDgiLCJraWQiOiJiZjQwODcyMi1kY2MzLTQ1ZmItOGM3Zi0wMjBkYTVmYzU0NGQifQ=="
        And I click the "Save" button
        Then I should see the snackbar "Public key added successfully."

        # add the same key again
        When I navigate to "/settings/keys/add-public"
        And I fill in "Nickname" with "testPublicKey2"
        And I fill in "Public key" with "eyJrdHkiOiJPS1AiLCJjcnYiOiJFZDI1NTE5IiwieCI6IkFLdUFZY0RQa0ltdnlZZFAzQkZTT0w3UUlrc21tOENNWE5jMmhMNXRwcDgiLCJraWQiOiJiZjQwODcyMi1kY2MzLTQ1ZmItOGM3Zi0wMjBkYTVmYzU0NGQifQ=="
        And I click the "Save" button 
        Then I should see the snackbar " This public key is already linked to your wallet."