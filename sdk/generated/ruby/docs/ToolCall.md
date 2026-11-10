# PulseSDK::ToolCall

## Properties

| Name | Type | Description | Notes |
| ---- | ---- | ----------- | ----- |
| **ai_tool_call_id** | **String** |  |  |
| **tool_key** | **String** |  |  |
| **status** | **String** |  |  |
| **requires_approval** | **Boolean** |  | [optional] |
| **input_json** | **Hash&lt;String, Object&gt;** |  | [optional] |
| **output_json** | **Hash&lt;String, Object&gt;** |  | [optional] |

## Example

```ruby
require 'pulse_sdk'

instance = PulseSDK::ToolCall.new(
  ai_tool_call_id: null,
  tool_key: null,
  status: null,
  requires_approval: null,
  input_json: null,
  output_json: null
)
```
