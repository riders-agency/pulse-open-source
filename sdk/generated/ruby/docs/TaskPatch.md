# PulseSDK::TaskPatch

## Properties

| Name | Type | Description | Notes |
| ---- | ---- | ----------- | ----- |
| **title** | **String** |  | [optional] |
| **description** | **String** |  | [optional] |
| **priority** | **String** |  | [optional] |
| **due_date** | **String** |  | [optional] |
| **billable** | **Boolean** |  | [optional] |
| **expected_revision** | **Integer** |  | [optional] |

## Example

```ruby
require 'pulse_sdk'

instance = PulseSDK::TaskPatch.new(
  title: null,
  description: null,
  priority: null,
  due_date: null,
  billable: null,
  expected_revision: null
)
```
