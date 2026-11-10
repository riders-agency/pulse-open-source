# PulseSDK::AuthApi

All URIs are relative to *https://pulse.example.test*

| Method | HTTP request | Description |
| ------ | ------------ | ----------- |
| [**get_identity**](AuthApi.md#get_identity) | **GET** /api/v1/auth/me | getIdentity |
| [**login**](AuthApi.md#login) | **POST** /api/v1/auth/login | login |
| [**logout**](AuthApi.md#logout) | **POST** /api/v1/auth/logout | logout |


## get_identity

> <Identity> get_identity

getIdentity

### Examples

```ruby
require 'time'
require 'pulse_sdk'
# setup authorization
PulseSDK.configure do |config|
  # Configure Bearer authorization: bearerAuth
  config.access_token = 'YOUR_BEARER_TOKEN'
end

api_instance = PulseSDK::AuthApi.new

begin
  # getIdentity
  result = api_instance.get_identity
  p result
rescue PulseSDK::ApiError => e
  puts "Error when calling AuthApi->get_identity: #{e}"
end
```

#### Using the get_identity_with_http_info variant

This returns an Array which contains the response data, status code and headers.

> <Array(<Identity>, Integer, Hash)> get_identity_with_http_info

```ruby
begin
  # getIdentity
  data, status_code, headers = api_instance.get_identity_with_http_info
  p status_code # => 2xx
  p headers # => { ... }
  p data # => <Identity>
rescue PulseSDK::ApiError => e
  puts "Error when calling AuthApi->get_identity_with_http_info: #{e}"
end
```

### Parameters

This endpoint does not need any parameter.

### Return type

[**Identity**](Identity.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json, application/problem+json


## login

> <TokenPair> login(login_input)

login

### Examples

```ruby
require 'time'
require 'pulse_sdk'

api_instance = PulseSDK::AuthApi.new
login_input = PulseSDK::LoginInput.new({email: 'email_example', password: 'password_example'}) # LoginInput |

begin
  # login
  result = api_instance.login(login_input)
  p result
rescue PulseSDK::ApiError => e
  puts "Error when calling AuthApi->login: #{e}"
end
```

#### Using the login_with_http_info variant

This returns an Array which contains the response data, status code and headers.

> <Array(<TokenPair>, Integer, Hash)> login_with_http_info(login_input)

```ruby
begin
  # login
  data, status_code, headers = api_instance.login_with_http_info(login_input)
  p status_code # => 2xx
  p headers # => { ... }
  p data # => <TokenPair>
rescue PulseSDK::ApiError => e
  puts "Error when calling AuthApi->login_with_http_info: #{e}"
end
```

### Parameters

| Name | Type | Description | Notes |
| ---- | ---- | ----------- | ----- |
| **login_input** | [**LoginInput**](LoginInput.md) |  |  |

### Return type

[**TokenPair**](TokenPair.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json, application/problem+json


## logout

> logout

logout

### Examples

```ruby
require 'time'
require 'pulse_sdk'
# setup authorization
PulseSDK.configure do |config|
  # Configure Bearer authorization: bearerAuth
  config.access_token = 'YOUR_BEARER_TOKEN'
end

api_instance = PulseSDK::AuthApi.new

begin
  # logout
  api_instance.logout
rescue PulseSDK::ApiError => e
  puts "Error when calling AuthApi->logout: #{e}"
end
```

#### Using the logout_with_http_info variant

This returns an Array which contains the response data (`nil` in this case), status code and headers.

> <Array(nil, Integer, Hash)> logout_with_http_info

```ruby
begin
  # logout
  data, status_code, headers = api_instance.logout_with_http_info
  p status_code # => 2xx
  p headers # => { ... }
  p data # => nil
rescue PulseSDK::ApiError => e
  puts "Error when calling AuthApi->logout_with_http_info: #{e}"
end
```

### Parameters

This endpoint does not need any parameter.

### Return type

nil (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/problem+json
