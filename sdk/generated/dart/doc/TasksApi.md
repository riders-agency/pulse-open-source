# pulse_sdk.api.TasksApi

## Load the API package
```dart
import 'package:pulse_sdk/api.dart';
```

All URIs are relative to *https://pulse.example.test*

Method | HTTP request | Description
------------- | ------------- | -------------
[**archiveTask**](TasksApi.md#archivetask) | **DELETE** /api/v1/work/tasks/{task_id} | archiveTask
[**createTask**](TasksApi.md#createtask) | **POST** /api/v1/work/tasks | createTask
[**getTask**](TasksApi.md#gettask) | **GET** /api/v1/work/tasks/{task_id} | getTask
[**listTasks**](TasksApi.md#listtasks) | **GET** /api/v1/work/tasks | listTasks
[**restoreTask**](TasksApi.md#restoretask) | **POST** /api/v1/work/tasks/{task_id}/restore | restoreTask
[**updateTask**](TasksApi.md#updatetask) | **PATCH** /api/v1/work/tasks/{task_id} | updateTask


# **archiveTask**
> ArchiveResult archiveTask(taskId)

archiveTask

### Example
```dart
import 'package:pulse_sdk/api.dart';
// TODO Configure HTTP Bearer authorization: bearerAuth
// Case 1. Use String Token
//defaultApiClient.getAuthentication<HttpBearerAuth>('bearerAuth').setAccessToken('YOUR_ACCESS_TOKEN');
// Case 2. Use Function which generate token.
// String yourTokenGeneratorFunction() { ... }
//defaultApiClient.getAuthentication<HttpBearerAuth>('bearerAuth').setAccessToken(yourTokenGeneratorFunction);

final api_instance = TasksApi();
final taskId = taskId_example; // String |

try {
    final result = api_instance.archiveTask(taskId);
    print(result);
} catch (e) {
    print('Exception when calling TasksApi->archiveTask: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **taskId** | **String**|  |

### Return type

[**ArchiveResult**](ArchiveResult.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **createTask**
> Task createTask(taskInput)

createTask

### Example
```dart
import 'package:pulse_sdk/api.dart';
// TODO Configure HTTP Bearer authorization: bearerAuth
// Case 1. Use String Token
//defaultApiClient.getAuthentication<HttpBearerAuth>('bearerAuth').setAccessToken('YOUR_ACCESS_TOKEN');
// Case 2. Use Function which generate token.
// String yourTokenGeneratorFunction() { ... }
//defaultApiClient.getAuthentication<HttpBearerAuth>('bearerAuth').setAccessToken(yourTokenGeneratorFunction);

final api_instance = TasksApi();
final taskInput = TaskInput(); // TaskInput |

try {
    final result = api_instance.createTask(taskInput);
    print(result);
} catch (e) {
    print('Exception when calling TasksApi->createTask: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **taskInput** | [**TaskInput**](TaskInput.md)|  |

### Return type

[**Task**](Task.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getTask**
> TaskEnvelope getTask(taskId)

getTask

### Example
```dart
import 'package:pulse_sdk/api.dart';
// TODO Configure HTTP Bearer authorization: bearerAuth
// Case 1. Use String Token
//defaultApiClient.getAuthentication<HttpBearerAuth>('bearerAuth').setAccessToken('YOUR_ACCESS_TOKEN');
// Case 2. Use Function which generate token.
// String yourTokenGeneratorFunction() { ... }
//defaultApiClient.getAuthentication<HttpBearerAuth>('bearerAuth').setAccessToken(yourTokenGeneratorFunction);

final api_instance = TasksApi();
final taskId = taskId_example; // String |

try {
    final result = api_instance.getTask(taskId);
    print(result);
} catch (e) {
    print('Exception when calling TasksApi->getTask: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **taskId** | **String**|  |

### Return type

[**TaskEnvelope**](TaskEnvelope.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listTasks**
> TaskPage listTasks(projectId, limit, offset, search, archived)

listTasks

### Example
```dart
import 'package:pulse_sdk/api.dart';
// TODO Configure HTTP Bearer authorization: bearerAuth
// Case 1. Use String Token
//defaultApiClient.getAuthentication<HttpBearerAuth>('bearerAuth').setAccessToken('YOUR_ACCESS_TOKEN');
// Case 2. Use Function which generate token.
// String yourTokenGeneratorFunction() { ... }
//defaultApiClient.getAuthentication<HttpBearerAuth>('bearerAuth').setAccessToken(yourTokenGeneratorFunction);

final api_instance = TasksApi();
final projectId = projectId_example; // String |
final limit = 56; // int |
final offset = 56; // int |
final search = search_example; // String |
final archived = true; // bool |

try {
    final result = api_instance.listTasks(projectId, limit, offset, search, archived);
    print(result);
} catch (e) {
    print('Exception when calling TasksApi->listTasks: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **projectId** | **String**|  |
 **limit** | **int**|  | [optional]
 **offset** | **int**|  | [optional]
 **search** | **String**|  | [optional]
 **archived** | **bool**|  | [optional]

### Return type

[**TaskPage**](TaskPage.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **restoreTask**
> ArchiveResult restoreTask(taskId)

restoreTask

### Example
```dart
import 'package:pulse_sdk/api.dart';
// TODO Configure HTTP Bearer authorization: bearerAuth
// Case 1. Use String Token
//defaultApiClient.getAuthentication<HttpBearerAuth>('bearerAuth').setAccessToken('YOUR_ACCESS_TOKEN');
// Case 2. Use Function which generate token.
// String yourTokenGeneratorFunction() { ... }
//defaultApiClient.getAuthentication<HttpBearerAuth>('bearerAuth').setAccessToken(yourTokenGeneratorFunction);

final api_instance = TasksApi();
final taskId = taskId_example; // String |

try {
    final result = api_instance.restoreTask(taskId);
    print(result);
} catch (e) {
    print('Exception when calling TasksApi->restoreTask: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **taskId** | **String**|  |

### Return type

[**ArchiveResult**](ArchiveResult.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateTask**
> Task updateTask(taskId, taskPatch)

updateTask

### Example
```dart
import 'package:pulse_sdk/api.dart';
// TODO Configure HTTP Bearer authorization: bearerAuth
// Case 1. Use String Token
//defaultApiClient.getAuthentication<HttpBearerAuth>('bearerAuth').setAccessToken('YOUR_ACCESS_TOKEN');
// Case 2. Use Function which generate token.
// String yourTokenGeneratorFunction() { ... }
//defaultApiClient.getAuthentication<HttpBearerAuth>('bearerAuth').setAccessToken(yourTokenGeneratorFunction);

final api_instance = TasksApi();
final taskId = taskId_example; // String |
final taskPatch = TaskPatch(); // TaskPatch |

try {
    final result = api_instance.updateTask(taskId, taskPatch);
    print(result);
} catch (e) {
    print('Exception when calling TasksApi->updateTask: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **taskId** | **String**|  |
 **taskPatch** | [**TaskPatch**](TaskPatch.md)|  |

### Return type

[**Task**](Task.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)
