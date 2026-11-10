# PulseSDK::AIApi

All URIs are relative to *https://pulse.example.test*

| Method | HTTP request | Description |
| ------ | ------------ | ----------- |
| [**approve_proposal**](AIApi.md#approve_proposal) | **POST** /api/v1/projects/{project_id}/ai/tool-calls/{tool_call_id}/approve | approveProposal |
| [**cancel_session**](AIApi.md#cancel_session) | **POST** /api/v1/projects/{project_id}/ai/employee/sessions/{session_id}/cancel | cancelSession |
| [**create_session**](AIApi.md#create_session) | **POST** /api/v1/projects/{project_id}/ai/employee/sessions | createSession |
| [**edit_proposal**](AIApi.md#edit_proposal) | **PATCH** /api/v1/projects/{project_id}/ai/tool-calls/{tool_call_id} | editProposal |
| [**get_session**](AIApi.md#get_session) | **GET** /api/v1/projects/{project_id}/ai/employee/sessions/{session_id} | getSession |
| [**reject_proposal**](AIApi.md#reject_proposal) | **POST** /api/v1/projects/{project_id}/ai/tool-calls/{tool_call_id}/reject | rejectProposal |
| [**send_message**](AIApi.md#send_message) | **POST** /api/v1/projects/{project_id}/ai/employee/sessions/{session_id}/messages | sendMessage |


## approve_proposal

> <Session> approve_proposal(project_id, tool_call_id)

approveProposal

### Examples

```ruby
require 'time'
require 'pulse_sdk'
# setup authorization
PulseSDK.configure do |config|
  # Configure Bearer authorization: bearerAuth
  config.access_token = 'YOUR_BEARER_TOKEN'
end

api_instance = PulseSDK::AIApi.new
project_id = 'project_id_example' # String |
tool_call_id = 'tool_call_id_example' # String |

begin
  # approveProposal
  result = api_instance.approve_proposal(project_id, tool_call_id)
  p result
rescue PulseSDK::ApiError => e
  puts "Error when calling AIApi->approve_proposal: #{e}"
end
```

#### Using the approve_proposal_with_http_info variant

This returns an Array which contains the response data, status code and headers.

> <Array(<Session>, Integer, Hash)> approve_proposal_with_http_info(project_id, tool_call_id)

```ruby
begin
  # approveProposal
  data, status_code, headers = api_instance.approve_proposal_with_http_info(project_id, tool_call_id)
  p status_code # => 2xx
  p headers # => { ... }
  p data # => <Session>
rescue PulseSDK::ApiError => e
  puts "Error when calling AIApi->approve_proposal_with_http_info: #{e}"
end
```

### Parameters

| Name | Type | Description | Notes |
| ---- | ---- | ----------- | ----- |
| **project_id** | **String** |  |  |
| **tool_call_id** | **String** |  |  |

### Return type

[**Session**](Session.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json, application/problem+json


## cancel_session

> <Session> cancel_session(project_id, session_id)

cancelSession

### Examples

```ruby
require 'time'
require 'pulse_sdk'
# setup authorization
PulseSDK.configure do |config|
  # Configure Bearer authorization: bearerAuth
  config.access_token = 'YOUR_BEARER_TOKEN'
end

api_instance = PulseSDK::AIApi.new
project_id = 'project_id_example' # String |
session_id = 'session_id_example' # String |

begin
  # cancelSession
  result = api_instance.cancel_session(project_id, session_id)
  p result
rescue PulseSDK::ApiError => e
  puts "Error when calling AIApi->cancel_session: #{e}"
end
```

#### Using the cancel_session_with_http_info variant

This returns an Array which contains the response data, status code and headers.

> <Array(<Session>, Integer, Hash)> cancel_session_with_http_info(project_id, session_id)

```ruby
begin
  # cancelSession
  data, status_code, headers = api_instance.cancel_session_with_http_info(project_id, session_id)
  p status_code # => 2xx
  p headers # => { ... }
  p data # => <Session>
rescue PulseSDK::ApiError => e
  puts "Error when calling AIApi->cancel_session_with_http_info: #{e}"
end
```

### Parameters

| Name | Type | Description | Notes |
| ---- | ---- | ----------- | ----- |
| **project_id** | **String** |  |  |
| **session_id** | **String** |  |  |

### Return type

[**Session**](Session.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json, application/problem+json


## create_session

> <Session> create_session(project_id, session_input)

createSession

### Examples

```ruby
require 'time'
require 'pulse_sdk'
# setup authorization
PulseSDK.configure do |config|
  # Configure Bearer authorization: bearerAuth
  config.access_token = 'YOUR_BEARER_TOKEN'
end

api_instance = PulseSDK::AIApi.new
project_id = 'project_id_example' # String |
session_input = PulseSDK::SessionInput.new({screen_scope: 'screen_scope_example', screen_entity_type: 'screen_entity_type_example', screen_entity_id: 'screen_entity_id_example'}) # SessionInput |

begin
  # createSession
  result = api_instance.create_session(project_id, session_input)
  p result
rescue PulseSDK::ApiError => e
  puts "Error when calling AIApi->create_session: #{e}"
end
```

#### Using the create_session_with_http_info variant

This returns an Array which contains the response data, status code and headers.

> <Array(<Session>, Integer, Hash)> create_session_with_http_info(project_id, session_input)

```ruby
begin
  # createSession
  data, status_code, headers = api_instance.create_session_with_http_info(project_id, session_input)
  p status_code # => 2xx
  p headers # => { ... }
  p data # => <Session>
rescue PulseSDK::ApiError => e
  puts "Error when calling AIApi->create_session_with_http_info: #{e}"
end
```

### Parameters

| Name | Type | Description | Notes |
| ---- | ---- | ----------- | ----- |
| **project_id** | **String** |  |  |
| **session_input** | [**SessionInput**](SessionInput.md) |  |  |

### Return type

[**Session**](Session.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json, application/problem+json


## edit_proposal

> <Session> edit_proposal(project_id, tool_call_id, proposal_patch)

editProposal

### Examples

```ruby
require 'time'
require 'pulse_sdk'
# setup authorization
PulseSDK.configure do |config|
  # Configure Bearer authorization: bearerAuth
  config.access_token = 'YOUR_BEARER_TOKEN'
end

api_instance = PulseSDK::AIApi.new
project_id = 'project_id_example' # String |
tool_call_id = 'tool_call_id_example' # String |
proposal_patch = PulseSDK::ProposalPatch.new({input_json: { key: 3.56}}) # ProposalPatch |

begin
  # editProposal
  result = api_instance.edit_proposal(project_id, tool_call_id, proposal_patch)
  p result
rescue PulseSDK::ApiError => e
  puts "Error when calling AIApi->edit_proposal: #{e}"
end
```

#### Using the edit_proposal_with_http_info variant

This returns an Array which contains the response data, status code and headers.

> <Array(<Session>, Integer, Hash)> edit_proposal_with_http_info(project_id, tool_call_id, proposal_patch)

```ruby
begin
  # editProposal
  data, status_code, headers = api_instance.edit_proposal_with_http_info(project_id, tool_call_id, proposal_patch)
  p status_code # => 2xx
  p headers # => { ... }
  p data # => <Session>
rescue PulseSDK::ApiError => e
  puts "Error when calling AIApi->edit_proposal_with_http_info: #{e}"
end
```

### Parameters

| Name | Type | Description | Notes |
| ---- | ---- | ----------- | ----- |
| **project_id** | **String** |  |  |
| **tool_call_id** | **String** |  |  |
| **proposal_patch** | [**ProposalPatch**](ProposalPatch.md) |  |  |

### Return type

[**Session**](Session.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json, application/problem+json


## get_session

> <Session> get_session(project_id, session_id)

getSession

### Examples

```ruby
require 'time'
require 'pulse_sdk'
# setup authorization
PulseSDK.configure do |config|
  # Configure Bearer authorization: bearerAuth
  config.access_token = 'YOUR_BEARER_TOKEN'
end

api_instance = PulseSDK::AIApi.new
project_id = 'project_id_example' # String |
session_id = 'session_id_example' # String |

begin
  # getSession
  result = api_instance.get_session(project_id, session_id)
  p result
rescue PulseSDK::ApiError => e
  puts "Error when calling AIApi->get_session: #{e}"
end
```

#### Using the get_session_with_http_info variant

This returns an Array which contains the response data, status code and headers.

> <Array(<Session>, Integer, Hash)> get_session_with_http_info(project_id, session_id)

```ruby
begin
  # getSession
  data, status_code, headers = api_instance.get_session_with_http_info(project_id, session_id)
  p status_code # => 2xx
  p headers # => { ... }
  p data # => <Session>
rescue PulseSDK::ApiError => e
  puts "Error when calling AIApi->get_session_with_http_info: #{e}"
end
```

### Parameters

| Name | Type | Description | Notes |
| ---- | ---- | ----------- | ----- |
| **project_id** | **String** |  |  |
| **session_id** | **String** |  |  |

### Return type

[**Session**](Session.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json, application/problem+json


## reject_proposal

> <Session> reject_proposal(project_id, tool_call_id, rejection)

rejectProposal

### Examples

```ruby
require 'time'
require 'pulse_sdk'
# setup authorization
PulseSDK.configure do |config|
  # Configure Bearer authorization: bearerAuth
  config.access_token = 'YOUR_BEARER_TOKEN'
end

api_instance = PulseSDK::AIApi.new
project_id = 'project_id_example' # String |
tool_call_id = 'tool_call_id_example' # String |
rejection = PulseSDK::Rejection.new # Rejection |

begin
  # rejectProposal
  result = api_instance.reject_proposal(project_id, tool_call_id, rejection)
  p result
rescue PulseSDK::ApiError => e
  puts "Error when calling AIApi->reject_proposal: #{e}"
end
```

#### Using the reject_proposal_with_http_info variant

This returns an Array which contains the response data, status code and headers.

> <Array(<Session>, Integer, Hash)> reject_proposal_with_http_info(project_id, tool_call_id, rejection)

```ruby
begin
  # rejectProposal
  data, status_code, headers = api_instance.reject_proposal_with_http_info(project_id, tool_call_id, rejection)
  p status_code # => 2xx
  p headers # => { ... }
  p data # => <Session>
rescue PulseSDK::ApiError => e
  puts "Error when calling AIApi->reject_proposal_with_http_info: #{e}"
end
```

### Parameters

| Name | Type | Description | Notes |
| ---- | ---- | ----------- | ----- |
| **project_id** | **String** |  |  |
| **tool_call_id** | **String** |  |  |
| **rejection** | [**Rejection**](Rejection.md) |  |  |

### Return type

[**Session**](Session.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json, application/problem+json


## send_message

> <Session> send_message(project_id, session_id, message_input)

sendMessage

### Examples

```ruby
require 'time'
require 'pulse_sdk'
# setup authorization
PulseSDK.configure do |config|
  # Configure Bearer authorization: bearerAuth
  config.access_token = 'YOUR_BEARER_TOKEN'
end

api_instance = PulseSDK::AIApi.new
project_id = 'project_id_example' # String |
session_id = 'session_id_example' # String |
message_input = PulseSDK::MessageInput.new({body: 'body_example'}) # MessageInput |

begin
  # sendMessage
  result = api_instance.send_message(project_id, session_id, message_input)
  p result
rescue PulseSDK::ApiError => e
  puts "Error when calling AIApi->send_message: #{e}"
end
```

#### Using the send_message_with_http_info variant

This returns an Array which contains the response data, status code and headers.

> <Array(<Session>, Integer, Hash)> send_message_with_http_info(project_id, session_id, message_input)

```ruby
begin
  # sendMessage
  data, status_code, headers = api_instance.send_message_with_http_info(project_id, session_id, message_input)
  p status_code # => 2xx
  p headers # => { ... }
  p data # => <Session>
rescue PulseSDK::ApiError => e
  puts "Error when calling AIApi->send_message_with_http_info: #{e}"
end
```

### Parameters

| Name | Type | Description | Notes |
| ---- | ---- | ----------- | ----- |
| **project_id** | **String** |  |  |
| **session_id** | **String** |  |  |
| **message_input** | [**MessageInput**](MessageInput.md) |  |  |

### Return type

[**Session**](Session.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json, application/problem+json
