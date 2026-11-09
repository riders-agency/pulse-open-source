# \TasksAPI

All URIs are relative to *https://pulse.example.test*

Method | HTTP request | Description
------------- | ------------- | -------------
[**ArchiveTask**](TasksAPI.md#ArchiveTask) | **Delete** /api/v1/work/tasks/{task_id} | archiveTask
[**CreateTask**](TasksAPI.md#CreateTask) | **Post** /api/v1/work/tasks | createTask
[**GetTask**](TasksAPI.md#GetTask) | **Get** /api/v1/work/tasks/{task_id} | getTask
[**ListTasks**](TasksAPI.md#ListTasks) | **Get** /api/v1/work/tasks | listTasks
[**RestoreTask**](TasksAPI.md#RestoreTask) | **Post** /api/v1/work/tasks/{task_id}/restore | restoreTask
[**UpdateTask**](TasksAPI.md#UpdateTask) | **Patch** /api/v1/work/tasks/{task_id} | updateTask



## ArchiveTask

> ArchiveResult ArchiveTask(ctx, taskId).Execute()

archiveTask

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/riders-agency/pulse-sdk-go"
)

func main() {
	taskId := "taskId_example" // string |

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TasksAPI.ArchiveTask(context.Background(), taskId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TasksAPI.ArchiveTask``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ArchiveTask`: ArchiveResult
	fmt.Fprintf(os.Stdout, "Response from `TasksAPI.ArchiveTask`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**taskId** | **string** |  |

### Other Parameters

Other parameters are passed through a pointer to a apiArchiveTaskRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**ArchiveResult**](ArchiveResult.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateTask

> Task CreateTask(ctx).TaskInput(taskInput).Execute()

createTask

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/riders-agency/pulse-sdk-go"
)

func main() {
	taskInput := *openapiclient.NewTaskInput("ProjectId_example", "Title_example") // TaskInput |

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TasksAPI.CreateTask(context.Background()).TaskInput(taskInput).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TasksAPI.CreateTask``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateTask`: Task
	fmt.Fprintf(os.Stdout, "Response from `TasksAPI.CreateTask`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateTaskRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **taskInput** | [**TaskInput**](TaskInput.md) |  |

### Return type

[**Task**](Task.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetTask

> TaskEnvelope GetTask(ctx, taskId).Execute()

getTask

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/riders-agency/pulse-sdk-go"
)

func main() {
	taskId := "taskId_example" // string |

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TasksAPI.GetTask(context.Background(), taskId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TasksAPI.GetTask``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetTask`: TaskEnvelope
	fmt.Fprintf(os.Stdout, "Response from `TasksAPI.GetTask`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**taskId** | **string** |  |

### Other Parameters

Other parameters are passed through a pointer to a apiGetTaskRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**TaskEnvelope**](TaskEnvelope.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListTasks

> TaskPage ListTasks(ctx).ProjectId(projectId).Limit(limit).Offset(offset).Search(search).Archived(archived).Execute()

listTasks

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/riders-agency/pulse-sdk-go"
)

func main() {
	projectId := "projectId_example" // string |
	limit := int32(56) // int32 |  (optional)
	offset := int32(56) // int32 |  (optional)
	search := "search_example" // string |  (optional)
	archived := true // bool |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TasksAPI.ListTasks(context.Background()).ProjectId(projectId).Limit(limit).Offset(offset).Search(search).Archived(archived).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TasksAPI.ListTasks``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListTasks`: TaskPage
	fmt.Fprintf(os.Stdout, "Response from `TasksAPI.ListTasks`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListTasksRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **projectId** | **string** |  |
 **limit** | **int32** |  |
 **offset** | **int32** |  |
 **search** | **string** |  |
 **archived** | **bool** |  |

### Return type

[**TaskPage**](TaskPage.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## RestoreTask

> ArchiveResult RestoreTask(ctx, taskId).Execute()

restoreTask

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/riders-agency/pulse-sdk-go"
)

func main() {
	taskId := "taskId_example" // string |

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TasksAPI.RestoreTask(context.Background(), taskId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TasksAPI.RestoreTask``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `RestoreTask`: ArchiveResult
	fmt.Fprintf(os.Stdout, "Response from `TasksAPI.RestoreTask`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**taskId** | **string** |  |

### Other Parameters

Other parameters are passed through a pointer to a apiRestoreTaskRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**ArchiveResult**](ArchiveResult.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateTask

> Task UpdateTask(ctx, taskId).TaskPatch(taskPatch).Execute()

updateTask

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/riders-agency/pulse-sdk-go"
)

func main() {
	taskId := "taskId_example" // string |
	taskPatch := *openapiclient.NewTaskPatch() // TaskPatch |

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TasksAPI.UpdateTask(context.Background(), taskId).TaskPatch(taskPatch).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TasksAPI.UpdateTask``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateTask`: Task
	fmt.Fprintf(os.Stdout, "Response from `TasksAPI.UpdateTask`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**taskId** | **string** |  |

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateTaskRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **taskPatch** | [**TaskPatch**](TaskPatch.md) |  |

### Return type

[**Task**](Task.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)
