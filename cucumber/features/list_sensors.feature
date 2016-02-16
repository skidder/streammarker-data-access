Feature: List sensors

  @happy
  Scenario: Retrieve a single sensor
    Given I have one sensor "1" in the database for account "account1"
    When I retrieve the list of sensors for account "account1"
    Then the result should be a 200
    And the response should be equal to "sensors"

  @happy
  Scenario: Retrieve a single sensor with location
    Given I have one sensor "1" with location in the database for account "account1"
    When I retrieve the list of sensors for account "account1"
    Then the result should be a 200
    And the response should be equal to "sensors_with_location"
