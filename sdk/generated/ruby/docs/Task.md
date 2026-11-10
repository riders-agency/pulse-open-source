# PulseSDK::Task

## Properties

| Name | Type | Description | Notes |
| ---- | ---- | ----------- | ----- |
| **task_id** | **String** |  |  |
| **task_key** | **String** |  | [optional] |
| **project_id** | **String** |  | [optional] |
| **title** | **String** |  |  |
| **description** | **String** |  | [optional] |
| **type** | **String** |  | [optional] |
| **priority** | **String** |  | [optional] |
| **severity** | **String** |  | [optional] |
| **task_workflow_status_id** | **String** |  | [optional] |
| **due_date** | **String** |  | [optional] |
| **revision** | **Integer** |  | [optional] |
| **billable** | **Boolean** |  | [optional] |

## Example

```ruby
require 'pulse_sdk'

instance = PulseSDK::Task.new(
  task_id: null,
  task_key: null,
  project_id: null,
  title: null,
  description: null,
  type: null,
  priority: null,
  severity: null,
  task_workflow_status_id: null,
  due_date: null,
  revision: null,
  billable: null
)
```
