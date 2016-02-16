Given(/^I have no sensor readings in the database$/) do
  begin
    delete_sensor_readings_table(DateTime.now)
  rescue Exception => e
    # ignore    
  end

  begin
    delete_sensor_readings_table(DateTime.now.prev_month)
  rescue Exception => e
    # ignore    
  end

  begin
    delete_sensor_readings_table(DateTime.now.prev_month.prev_month)
  rescue Exception => e
    # ignore    
  end

  create_sensor_readings_table(DateTime.now)
  create_sensor_readings_table(DateTime.now.prev_month)
  create_sensor_readings_table(DateTime.now.prev_month.prev_month)
end

Given(/^I have two sensors without a readings table in the database for account "([^"]*)"$/) do |account_id|
  put_sensor_record(account_id, "1", "active")
  put_sensor_record(account_id, "2", "active")
  delete_sensor_readings_table(Time.now)
end

Given(/^I have one sensor "(.*)" in the database for account "(.*)"$/) do |sensor_id, account_id|
  put_sensor_record(account_id, sensor_id, "active")
end

Given(/^I have two sensors with readings in the database for account "(.*)"$/) do |account_id|
  put_sensor_record(account_id, "1", "active")
  put_sensor_reading(account_id, "1", 24.0, 78)
  put_sensor_reading(account_id, "1", 22.0, 80)
  put_sensor_record(account_id, "2", "active")
  put_sensor_reading(account_id, "2", 23.0, 75)
  put_sensor_reading(account_id, "2", 18.0, 82)
end

Given(/^I have two sensors without readings in the database for account "(.*)"$/) do |account_id|
  put_sensor_record(account_id, "1", "active")
  put_sensor_record(account_id, "2", "active")
end

Given(/^there are multiple readings for the current month with account "(.*)" and sensor "([^"]*)"$/) do |account_id, sensor_id|
  put_sensor_reading(account_id, sensor_id, 24.0, 78, DateTime.now.prev_day)
  put_sensor_reading(account_id, sensor_id, 22.0, 56, DateTime.now)
end

Given(/^there are multiple readings for the previous month with account "([^"]*)" and sensor "([^"]*)"$/) do |account_id, sensor_id|
  put_sensor_reading(account_id, sensor_id, 24.0, 78, DateTime.now.prev_month.prev_day)
  put_sensor_reading(account_id, sensor_id, 22.0, 81, DateTime.now.prev_month)
end

When(/^I query for sensor readings without time ranges for account "(.*)" and sensor "([^"]*)"$/) do |account_id, sensor_id|
  @response = HTTParty.get("#{APPLICATION_ENDPOINT}/data-access/v1/sensor_readings?account_id=#{account_id}&sensor_id=#{sensor_id}",
                            headers: { 'Accept' => 'application/json',
                                        'Accept-Encoding' => 'gzip' })
end

When(/^I query for readings from (\d+) months ago for account "([^"]*)" and sensor "([^"]*)"$/) do |months_ago, account_id, sensor_id|
  @response = HTTParty.get("#{APPLICATION_ENDPOINT}/data-access/v1/sensor_readings?account_id=#{account_id}&sensor_id=#{sensor_id}&start_time=#{DateTime.now.prev_month(months_ago.to_i).strftime("%s")}&end_time=#{Time.now.to_i}",
                            headers: { 'Accept' => 'application/json',
                                        'Accept-Encoding' => 'gzip' })
end

When(/^I retrieve the list of sensors for account "(.*)"$/) do |account_id|
  @response = HTTParty.get("#{APPLICATION_ENDPOINT}/data-access/v1/sensors/account/#{account_id}",
                            headers: { 'Accept' => 'application/json',
                                        'Accept-Encoding' => 'gzip' })

end

When(/^I retrieve the latest sensor readings for account "([^"]*)"$/) do |account_id|
  @response = HTTParty.get("#{APPLICATION_ENDPOINT}/data-access/v1/last_sensor_readings/account/#{account_id}",
                            headers: { 'Accept' => 'application/json',
                                        'Accept-Encoding' => 'gzip' })
end

Then(/^the response should be equal to "(.*)"$/) do |fixture|
  @response.body.should be_json_eql(get_response(fixture)).excluding("account_id", "sensor_id", "timestamp")
end

Then(/^the response is GZIP\-compressed$/) do
  puts @response.headers.inspect
  @response.headers['Vary'].should eq('Accept-Encoding')
end