Feature: Botanist Admin Portal
  As an admin
  I want to open the admin portal
  So that I can manage the Interledger wallet platform

  Background:
    Given the admin portal is running at "https://admin.mgnt.interledger.test"

  @botanist @smoke
  Scenario: Admin portal renders the navigation menu
    Given I navigate to the admin portal
    And I take a screenshot "admin-portal-loaded"
    Then the page title should be "Interledger Wallet Admin"
    And I take a screenshot "page-title-verified"
    And the navigation menu should be visible
    And I take a screenshot "navigation-menu-visible"
    And the "Home" menu item should be visible
    And I take a screenshot "home-menu-item"
    And the "Waitlist" menu item should be visible
    And I take a screenshot "waitlist-menu-item"
    And the "Wallets" menu item should be visible
    And I take a screenshot "wallets-menu-item"
    And the "Reviews" menu item should be visible
    And I take a screenshot "reviews-menu-item"
