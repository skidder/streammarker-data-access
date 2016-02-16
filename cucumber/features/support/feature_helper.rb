def get_request(name)
  get_json_from_fixture_file_as_hash('requests.json', name).to_json
end

def get_response(name)
  get_json_from_fixture_file_as_hash('responses.json', name).to_json
end

def get_json_from_fixture_file_as_hash(file, name)
  request = get_fixture_file_as_string(file)
  json = JSON.parse(request)[name]
  raise "Unable to find key '#{name}' in '#{file}'" if json.nil?
  json
end

def get_fixture_file_as_string(filename)
  File.read(File.join(CUCUMBER_BASE, 'fixtures', filename))
end

def put_sensor_record(account_id, sensor_id, state, latitude=nil, longitude=nil)
  ddb = get_dynamo_client
  if latitude && longitude
    ddb.put_item(table_name: "sensors",
                 item: {
                   "id" => "#{sensor_id}",
                   "account_id" => account_id,
                   "name" => "Sensor X",
                   "state" => state,
                   "location_enabled" => true,
                   "latitude" => latitude,
                   "longitude" => longitude,
                  }
                )
  else
    ddb.put_item(table_name: "sensors",
                 item: {
                   "id" => "#{sensor_id}",
                   "account_id" => account_id,
                   "name" => "Sensor X",
                   "state" => state,
                   "location_enabled" => false,
                  }
                )

  end
end

def put_hourly_sensor_reading(account_id, sensor_id, min_reading, max_reading, timestamp = Time.now)
  ddb = get_dynamo_client
  ddb.put_item(table_name: "hourly_sensor_readings_#{timestamp.strftime('%Y-%m')}",
             item: {
               "id" => "#{account_id}:#{sensor_id}",
               "sensor_id" => sensor_id,
               "account_id" => account_id,
               "name" => "Sensor",
               "state" => "active",
               "timestamp" => hourly_floor(timestamp).strftime("%s").to_i,
               "measurements" => "[{\"name\":\"soil_moisture_1\",\"min\":{\"name\":\"soil_moisture_1\",\"value\":#{min_reading},\"unit\":\"VWC\"},\"max\":{\"name\":\"soil_moisture_1\",\"value\":#{max_reading},\"unit\":\"VWC\"}}]",
              }
            )  
end

def hourly_floor(t)
  rounded = Time.at((t.to_time.to_i / 3600.0).floor * 3600)
  t.is_a?(DateTime) ? rounded.to_datetime : rounded
end

def put_sensor_reading(account_id, sensor_id, temperature, humidity, timestamp = Time.now)
  ddb = get_dynamo_client
  ddb.put_item(table_name: "sensor_readings_#{timestamp.strftime('%Y-%m')}",
             item: {
               "id" => "#{account_id}:#{sensor_id}",
               "sensor_id" => sensor_id,
               "account_id" => account_id,
               "name" => "Sensor",
               "state" => "active",
               "timestamp" => timestamp.strftime("%s").to_i,
               "measurements" => "[{\"name\":\"temperature\", \"value\": #{temperature}, \"unit\": \"Celsius\"}, {\"name\":\"humidity\", \"value\": #{humidity}, \"unit\": \"%\"}]",
              }
            )
end

def create_sensor_readings_table(timestamp)
  ddb = get_dynamo_client
  ddb.create_table(
      attribute_definitions: [{
                                  attribute_name: "id",
                                  attribute_type: "S",
                              },
                              {
                                  attribute_name: "timestamp",
                                  attribute_type: "N",
                              }],
      table_name: "sensor_readings_#{timestamp.strftime('%Y-%m')}",
      key_schema: [{
                       attribute_name: "id",
                       key_type: "HASH",
                   },
                   {
                       attribute_name: "timestamp",
                       key_type: "RANGE",
                   }],
      provisioned_throughput: {
          read_capacity_units: 1,
          write_capacity_units: 1,
      })
end

def create_hourly_sensor_readings_table(timestamp)
  ddb = get_dynamo_client
  ddb.create_table(
      attribute_definitions: [{
                                  attribute_name: "id",
                                  attribute_type: "S",
                              },
                              {
                                  attribute_name: "timestamp",
                                  attribute_type: "N",
                              }],
      table_name: "hourly_sensor_readings_#{timestamp.strftime('%Y-%m')}",
      key_schema: [{
                       attribute_name: "id",
                       key_type: "HASH",
                   },
                   {
                       attribute_name: "timestamp",
                       key_type: "RANGE",
                   }],
      provisioned_throughput: {
          read_capacity_units: 1,
          write_capacity_units: 1,
      })
end

def delete_sensor_readings_table(timestamp)
  ddb = get_dynamo_client
  ddb.delete_table(table_name: "sensor_readings_#{timestamp.strftime('%Y-%m')}")
end

def delete_hourly_sensor_readings_table(timestamp)
  ddb = get_dynamo_client
  ddb.delete_table(table_name: "hourly_sensor_readings_#{timestamp.strftime('%Y-%m')}")
end

def get_dynamo_client
  Aws::DynamoDB::Client.new(
    access_key_id: 'x',
    secret_access_key: 'y',
    endpoint: ENV['STREAMMARKER_DYNAMO_ENDPOINT']
  )
end
