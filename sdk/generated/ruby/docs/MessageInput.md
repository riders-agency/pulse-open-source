# PulseSDK::MessageInput

## Properties

| Name | Type | Description | Notes |
| ---- | ---- | ----------- | ----- |
| **body** | **String** |  |  |
| **page_context_json** | **Hash&lt;String, Object&gt;** |  | [optional] |
| **client_context_json** | **Hash&lt;String, Object&gt;** |  | [optional] |
| **active_tool_call_id** | **String** |  | [optional] |
| **local_date** | **String** |  | [optional] |
| **timezone** | **String** |  | [optional] |

## Example

```ruby
require 'pulse_sdk'

instance = PulseSDK::MessageInput.new(
  body: null,
  page_context_json: null,
  client_context_json: null,
  active_tool_call_id: null,
  local_date: null,
  timezone: null
)
```
