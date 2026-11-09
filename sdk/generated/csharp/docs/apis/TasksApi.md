# Riders.Pulse.Api.TasksApi

All URIs are relative to *https://pulse.example.test*

| Method | HTTP request | Description |
|--------|--------------|-------------|
| [**ArchiveTask**](TasksApi.md#archivetask) | **DELETE** /api/v1/work/tasks/{task_id} | archiveTask |
| [**CreateTask**](TasksApi.md#createtask) | **POST** /api/v1/work/tasks | createTask |
| [**GetTask**](TasksApi.md#gettask) | **GET** /api/v1/work/tasks/{task_id} | getTask |
| [**ListTasks**](TasksApi.md#listtasks) | **GET** /api/v1/work/tasks | listTasks |
| [**RestoreTask**](TasksApi.md#restoretask) | **POST** /api/v1/work/tasks/{task_id}/restore | restoreTask |
| [**UpdateTask**](TasksApi.md#updatetask) | **PATCH** /api/v1/work/tasks/{task_id} | updateTask |

<a id="archivetask"></a>
# **ArchiveTask**
> ArchiveResult ArchiveTask (string taskId)

archiveTask


### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **taskId** | **string** |  |  |

### Return type

[**ArchiveResult**](ArchiveResult.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Success |  -  |
| **0** | API error |  -  |

[[Back to top]](#) [[Back to API list]](../../README.md#documentation-for-api-endpoints) [[Back to Model list]](../../README.md#documentation-for-models) [[Back to README]](../../README.md)

<a id="createtask"></a>
# **CreateTask**
> Task CreateTask (TaskInput taskInput)

createTask


### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **taskInput** | [**TaskInput**](TaskInput.md) |  |  |

### Return type

[**Task**](Task.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **201** | Success |  -  |
| **0** | API error |  -  |

[[Back to top]](#) [[Back to API list]](../../README.md#documentation-for-api-endpoints) [[Back to Model list]](../../README.md#documentation-for-models) [[Back to README]](../../README.md)

<a id="gettask"></a>
# **GetTask**
> TaskEnvelope GetTask (string taskId)

getTask


### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **taskId** | **string** |  |  |

### Return type

[**TaskEnvelope**](TaskEnvelope.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Success |  -  |
| **0** | API error |  -  |

[[Back to top]](#) [[Back to API list]](../../README.md#documentation-for-api-endpoints) [[Back to Model list]](../../README.md#documentation-for-models) [[Back to README]](../../README.md)

<a id="listtasks"></a>
# **ListTasks**
> TaskPage ListTasks (string projectId, int limit = null, int offset = null, string search = null, bool archived = null)

listTasks


### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **projectId** | **string** |  |  |
| **limit** | **int** |  | [optional]  |
| **offset** | **int** |  | [optional]  |
| **search** | **string** |  | [optional]  |
| **archived** | **bool** |  | [optional]  |

### Return type

[**TaskPage**](TaskPage.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Success |  -  |
| **0** | API error |  -  |

[[Back to top]](#) [[Back to API list]](../../README.md#documentation-for-api-endpoints) [[Back to Model list]](../../README.md#documentation-for-models) [[Back to README]](../../README.md)

<a id="restoretask"></a>
# **RestoreTask**
> ArchiveResult RestoreTask (string taskId)

restoreTask


### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **taskId** | **string** |  |  |

### Return type

[**ArchiveResult**](ArchiveResult.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Success |  -  |
| **0** | API error |  -  |

[[Back to top]](#) [[Back to API list]](../../README.md#documentation-for-api-endpoints) [[Back to Model list]](../../README.md#documentation-for-models) [[Back to README]](../../README.md)

<a id="updatetask"></a>
# **UpdateTask**
> Task UpdateTask (string taskId, TaskPatch taskPatch)

updateTask


### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **taskId** | **string** |  |  |
| **taskPatch** | [**TaskPatch**](TaskPatch.md) |  |  |

### Return type

[**Task**](Task.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Success |  -  |
| **0** | API error |  -  |

[[Back to top]](#) [[Back to API list]](../../README.md#documentation-for-api-endpoints) [[Back to Model list]](../../README.md#documentation-for-models) [[Back to README]](../../README.md)
