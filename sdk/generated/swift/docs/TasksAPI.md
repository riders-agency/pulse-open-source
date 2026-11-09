# TasksAPI

All URIs are relative to *https://pulse.example.test*

Method | HTTP request | Description
------------- | ------------- | -------------
[**archiveTask**](TasksAPI.md#archivetask) | **DELETE** /api/v1/work/tasks/{task_id} | archiveTask
[**createTask**](TasksAPI.md#createtask) | **POST** /api/v1/work/tasks | createTask
[**getTask**](TasksAPI.md#gettask) | **GET** /api/v1/work/tasks/{task_id} | getTask
[**listTasks**](TasksAPI.md#listtasks) | **GET** /api/v1/work/tasks | listTasks
[**restoreTask**](TasksAPI.md#restoretask) | **POST** /api/v1/work/tasks/{task_id}/restore | restoreTask
[**updateTask**](TasksAPI.md#updatetask) | **PATCH** /api/v1/work/tasks/{task_id} | updateTask


# **archiveTask**
```swift
    open class func archiveTask(taskId: String, completion: @escaping (_ data: ArchiveResult?, _ error: Error?) -> Void)
```

archiveTask

### Example
```swift
// The following code samples are still beta. For any issue, please report via http://github.com/OpenAPITools/openapi-generator/issues/new
import PulseSDK

let taskId = "taskId_example" // String |

// archiveTask
TasksAPI.archiveTask(taskId: taskId) { (response, error) in
    guard error == nil else {
        print(error)
        return
    }

    if (response) {
        dump(response)
    }
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **taskId** | **String** |  |

### Return type

[**ArchiveResult**](ArchiveResult.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **createTask**
```swift
    open class func createTask(taskInput: TaskInput, completion: @escaping (_ data: Task?, _ error: Error?) -> Void)
```

createTask

### Example
```swift
// The following code samples are still beta. For any issue, please report via http://github.com/OpenAPITools/openapi-generator/issues/new
import PulseSDK

let taskInput = TaskInput(projectId: "projectId_example", title: "title_example", description: "description_example", type: "type_example", priority: "priority_example", severity: "severity_example", taskWorkflowStatusId: "taskWorkflowStatusId_example", clientId: "clientId_example", dueDate: "dueDate_example", billable: false) // TaskInput |

// createTask
TasksAPI.createTask(taskInput: taskInput) { (response, error) in
    guard error == nil else {
        print(error)
        return
    }

    if (response) {
        dump(response)
    }
}
```

### Parameters

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getTask**
```swift
    open class func getTask(taskId: String, completion: @escaping (_ data: TaskEnvelope?, _ error: Error?) -> Void)
```

getTask

### Example
```swift
// The following code samples are still beta. For any issue, please report via http://github.com/OpenAPITools/openapi-generator/issues/new
import PulseSDK

let taskId = "taskId_example" // String |

// getTask
TasksAPI.getTask(taskId: taskId) { (response, error) in
    guard error == nil else {
        print(error)
        return
    }

    if (response) {
        dump(response)
    }
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **taskId** | **String** |  |

### Return type

[**TaskEnvelope**](TaskEnvelope.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listTasks**
```swift
    open class func listTasks(projectId: String, limit: Int? = nil, offset: Int? = nil, search: String? = nil, archived: Bool? = nil, completion: @escaping (_ data: TaskPage?, _ error: Error?) -> Void)
```

listTasks

### Example
```swift
// The following code samples are still beta. For any issue, please report via http://github.com/OpenAPITools/openapi-generator/issues/new
import PulseSDK

let projectId = "projectId_example" // String |
let limit = 987 // Int |  (optional)
let offset = 987 // Int |  (optional)
let search = "search_example" // String |  (optional)
let archived = true // Bool |  (optional)

// listTasks
TasksAPI.listTasks(projectId: projectId, limit: limit, offset: offset, search: search, archived: archived) { (response, error) in
    guard error == nil else {
        print(error)
        return
    }

    if (response) {
        dump(response)
    }
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **projectId** | **String** |  |
 **limit** | **Int** |  | [optional]
 **offset** | **Int** |  | [optional]
 **search** | **String** |  | [optional]
 **archived** | **Bool** |  | [optional]

### Return type

[**TaskPage**](TaskPage.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **restoreTask**
```swift
    open class func restoreTask(taskId: String, completion: @escaping (_ data: ArchiveResult?, _ error: Error?) -> Void)
```

restoreTask

### Example
```swift
// The following code samples are still beta. For any issue, please report via http://github.com/OpenAPITools/openapi-generator/issues/new
import PulseSDK

let taskId = "taskId_example" // String |

// restoreTask
TasksAPI.restoreTask(taskId: taskId) { (response, error) in
    guard error == nil else {
        print(error)
        return
    }

    if (response) {
        dump(response)
    }
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **taskId** | **String** |  |

### Return type

[**ArchiveResult**](ArchiveResult.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateTask**
```swift
    open class func updateTask(taskId: String, taskPatch: TaskPatch, completion: @escaping (_ data: Task?, _ error: Error?) -> Void)
```

updateTask

### Example
```swift
// The following code samples are still beta. For any issue, please report via http://github.com/OpenAPITools/openapi-generator/issues/new
import PulseSDK

let taskId = "taskId_example" // String |
let taskPatch = TaskPatch(title: "title_example", description: "description_example", priority: "priority_example", dueDate: "dueDate_example", billable: false, expectedRevision: 123) // TaskPatch |

// updateTask
TasksAPI.updateTask(taskId: taskId, taskPatch: taskPatch) { (response, error) in
    guard error == nil else {
        print(error)
        return
    }

    if (response) {
        dump(response)
    }
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **taskId** | **String** |  |
 **taskPatch** | [**TaskPatch**](TaskPatch.md) |  |

### Return type

[**Task**](Task.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)
