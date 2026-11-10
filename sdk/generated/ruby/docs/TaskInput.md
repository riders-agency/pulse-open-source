# PulseSDK::TaskInput

## Properties

| Name | Type | Description | Notes |
| ---- | ---- | ----------- | ----- |
| **project_id** | **String** |  |  |
| **title** | **String** |  |  |
| **description** | **String** |  | [optional] |
| **type** | **String** |  | [optional] |
| **priority** | **String** |  | [optional] |
| **severity** | **String** |  | [optional] |
| **task_workflow_status_id** | **String** |  | [optional] |
| **client_id** | **String** |  | [optional] |
| **due_date** | **String** |  | [optional] |
| **billable** | **Boolean** |  | [optional] |

## Example

```ruby
require 'pulse_sdk'

instance = PulseSDK::TaskInput.new(
  project_id: null,
  title: null,
  description: null,
  type: null,
  priority: null,
  severity: null,
  task_workflow_status_id: null,
  client_id: null,
  due_date: null,
  billable: null
)
```
