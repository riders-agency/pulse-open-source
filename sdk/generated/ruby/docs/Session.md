# PulseSDK::Session

## Properties

| Name | Type | Description | Notes |
| ---- | ---- | ----------- | ----- |
| **ai_chat_session_id** | **String** |  |  |
| **project_id** | **String** |  |  |
| **status** | **String** |  |  |
| **messages** | [**Array&lt;Message&gt;**](Message.md) |  |  |
| **tool_calls** | [**Array&lt;ToolCall&gt;**](ToolCall.md) |  | [optional] |

## Example

```ruby
require 'pulse_sdk'

instance = PulseSDK::Session.new(
  ai_chat_session_id: null,
  project_id: null,
  status: null,
  messages: null,
  tool_calls: null
)
```
