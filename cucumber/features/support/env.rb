require 'net/http'
require 'fileutils'
require 'childprocess'
require 'tempfile'
require 'httparty'
require 'json_spec'
require 'erubis'
require 'aws-sdk-v1'
require 'aws-sdk'
require 'json'
require 'time'
require 'influxdb'

require_relative 'feature_helper'

# Setup JSON Spec in a way which doesn't pull in Cucumber definitions which can conflict with our own.
# Also set it to ignore some fields by default.
World(JsonSpec::Helpers, JsonSpec::Matchers)
JsonSpec.configure do
  exclude_keys 'created_at', 'updated_at'
end

APPLICATION_ENDPOINT = 'http://localhost:3000'
HEALTHCHECK_ENDPOINT = 'http://localhost:3100'
LOG_DIR = 'cuke-logs'
DYNAMODB_DIR = 'dynamodb_local'
CUCUMBER_BASE = '.'
APPLICATION_LOG_FILE = File.join(LOG_DIR, 'application.log')

FAKES3_HOST = 'localhost'
FAKES3_PORT = '10020'
FAKES3_ROOT = '/tmp/fakes3_root'

FAKEDYNAMO_ROOT = Tempfile.new('dynamo').path # TODO: Clean me up
FAKEDYNAMO_HOST = 'localhost'
FAKEDYNAMO_PORT = '10040'

FAKESQS_HOST = 'localhost'
FAKESQS_PORT = '10030'

ENV['AWS_ACCESS_KEY_ID'] = "a"
ENV['AWS_SECRET_ACCESS_KEY'] = "b"
ENV['STREAMMARKER_DYNAMO_WAIT_TIME'] = '0s'
ENV['STREAMMARKER_DYNAMO_ACCOUNTS_TABLE'] = 'accounts'
ENV['STREAMMARKER_DYNAMO_RELAYS_TABLE'] = 'relays'
ENV['STREAMMARKER_DYNAMO_SENSOR_DEVICES_TABLE'] = 'sensors'
ENV['STREAMMARKER_DYNAMO_DEVICE_READINGS_TABLE'] = 'sensor_readings'
ENV['STREAMMARKER_DYNAMO_ENDPOINT'] = "http://#{FAKEDYNAMO_HOST}:#{FAKEDYNAMO_PORT}"
ENV['STREAMMARKER_DYNAMODB_DISABLE_SSL'] = 'TRUE'
ENV['AWS_REGION'] = 'us-east-1'
ENV['GOOGLE_API_KEY'] = 'foobar'

def wait_till_up_or_timeout
  healthy = false
  i = 0
  puts 'Waiting for system under test to start...'
  while (!healthy) && i < 30 do

    unless @app_process.alive?
      shutdown
      raise "The Application's child process exited undepectedly. Check #{APPLICATION_LOG_FILE} for details"
    end

    begin
      response = Net::HTTP.get_response(URI.parse(HEALTHCHECK_ENDPOINT + '/healthcheck'))
      if response.code == '200'
        healthy = true
      else
        puts 'Health check returned status code: ' + response.code
      end
    rescue Exception => e
      puts 'Encountered exception while polling Health check URL: ' + e.to_s
    end
    i = i + 1
    sleep(1) unless healthy
  end

  unless healthy
    shutdown
    raise 'Application failed to pass healthchecks within an acceptable amount of time. Declining to run tests.'
  end
end

def startup
  File.delete(DYNAMODB_DIR + "/shared-local-instance.db") if File.exists?(DYNAMODB_DIR + "/shared-local-instance.db")
  @fakedynamo_process = ChildProcess.build('java', '-Djava.library.path=./DynamoDBLocal_lib', '-jar', 'DynamoDBLocal.jar', '-sharedDb', '--port', FAKEDYNAMO_PORT)
  @fakedynamo_process.io.stdout = File.new(LOG_DIR + '/fakedynamo.log', 'w')
  @fakedynamo_process.cwd =DYNAMODB_DIR
  @fakedynamo_process.io.stderr = @fakedynamo_process.io.stdout
  @fakedynamo_process.leader = true
  @fakedynamo_process.start

  # Again, give dynamodb a second to start before we try to use it.
  sleep(1)

  @influxdb_process = ChildProcess.build('influxd')
  @influxdb_process.io.stdout = File.new(LOG_DIR + '/influxdb.log', 'w')
  @influxdb_process.io.stderr = @influxdb_process.io.stdout
  @influxdb_process.start

  # Again, give dynamodb a second to start before we try to use it.
  sleep(1)
  ddb = Aws::DynamoDB::Client.new(
      access_key_id: 'x',
      secret_access_key: 'y',
      endpoint: ENV['STREAMMARKER_DYNAMO_ENDPOINT']
  )

  resp = ddb.create_table(
      attribute_definitions: [{
                                  attribute_name: "id",
                                  attribute_type: "S",
                              }],
      table_name: "accounts",
      key_schema: [{
                       attribute_name: "id",
                       key_type: "HASH",
                   }],
      provisioned_throughput: {
          read_capacity_units: 1,
          write_capacity_units: 1,
      })

  resp = ddb.create_table(
      attribute_definitions: [{
                                  attribute_name: "id",
                                  attribute_type: "S",
                              }],
      table_name: "relays",
      key_schema: [{
                       attribute_name: "id",
                       key_type: "HASH",
                   }],
      provisioned_throughput: {
          read_capacity_units: 1,
          write_capacity_units: 1,
      })

  resp = ddb.create_table(
      attribute_definitions: [{
                                  attribute_name: "id",
                                  attribute_type: "S",
                              },
                              {
                                  attribute_name: "account_id",
                                  attribute_type: "S",
                              }],
      table_name: "sensors",
      global_secondary_indexes: [
        {
          index_name: "account_id-index", # required
          key_schema: [ # required
            {
              attribute_name: "account_id", # required
              key_type: "HASH", # required, accepts HASH, RANGE
            }
          ],
          projection: { # required
            projection_type: "ALL",
          },
          provisioned_throughput: {
              read_capacity_units: 1,
              write_capacity_units: 1,
          }
        },
      ],
      key_schema: [{
                       attribute_name: "id",
                       key_type: "HASH",
                   }],
      provisioned_throughput: {
          read_capacity_units: 1,
          write_capacity_units: 1,
      })

  puts 'Forking to start application under test'
  @app_process = ChildProcess.build('go', 'run', '../data_access.go')
  @app_process.io.stdout = File.new(APPLICATION_LOG_FILE, 'w')
  @app_process.io.stderr = @app_process.io.stdout
  @app_process.leader = true
  @app_process.start
end

def shutdown
  @app_process.stop
  @influxdb_process.stop
  @fakedynamo_process.stop
end

# Cucumber entry point

puts 'Application Endpoint: ' + APPLICATION_ENDPOINT.to_s
puts 'Log Directory: ' + LOG_DIR.to_s
puts "fakedynamo running at: #{FAKEDYNAMO_HOST}:#{FAKEDYNAMO_PORT}"

AWS.config(s3_endpoint: FAKES3_HOST, s3_port: FAKES3_PORT, use_ssl: false, s3_force_path_style: true, :access_key_id => 'x', :secret_access_key => 'y')

startup
wait_till_up_or_timeout

# ----- Cucumber Hooks ----- #

# Hook Cucumber exiting
at_exit do
  shutdown
end
