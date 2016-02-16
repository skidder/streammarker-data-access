Given(/^I have no hourly sensor readings in the database$/) do
  begin
    delete_hourly_sensor_readings_table(DateTime.now)
    delete_hourly_sensor_readings_table(DateTime.now.prev_month)
    delete_hourly_sensor_readings_table(DateTime.now.prev_month.prev_month)
  rescue Exception => e
    # ignore    
  end

  create_hourly_sensor_readings_table(DateTime.now)
  create_hourly_sensor_readings_table(DateTime.now.prev_month)
  create_hourly_sensor_readings_table(DateTime.now.prev_month.prev_month)
end

Given(/^there are multiple hourly sensor readings for the current month with account "([^"]*)" and sensor "([^"]*)"$/) do |account_id, sensor_id|
  put_hourly_sensor_reading(account_id, sensor_id, 24.0, 28.0, hourly_floor(DateTime.now - 1))
  put_hourly_sensor_reading(account_id, sensor_id, 22.0, 30.0, hourly_floor(DateTime.now))
end

Given(/^there are multiple hourly sensor readings for the previous month with account "([^"]*)" and sensor "([^"]*)"$/) do |account_id, sensor_id|
  put_hourly_sensor_reading(account_id, sensor_id, 24.0, 28.0, hourly_floor(DateTime.now.prev_month - 1))
  put_hourly_sensor_reading(account_id, sensor_id, 22.0, 30.0, hourly_floor(DateTime.now.prev_month))
end

When(/^I query for hourly sensor readings without time ranges for account "([^"]*)" and sensor "([^"]*)"$/) do |account_id, sensor_id|
  @response = HTTParty.get("#{APPLICATION_ENDPOINT}/data-access/v1/hourly_sensor_readings?account_id=#{account_id}&sensor_id=#{sensor_id}",
                            headers: { 'Accept' => 'application/json'})
end

When(/^I query for hourly readings from (\d+) months ago for account "([^"]*)" and sensor "([^"]*)"$/) do |months_ago, account_id, sensor_id|
  @response = HTTParty.get("#{APPLICATION_ENDPOINT}/data-access/v1/hourly_sensor_readings?account_id=#{account_id}&sensor_id=#{sensor_id}&start_time=#{DateTime.now.prev_month(months_ago.to_i).strftime("%s")}&end_time=#{Time.now.to_i}",
                            headers: { 'Accept' => 'application/json'})
end
