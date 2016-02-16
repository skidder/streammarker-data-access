Given(/^I have one sensor "(.*)" with location in the database for account "(.*)"$/) do |sensor_id, account_id|
  put_sensor_record(account_id, sensor_id, "active", 38.093455, -122.181369)
end

When(/^I retrieve the sensor "([^"]*)"$/) do |sensor_id|
  @response = HTTParty.get("#{APPLICATION_ENDPOINT}/data-access/v1/sensor/#{sensor_id}",
                            headers: { 'Accept' => 'application/json'})
end
