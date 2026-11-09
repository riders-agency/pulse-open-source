# \TasksApi

All URIs are relative to *https://pulse.example.test*

Method | HTTP request | Description
------------- | ------------- | -------------
[**archive_task**](TasksApi.md#archive_task) | **DELETE** /api/v1/work/tasks/{task_id} | archiveTask
[**create_task**](TasksApi.md#create_task) | **POST** /api/v1/work/tasks | createTask
[**get_task**](TasksApi.md#get_task) | **GET** /api/v1/work/tasks/{task_id} | getTask
[**list_tasks**](TasksApi.md#list_tasks) | **GET** /api/v1/work/tasks | listTasks
[**restore_task**](TasksApi.md#restore_task) | **POST** /api/v1/work/tasks/{task_id}/restore | restoreTask
[**update_task**](TasksApi.md#update_task) | **PATCH** /api/v1/work/tasks/{task_id} | updateTask



## archive_task

> models::ArchiveResult archive_task(task_id)
archiveTask

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**task_id** | **String** |  | [required] |

### Return type

[**models::ArchiveResult**](ArchiveResult.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## create_task

> models::Task create_task(task_input)
createTask

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**task_input** | [**TaskInput**](TaskInput.md) |  | [required] |

### Return type

[**models::Task**](Task.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## get_task

> models::TaskEnvelope get_task(task_id)
getTask

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**task_id** | **String** |  | [required] |

### Return type

[**models::TaskEnvelope**](TaskEnvelope.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## list_tasks

> models::TaskPage list_tasks(project_id, limit, offset, search, archived)
listTasks

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**project_id** | **String** |  | [required] |
**limit** | Option<**i32**> |  |  |
**offset** | Option<**i32**> |  |  |
**search** | Option<**String**> |  |  |
**archived** | Option<**bool**> |  |  |

### Return type

[**models::TaskPage**](TaskPage.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## restore_task

> models::ArchiveResult restore_task(task_id)
restoreTask

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**task_id** | **String** |  | [required] |

### Return type

[**models::ArchiveResult**](ArchiveResult.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## update_task

> models::Task update_task(task_id, task_patch)
updateTask

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**task_id** | **String** |  | [required] |
**task_patch** | [**TaskPatch**](TaskPatch.md) |  | [required] |

### Return type

[**models::Task**](Task.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)
