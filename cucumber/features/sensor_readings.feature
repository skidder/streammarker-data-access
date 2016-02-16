Feature: Sensor Readings API

  Background:
    Given I have no sensor readings in the database

  @happy
  Scenario: Query for sensor readings
    When I query for sensor readings without time ranges for account "account5" and sensor "sensor5"  
    Then the result should be a 200
    And the response is GZIP-compressed

  @happy
  Scenario: Query for readings in the current month without time bound
    Given I have one sensor "1" in the database for account "account1"
    And there are multiple readings for the current month with account "account1" and sensor "1"
    When I query for sensor readings without time ranges for account "account1" and sensor "1"
    Then the result should be a 200
    And the response should be equal to "query_for_readings_current_month"
    And the response is GZIP-compressed

  @happy
  Scenario: Query for readings in the current month with start-time one month ago
    Given I have one sensor "1" in the database for account "account1"
    And there are multiple readings for the current month with account "account1" and sensor "1"
    And there are multiple readings for the previous month with account "account1" and sensor "1"
    When I query for readings from 2 months ago for account "account1" and sensor "1"
    Then the result should be a 200
    And the response should be equal to "query_for_readings_last_2_months"
    And the response is GZIP-compressed

  @happy
  Scenario: Query for readings over last two months with readings only in previous month
    Given I have one sensor "1" in the database for account "account1"
    And there are multiple readings for the previous month with account "account1" and sensor "1"
    When I query for readings from 2 months ago for account "account1" and sensor "1"
    Then the result should be a 200
    And the response should be equal to "query_for_readings_last_2_months_with_1_month_activity"
    And the response is GZIP-compressed
