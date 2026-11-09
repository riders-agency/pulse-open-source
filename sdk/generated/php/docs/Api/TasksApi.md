# Riders\Pulse\TasksApi



All URIs are relative to https://pulse.example.test, except if the operation defines another base path.

| Method | HTTP request | Description |
| ------------- | ------------- | ------------- |
| [**archiveTask()**](TasksApi.md#archiveTask) | **DELETE** /api/v1/work/tasks/{task_id} | archiveTask |
| [**createTask()**](TasksApi.md#createTask) | **POST** /api/v1/work/tasks | createTask |
| [**getTask()**](TasksApi.md#getTask) | **GET** /api/v1/work/tasks/{task_id} | getTask |
| [**listTasks()**](TasksApi.md#listTasks) | **GET** /api/v1/work/tasks | listTasks |
| [**restoreTask()**](TasksApi.md#restoreTask) | **POST** /api/v1/work/tasks/{task_id}/restore | restoreTask |
| [**updateTask()**](TasksApi.md#updateTask) | **PATCH** /api/v1/work/tasks/{task_id} | updateTask |


## `archiveTask()`

```php
archiveTask($task_id): \Riders\Pulse\Model\ArchiveResult
```

archiveTask

### Example

```php
<?php
require_once(__DIR__ . '/vendor/autoload.php');


// Configure Bearer authorization: bearerAuth
$config = Riders\Pulse\Configuration::getDefaultConfiguration()->setAccessToken('YOUR_ACCESS_TOKEN');


$apiInstance = new Riders\Pulse\Api\TasksApi(
    // If you want use custom http client, pass your client which implements `GuzzleHttp\ClientInterface`.
    // This is optional, `GuzzleHttp\Client` will be used as default.
    new GuzzleHttp\Client(),
    $config
);
$task_id = 'task_id_example'; // string

try {
    $result = $apiInstance->archiveTask($task_id);
    print_r($result);
} catch (Exception $e) {
    echo 'Exception when calling TasksApi->archiveTask: ', $e->getMessage(), PHP_EOL;
}
```

### Parameters

| Name | Type | Description  | Notes |
| ------------- | ------------- | ------------- | ------------- |
| **task_id** | **string**|  | |

### Return type

[**\Riders\Pulse\Model\ArchiveResult**](../Model/ArchiveResult.md)

### Authorization

[bearerAuth](../../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`, `application/problem+json`

[[Back to top]](#) [[Back to API list]](../../README.md#endpoints)
[[Back to Model list]](../../README.md#models)
[[Back to README]](../../README.md)

## `createTask()`

```php
createTask($task_input): \Riders\Pulse\Model\Task
```

createTask

### Example

```php
<?php
require_once(__DIR__ . '/vendor/autoload.php');


// Configure Bearer authorization: bearerAuth
$config = Riders\Pulse\Configuration::getDefaultConfiguration()->setAccessToken('YOUR_ACCESS_TOKEN');


$apiInstance = new Riders\Pulse\Api\TasksApi(
    // If you want use custom http client, pass your client which implements `GuzzleHttp\ClientInterface`.
    // This is optional, `GuzzleHttp\Client` will be used as default.
    new GuzzleHttp\Client(),
    $config
);
$task_input = new \Riders\Pulse\Model\TaskInput(); // \Riders\Pulse\Model\TaskInput

try {
    $result = $apiInstance->createTask($task_input);
    print_r($result);
} catch (Exception $e) {
    echo 'Exception when calling TasksApi->createTask: ', $e->getMessage(), PHP_EOL;
}
```

### Parameters

| Name | Type | Description  | Notes |
| ------------- | ------------- | ------------- | ------------- |
| **task_input** | [**\Riders\Pulse\Model\TaskInput**](../Model/TaskInput.md)|  | |

### Return type

[**\Riders\Pulse\Model\Task**](../Model/Task.md)

### Authorization

[bearerAuth](../../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: `application/json`
- **Accept**: `application/json`, `application/problem+json`

[[Back to top]](#) [[Back to API list]](../../README.md#endpoints)
[[Back to Model list]](../../README.md#models)
[[Back to README]](../../README.md)

## `getTask()`

```php
getTask($task_id): \Riders\Pulse\Model\TaskEnvelope
```

getTask

### Example

```php
<?php
require_once(__DIR__ . '/vendor/autoload.php');


// Configure Bearer authorization: bearerAuth
$config = Riders\Pulse\Configuration::getDefaultConfiguration()->setAccessToken('YOUR_ACCESS_TOKEN');


$apiInstance = new Riders\Pulse\Api\TasksApi(
    // If you want use custom http client, pass your client which implements `GuzzleHttp\ClientInterface`.
    // This is optional, `GuzzleHttp\Client` will be used as default.
    new GuzzleHttp\Client(),
    $config
);
$task_id = 'task_id_example'; // string

try {
    $result = $apiInstance->getTask($task_id);
    print_r($result);
} catch (Exception $e) {
    echo 'Exception when calling TasksApi->getTask: ', $e->getMessage(), PHP_EOL;
}
```

### Parameters

| Name | Type | Description  | Notes |
| ------------- | ------------- | ------------- | ------------- |
| **task_id** | **string**|  | |

### Return type

[**\Riders\Pulse\Model\TaskEnvelope**](../Model/TaskEnvelope.md)

### Authorization

[bearerAuth](../../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`, `application/problem+json`

[[Back to top]](#) [[Back to API list]](../../README.md#endpoints)
[[Back to Model list]](../../README.md#models)
[[Back to README]](../../README.md)

## `listTasks()`

```php
listTasks($project_id, $limit, $offset, $search, $archived): \Riders\Pulse\Model\TaskPage
```

listTasks

### Example

```php
<?php
require_once(__DIR__ . '/vendor/autoload.php');


// Configure Bearer authorization: bearerAuth
$config = Riders\Pulse\Configuration::getDefaultConfiguration()->setAccessToken('YOUR_ACCESS_TOKEN');


$apiInstance = new Riders\Pulse\Api\TasksApi(
    // If you want use custom http client, pass your client which implements `GuzzleHttp\ClientInterface`.
    // This is optional, `GuzzleHttp\Client` will be used as default.
    new GuzzleHttp\Client(),
    $config
);
$project_id = 'project_id_example'; // string
$limit = 56; // int
$offset = 56; // int
$search = 'search_example'; // string
$archived = True; // bool

try {
    $result = $apiInstance->listTasks($project_id, $limit, $offset, $search, $archived);
    print_r($result);
} catch (Exception $e) {
    echo 'Exception when calling TasksApi->listTasks: ', $e->getMessage(), PHP_EOL;
}
```

### Parameters

| Name | Type | Description  | Notes |
| ------------- | ------------- | ------------- | ------------- |
| **project_id** | **string**|  | |
| **limit** | **int**|  | [optional] |
| **offset** | **int**|  | [optional] |
| **search** | **string**|  | [optional] |
| **archived** | **bool**|  | [optional] |

### Return type

[**\Riders\Pulse\Model\TaskPage**](../Model/TaskPage.md)

### Authorization

[bearerAuth](../../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`, `application/problem+json`

[[Back to top]](#) [[Back to API list]](../../README.md#endpoints)
[[Back to Model list]](../../README.md#models)
[[Back to README]](../../README.md)

## `restoreTask()`

```php
restoreTask($task_id): \Riders\Pulse\Model\ArchiveResult
```

restoreTask

### Example

```php
<?php
require_once(__DIR__ . '/vendor/autoload.php');


// Configure Bearer authorization: bearerAuth
$config = Riders\Pulse\Configuration::getDefaultConfiguration()->setAccessToken('YOUR_ACCESS_TOKEN');


$apiInstance = new Riders\Pulse\Api\TasksApi(
    // If you want use custom http client, pass your client which implements `GuzzleHttp\ClientInterface`.
    // This is optional, `GuzzleHttp\Client` will be used as default.
    new GuzzleHttp\Client(),
    $config
);
$task_id = 'task_id_example'; // string

try {
    $result = $apiInstance->restoreTask($task_id);
    print_r($result);
} catch (Exception $e) {
    echo 'Exception when calling TasksApi->restoreTask: ', $e->getMessage(), PHP_EOL;
}
```

### Parameters

| Name | Type | Description  | Notes |
| ------------- | ------------- | ------------- | ------------- |
| **task_id** | **string**|  | |

### Return type

[**\Riders\Pulse\Model\ArchiveResult**](../Model/ArchiveResult.md)

### Authorization

[bearerAuth](../../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`, `application/problem+json`

[[Back to top]](#) [[Back to API list]](../../README.md#endpoints)
[[Back to Model list]](../../README.md#models)
[[Back to README]](../../README.md)

## `updateTask()`

```php
updateTask($task_id, $task_patch): \Riders\Pulse\Model\Task
```

updateTask

### Example

```php
<?php
require_once(__DIR__ . '/vendor/autoload.php');


// Configure Bearer authorization: bearerAuth
$config = Riders\Pulse\Configuration::getDefaultConfiguration()->setAccessToken('YOUR_ACCESS_TOKEN');


$apiInstance = new Riders\Pulse\Api\TasksApi(
    // If you want use custom http client, pass your client which implements `GuzzleHttp\ClientInterface`.
    // This is optional, `GuzzleHttp\Client` will be used as default.
    new GuzzleHttp\Client(),
    $config
);
$task_id = 'task_id_example'; // string
$task_patch = new \Riders\Pulse\Model\TaskPatch(); // \Riders\Pulse\Model\TaskPatch

try {
    $result = $apiInstance->updateTask($task_id, $task_patch);
    print_r($result);
} catch (Exception $e) {
    echo 'Exception when calling TasksApi->updateTask: ', $e->getMessage(), PHP_EOL;
}
```

### Parameters

| Name | Type | Description  | Notes |
| ------------- | ------------- | ------------- | ------------- |
| **task_id** | **string**|  | |
| **task_patch** | [**\Riders\Pulse\Model\TaskPatch**](../Model/TaskPatch.md)|  | |

### Return type

[**\Riders\Pulse\Model\Task**](../Model/Task.md)

### Authorization

[bearerAuth](../../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: `application/json`
- **Accept**: `application/json`, `application/problem+json`

[[Back to top]](#) [[Back to API list]](../../README.md#endpoints)
[[Back to Model list]](../../README.md#models)
[[Back to README]](../../README.md)
