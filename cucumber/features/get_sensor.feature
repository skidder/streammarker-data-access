Feature: List sensors

  @happy
  Scenario: Retrieve a single sensor
    Given I have one sensor "1" with location in the database for account "account1"
    When I retrieve the sensor "1"
    Then the result should be a 200
    And the response should be equal to "sensor_1"
