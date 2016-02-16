Feature: Update sensor

  @happy
  Scenario: Update a sensor
    Given I have one sensor "1" with location in the database for account "account1"
    When I update sensor "1" with "sensor_update"
    Then the result should be a 200
    And the response should be equal to "sensor_1_update"

  @sad
  Scenario: Update nonexistent sensor
    Given I have one sensor "1" with location in the database for account "account1"
    When I update sensor "999" with "sensor_update"
    Then the result should be a 400
