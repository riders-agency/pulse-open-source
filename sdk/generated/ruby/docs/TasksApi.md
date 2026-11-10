# PulseSDK::TasksApi

All URIs are relative to *https://pulse.example.test*

| Method | HTTP request | Description |
| ------ | ------------ | ----------- |
| [**archive_task**](TasksApi.md#archive_task) | **DELETE** /api/v1/work/tasks/{task_id} | archiveTask |
| [**create_task**](TasksApi.md#create_task) | **POST** /api/v1/work/tasks | createTask |
| [**get_task**](TasksApi.md#get_task) | **GET** /api/v1/work/tasks/{task_id} | getTask |
| [**list_tasks**](TasksApi.md#list_tasks) | **GET** /api/v1/work/tasks | listTasks |
| [**restore_task**](TasksApi.md#restore_task) | **POST** /api/v1/work/tasks/{task_id}/restore | restoreTask |
| [**update_task**](TasksApi.md#update_task) | **PATCH** /api/v1/work/tasks/{task_id} | updateTask |


## archive_task

> <ArchiveResult> archive_task(task_id)

archiveTask

### Examples

```ruby
require 'time'
require 'pulse_sdk'
# setup authorization
PulseSDK.configure do |config|
  # Configure Bearer authorization: bearerAuth
  config.access_token = 'YOUR_BEARER_TOKEN'
end

api_instance = PulseSDK::TasksApi.new
task_id = 'task_id_example' # String |

begin
  # archiveTask
  result = api_instance.archive_task(task_id)
  p result
rescue PulseSDK::ApiError => e
  puts "Error when calling TasksApi->archive_task: #{e}"
end
```

#### Using the archive_task_with_http_info variant

This returns an Array which contains the response data, status code and headers.

> <Array(<ArchiveResult>, Integer, Hash)> archive_task_with_http_info(task_id)

```ruby
begin
  # archiveTask
  data, status_code, headers = api_instance.archive_task_with_http_info(task_id)
  p status_code # => 2xx
  p headers # => { ... }
  p data # => <ArchiveResult>
rescue PulseSDK::ApiError => e
  puts "Error when calling TasksApi->archive_task_with_http_info: #{e}"
end
```

### Parameters

| Name | Type | Description | Notes |
| ---- | ---- | ----------- | ----- |
| **task_id** | **String** |  |  |

### Return type

[**ArchiveResult**](ArchiveResult.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json, application/problem+json


## create_task

> <Task> create_task(task_input)

createTask

### Examples

```ruby
require 'time'
require 'pulse_sdk'
# setup authorization
PulseSDK.configure do |config|
  # Configure Bearer authorization: bearerAuth
  config.access_token = 'YOUR_BEARER_TOKEN'
end

api_instance = PulseSDK::TasksApi.new
task_input = PulseSDK::TaskInput.new({project_id: 'project_id_example', title: 'title_example'}) # TaskInput |

begin
  # createTask
  result = api_instance.create_task(task_input)
  p result
rescue PulseSDK::ApiError => e
  puts "Error when calling TasksApi->create_task: #{e}"
end
```

#### Using the create_task_with_http_info variant

This returns an Array which contains the response data, status code and headers.

> <Array(<Task>, Integer, Hash)> create_task_with_http_info(task_input)

```ruby
begin
  # createTask
  data, status_code, headers = api_instance.create_task_with_http_info(task_input)
  p status_code # => 2xx
  p headers # => { ... }
  p data # => <Task>
rescue PulseSDK::ApiError => e
  puts "Error when calling TasksApi->create_task_with_http_info: #{e}"
end
```

### Parameters

| Name | Type | Description | Notes |
| ---- | ---- | ----------- | ----- |
| **task_input** | [**TaskInput**](TaskInput.md) |  |  |

### Return type

[**Task**](Task.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json, application/problem+json


## get_task

> <TaskEnvelope> get_task(task_id)

getTask

### Examples

```ruby
require 'time'
require 'pulse_sdk'
# setup authorization
PulseSDK.configure do |config|
  # Configure Bearer authorization: bearerAuth
  config.access_token = 'YOUR_BEARER_TOKEN'
end

api_instance = PulseSDK::TasksApi.new
task_id = 'task_id_example' # String |

begin
  # getTask
  result = api_instance.get_task(task_id)
  p result
rescue PulseSDK::ApiError => e
  puts "Error when calling TasksApi->get_task: #{e}"
end
```

#### Using the get_task_with_http_info variant

This returns an Array which contains the response data, status code and headers.

> <Array(<TaskEnvelope>, Integer, Hash)> get_task_with_http_info(task_id)

```ruby
begin
  # getTask
  data, status_code, headers = api_instance.get_task_with_http_info(task_id)
  p status_code # => 2xx
  p headers # => { ... }
  p data # => <TaskEnvelope>
rescue PulseSDK::ApiError => e
  puts "Error when calling TasksApi->get_task_with_http_info: #{e}"
end
```

### Parameters

| Name | Type | Description | Notes |
| ---- | ---- | ----------- | ----- |
| **task_id** | **String** |  |  |

### Return type

[**TaskEnvelope**](TaskEnvelope.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json, application/problem+json


## list_tasks

> <TaskPage> list_tasks(project_id, opts)

listTasks

### Examples

```ruby
require 'time'
require 'pulse_sdk'
# setup authorization
PulseSDK.configure do |config|
  # Configure Bearer authorization: bearerAuth
  config.access_token = 'YOUR_BEARER_TOKEN'
end

api_instance = PulseSDK::TasksApi.new
project_id = 'project_id_example' # String |
opts = {
  limit: 56, # Integer |
  offset: 56, # Integer |
  search: 'search_example', # String |
  archived: true # Boolean |
}

begin
  # listTasks
  result = api_instance.list_tasks(project_id, opts)
  p result
rescue PulseSDK::ApiError => e
  puts "Error when calling TasksApi->list_tasks: #{e}"
end
```

#### Using the list_tasks_with_http_info variant

This returns an Array which contains the response data, status code and headers.

> <Array(<TaskPage>, Integer, Hash)> list_tasks_with_http_info(project_id, opts)

```ruby
begin
  # listTasks
  data, status_code, headers = api_instance.list_tasks_with_http_info(project_id, opts)
  p status_code # => 2xx
  p headers # => { ... }
  p data # => <TaskPage>
rescue PulseSDK::ApiError => e
  puts "Error when calling TasksApi->list_tasks_with_http_info: #{e}"
end
```

### Parameters

| Name | Type | Description | Notes |
| ---- | ---- | ----------- | ----- |
| **project_id** | **String** |  |  |
| **limit** | **Integer** |  | [optional] |
| **offset** | **Integer** |  | [optional] |
| **search** | **String** |  | [optional] |
| **archived** | **Boolean** |  | [optional] |

### Return type

[**TaskPage**](TaskPage.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json, application/problem+json


## restore_task

> <ArchiveResult> restore_task(task_id)

restoreTask

### Examples

```ruby
require 'time'
require 'pulse_sdk'
# setup authorization
PulseSDK.configure do |config|
  # Configure Bearer authorization: bearerAuth
  config.access_token = 'YOUR_BEARER_TOKEN'
end

api_instance = PulseSDK::TasksApi.new
task_id = 'task_id_example' # String |

begin
  # restoreTask
  result = api_instance.restore_task(task_id)
  p result
rescue PulseSDK::ApiError => e
  puts "Error when calling TasksApi->restore_task: #{e}"
end
```

#### Using the restore_task_with_http_info variant

This returns an Array which contains the response data, status code and headers.

> <Array(<ArchiveResult>, Integer, Hash)> restore_task_with_http_info(task_id)

```ruby
begin
  # restoreTask
  data, status_code, headers = api_instance.restore_task_with_http_info(task_id)
  p status_code # => 2xx
  p headers # => { ... }
  p data # => <ArchiveResult>
rescue PulseSDK::ApiError => e
  puts "Error when calling TasksApi->restore_task_with_http_info: #{e}"
end
```

### Parameters

| Name | Type | Description | Notes |
| ---- | ---- | ----------- | ----- |
| **task_id** | **String** |  |  |

### Return type

[**ArchiveResult**](ArchiveResult.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json, application/problem+json


## update_task

> <Task> update_task(task_id, task_patch)

updateTask

### Examples

```ruby
require 'time'
require 'pulse_sdk'
# setup authorization
PulseSDK.configure do |config|
  # Configure Bearer authorization: bearerAuth
  config.access_token = 'YOUR_BEARER_TOKEN'
end

api_instance = PulseSDK::TasksApi.new
task_id = 'task_id_example' # String |
task_patch = PulseSDK::TaskPatch.new # TaskPatch |

begin
  # updateTask
  result = api_instance.update_task(task_id, task_patch)
  p result
rescue PulseSDK::ApiError => e
  puts "Error when calling TasksApi->update_task: #{e}"
end
```

#### Using the update_task_with_http_info variant

This returns an Array which contains the response data, status code and headers.

> <Array(<Task>, Integer, Hash)> update_task_with_http_info(task_id, task_patch)

```ruby
begin
  # updateTask
  data, status_code, headers = api_instance.update_task_with_http_info(task_id, task_patch)
  p status_code # => 2xx
  p headers # => { ... }
  p data # => <Task>
rescue PulseSDK::ApiError => e
  puts "Error when calling TasksApi->update_task_with_http_info: #{e}"
end
```

### Parameters

| Name | Type | Description | Notes |
| ---- | ---- | ----------- | ----- |
| **task_id** | **String** |  |  |
| **task_patch** | [**TaskPatch**](TaskPatch.md) |  |  |

### Return type

[**Task**](Task.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json, application/problem+json
