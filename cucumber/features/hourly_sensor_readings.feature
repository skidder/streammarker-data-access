Feature: Hourly Sensor Readings API

  Background:
    Given I have no hourly sensor readings in the database

  @happy
  Scenario: Query for hourly sensor readings
    When I query for hourly sensor readings without time ranges for account "account5" and sensor "sensor5"  
    Then the result should be a 200

  @happy
  Scenario: Query for hourly readings in the current month without time bound
    Given I have one sensor "1" in the database for account "account1"
    And there are multiple hourly sensor readings for the current month with account "account1" and sensor "1"
    When I query for hourly sensor readings without time ranges for account "account1" and sensor "1"
    Then the result should be a 200
    And the response should be equal to "query_for_hourly_readings_current_month"

  @happy
  Scenario: Query for hourly readings in the current month with start-time one month ago
    Given I have one sensor "1" in the database for account "account1"
    And there are multiple hourly sensor readings for the current month with account "account1" and sensor "1"
    And there are multiple hourly sensor readings for the previous month with account "account1" and sensor "1"
    When I query for hourly readings from 2 months ago for account "account1" and sensor "1"
    Then the result should be a 200
    And the response should be equal to "query_for_hourly_readings_last_2_months"

  @happy
  Scenario: Query for hourly readings over last two months with readings only in previous month
    Given I have one sensor "1" in the database for account "account1"
    And there are multiple hourly sensor readings for the previous month with account "account1" and sensor "1"
    When I query for hourly readings from 2 months ago for account "account1" and sensor "1"
    Then the result should be a 200
    And the response should be equal to "query_for_hourly_readings_last_2_months_with_1_month_activity"
