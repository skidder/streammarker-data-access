When(/^I update sensor "([^"]*)" with "([^"]*)"$/) do |sensor_id, updates_fixture|
  @request = get_request(updates_fixture)
  @response = HTTParty.put("#{APPLICATION_ENDPOINT}/data-access/v1/sensor/#{sensor_id}",
                            body: @request,
                            headers: { 'Content-Type' => 'application/json'})
end