Feature: Retrieve latest sensor readings

  Background:
    Given I have no sensor readings in the database

  @happy
  Scenario: Retrieve latest readings for sensors in account
    Given I have two sensors with readings in the database for account "account2"
    When I retrieve the latest sensor readings for account "account2"
    Then the result should be a 200
    And the response should be equal to "latest_sensor_readings"

  @sad
  Scenario: Retrieve latest readings with no sensors in the account
    When I retrieve the latest sensor readings for account "account3"
    Then the result should be a 200
    And the response should be equal to "latest_sensor_readings_empty"

  @sad
  Scenario: Retrieve latest readings with sensors but no readings
    Given I have two sensors without readings in the database for account "account4"
    When I retrieve the latest sensor readings for account "account4"
    Then the result should be a 200
    And the response should be equal to "latest_sensor_readings_no_readings"

  @sad
  Scenario: Retrieve latest readings with sensors and no readings tables
    Given I have two sensors without a readings table in the database for account "account6"
    When I retrieve the latest sensor readings for account "account6"
    Then the result should be a 200
    And the response should be equal to "latest_sensor_readings_no_readings"
