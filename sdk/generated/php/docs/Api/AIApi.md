# Riders\Pulse\AIApi



All URIs are relative to https://pulse.example.test, except if the operation defines another base path.

| Method | HTTP request | Description |
| ------------- | ------------- | ------------- |
| [**approveProposal()**](AIApi.md#approveProposal) | **POST** /api/v1/projects/{project_id}/ai/tool-calls/{tool_call_id}/approve | approveProposal |
| [**cancelSession()**](AIApi.md#cancelSession) | **POST** /api/v1/projects/{project_id}/ai/employee/sessions/{session_id}/cancel | cancelSession |
| [**createSession()**](AIApi.md#createSession) | **POST** /api/v1/projects/{project_id}/ai/employee/sessions | createSession |
| [**editProposal()**](AIApi.md#editProposal) | **PATCH** /api/v1/projects/{project_id}/ai/tool-calls/{tool_call_id} | editProposal |
| [**getSession()**](AIApi.md#getSession) | **GET** /api/v1/projects/{project_id}/ai/employee/sessions/{session_id} | getSession |
| [**rejectProposal()**](AIApi.md#rejectProposal) | **POST** /api/v1/projects/{project_id}/ai/tool-calls/{tool_call_id}/reject | rejectProposal |
| [**sendMessage()**](AIApi.md#sendMessage) | **POST** /api/v1/projects/{project_id}/ai/employee/sessions/{session_id}/messages | sendMessage |


## `approveProposal()`

```php
approveProposal($project_id, $tool_call_id): \Riders\Pulse\Model\Session
```

approveProposal

### Example

```php
<?php
require_once(__DIR__ . '/vendor/autoload.php');


// Configure Bearer authorization: bearerAuth
$config = Riders\Pulse\Configuration::getDefaultConfiguration()->setAccessToken('YOUR_ACCESS_TOKEN');


$apiInstance = new Riders\Pulse\Api\AIApi(
    // If you want use custom http client, pass your client which implements `GuzzleHttp\ClientInterface`.
    // This is optional, `GuzzleHttp\Client` will be used as default.
    new GuzzleHttp\Client(),
    $config
);
$project_id = 'project_id_example'; // string
$tool_call_id = 'tool_call_id_example'; // string

try {
    $result = $apiInstance->approveProposal($project_id, $tool_call_id);
    print_r($result);
} catch (Exception $e) {
    echo 'Exception when calling AIApi->approveProposal: ', $e->getMessage(), PHP_EOL;
}
```

### Parameters

| Name | Type | Description  | Notes |
| ------------- | ------------- | ------------- | ------------- |
| **project_id** | **string**|  | |
| **tool_call_id** | **string**|  | |

### Return type

[**\Riders\Pulse\Model\Session**](../Model/Session.md)

### Authorization

[bearerAuth](../../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`, `application/problem+json`

[[Back to top]](#) [[Back to API list]](../../README.md#endpoints)
[[Back to Model list]](../../README.md#models)
[[Back to README]](../../README.md)

## `cancelSession()`

```php
cancelSession($project_id, $session_id): \Riders\Pulse\Model\Session
```

cancelSession

### Example

```php
<?php
require_once(__DIR__ . '/vendor/autoload.php');


// Configure Bearer authorization: bearerAuth
$config = Riders\Pulse\Configuration::getDefaultConfiguration()->setAccessToken('YOUR_ACCESS_TOKEN');


$apiInstance = new Riders\Pulse\Api\AIApi(
    // If you want use custom http client, pass your client which implements `GuzzleHttp\ClientInterface`.
    // This is optional, `GuzzleHttp\Client` will be used as default.
    new GuzzleHttp\Client(),
    $config
);
$project_id = 'project_id_example'; // string
$session_id = 'session_id_example'; // string

try {
    $result = $apiInstance->cancelSession($project_id, $session_id);
    print_r($result);
} catch (Exception $e) {
    echo 'Exception when calling AIApi->cancelSession: ', $e->getMessage(), PHP_EOL;
}
```

### Parameters

| Name | Type | Description  | Notes |
| ------------- | ------------- | ------------- | ------------- |
| **project_id** | **string**|  | |
| **session_id** | **string**|  | |

### Return type

[**\Riders\Pulse\Model\Session**](../Model/Session.md)

### Authorization

[bearerAuth](../../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`, `application/problem+json`

[[Back to top]](#) [[Back to API list]](../../README.md#endpoints)
[[Back to Model list]](../../README.md#models)
[[Back to README]](../../README.md)

## `createSession()`

```php
createSession($project_id, $session_input): \Riders\Pulse\Model\Session
```

createSession

### Example

```php
<?php
require_once(__DIR__ . '/vendor/autoload.php');


// Configure Bearer authorization: bearerAuth
$config = Riders\Pulse\Configuration::getDefaultConfiguration()->setAccessToken('YOUR_ACCESS_TOKEN');


$apiInstance = new Riders\Pulse\Api\AIApi(
    // If you want use custom http client, pass your client which implements `GuzzleHttp\ClientInterface`.
    // This is optional, `GuzzleHttp\Client` will be used as default.
    new GuzzleHttp\Client(),
    $config
);
$project_id = 'project_id_example'; // string
$session_input = new \Riders\Pulse\Model\SessionInput(); // \Riders\Pulse\Model\SessionInput

try {
    $result = $apiInstance->createSession($project_id, $session_input);
    print_r($result);
} catch (Exception $e) {
    echo 'Exception when calling AIApi->createSession: ', $e->getMessage(), PHP_EOL;
}
```

### Parameters

| Name | Type | Description  | Notes |
| ------------- | ------------- | ------------- | ------------- |
| **project_id** | **string**|  | |
| **session_input** | [**\Riders\Pulse\Model\SessionInput**](../Model/SessionInput.md)|  | |

### Return type

[**\Riders\Pulse\Model\Session**](../Model/Session.md)

### Authorization

[bearerAuth](../../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: `application/json`
- **Accept**: `application/json`, `application/problem+json`

[[Back to top]](#) [[Back to API list]](../../README.md#endpoints)
[[Back to Model list]](../../README.md#models)
[[Back to README]](../../README.md)

## `editProposal()`

```php
editProposal($project_id, $tool_call_id, $proposal_patch): \Riders\Pulse\Model\Session
```

editProposal

### Example

```php
<?php
require_once(__DIR__ . '/vendor/autoload.php');


// Configure Bearer authorization: bearerAuth
$config = Riders\Pulse\Configuration::getDefaultConfiguration()->setAccessToken('YOUR_ACCESS_TOKEN');


$apiInstance = new Riders\Pulse\Api\AIApi(
    // If you want use custom http client, pass your client which implements `GuzzleHttp\ClientInterface`.
    // This is optional, `GuzzleHttp\Client` will be used as default.
    new GuzzleHttp\Client(),
    $config
);
$project_id = 'project_id_example'; // string
$tool_call_id = 'tool_call_id_example'; // string
$proposal_patch = new \Riders\Pulse\Model\ProposalPatch(); // \Riders\Pulse\Model\ProposalPatch

try {
    $result = $apiInstance->editProposal($project_id, $tool_call_id, $proposal_patch);
    print_r($result);
} catch (Exception $e) {
    echo 'Exception when calling AIApi->editProposal: ', $e->getMessage(), PHP_EOL;
}
```

### Parameters

| Name | Type | Description  | Notes |
| ------------- | ------------- | ------------- | ------------- |
| **project_id** | **string**|  | |
| **tool_call_id** | **string**|  | |
| **proposal_patch** | [**\Riders\Pulse\Model\ProposalPatch**](../Model/ProposalPatch.md)|  | |

### Return type

[**\Riders\Pulse\Model\Session**](../Model/Session.md)

### Authorization

[bearerAuth](../../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: `application/json`
- **Accept**: `application/json`, `application/problem+json`

[[Back to top]](#) [[Back to API list]](../../README.md#endpoints)
[[Back to Model list]](../../README.md#models)
[[Back to README]](../../README.md)

## `getSession()`

```php
getSession($project_id, $session_id): \Riders\Pulse\Model\Session
```

getSession

### Example

```php
<?php
require_once(__DIR__ . '/vendor/autoload.php');


// Configure Bearer authorization: bearerAuth
$config = Riders\Pulse\Configuration::getDefaultConfiguration()->setAccessToken('YOUR_ACCESS_TOKEN');


$apiInstance = new Riders\Pulse\Api\AIApi(
    // If you want use custom http client, pass your client which implements `GuzzleHttp\ClientInterface`.
    // This is optional, `GuzzleHttp\Client` will be used as default.
    new GuzzleHttp\Client(),
    $config
);
$project_id = 'project_id_example'; // string
$session_id = 'session_id_example'; // string

try {
    $result = $apiInstance->getSession($project_id, $session_id);
    print_r($result);
} catch (Exception $e) {
    echo 'Exception when calling AIApi->getSession: ', $e->getMessage(), PHP_EOL;
}
```

### Parameters

| Name | Type | Description  | Notes |
| ------------- | ------------- | ------------- | ------------- |
| **project_id** | **string**|  | |
| **session_id** | **string**|  | |

### Return type

[**\Riders\Pulse\Model\Session**](../Model/Session.md)

### Authorization

[bearerAuth](../../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`, `application/problem+json`

[[Back to top]](#) [[Back to API list]](../../README.md#endpoints)
[[Back to Model list]](../../README.md#models)
[[Back to README]](../../README.md)

## `rejectProposal()`

```php
rejectProposal($project_id, $tool_call_id, $rejection): \Riders\Pulse\Model\Session
```

rejectProposal

### Example

```php
<?php
require_once(__DIR__ . '/vendor/autoload.php');


// Configure Bearer authorization: bearerAuth
$config = Riders\Pulse\Configuration::getDefaultConfiguration()->setAccessToken('YOUR_ACCESS_TOKEN');


$apiInstance = new Riders\Pulse\Api\AIApi(
    // If you want use custom http client, pass your client which implements `GuzzleHttp\ClientInterface`.
    // This is optional, `GuzzleHttp\Client` will be used as default.
    new GuzzleHttp\Client(),
    $config
);
$project_id = 'project_id_example'; // string
$tool_call_id = 'tool_call_id_example'; // string
$rejection = new \Riders\Pulse\Model\Rejection(); // \Riders\Pulse\Model\Rejection

try {
    $result = $apiInstance->rejectProposal($project_id, $tool_call_id, $rejection);
    print_r($result);
} catch (Exception $e) {
    echo 'Exception when calling AIApi->rejectProposal: ', $e->getMessage(), PHP_EOL;
}
```

### Parameters

| Name | Type | Description  | Notes |
| ------------- | ------------- | ------------- | ------------- |
| **project_id** | **string**|  | |
| **tool_call_id** | **string**|  | |
| **rejection** | [**\Riders\Pulse\Model\Rejection**](../Model/Rejection.md)|  | |

### Return type

[**\Riders\Pulse\Model\Session**](../Model/Session.md)

### Authorization

[bearerAuth](../../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: `application/json`
- **Accept**: `application/json`, `application/problem+json`

[[Back to top]](#) [[Back to API list]](../../README.md#endpoints)
[[Back to Model list]](../../README.md#models)
[[Back to README]](../../README.md)

## `sendMessage()`

```php
sendMessage($project_id, $session_id, $message_input): \Riders\Pulse\Model\Session
```

sendMessage

### Example

```php
<?php
require_once(__DIR__ . '/vendor/autoload.php');


// Configure Bearer authorization: bearerAuth
$config = Riders\Pulse\Configuration::getDefaultConfiguration()->setAccessToken('YOUR_ACCESS_TOKEN');


$apiInstance = new Riders\Pulse\Api\AIApi(
    // If you want use custom http client, pass your client which implements `GuzzleHttp\ClientInterface`.
    // This is optional, `GuzzleHttp\Client` will be used as default.
    new GuzzleHttp\Client(),
    $config
);
$project_id = 'project_id_example'; // string
$session_id = 'session_id_example'; // string
$message_input = new \Riders\Pulse\Model\MessageInput(); // \Riders\Pulse\Model\MessageInput

try {
    $result = $apiInstance->sendMessage($project_id, $session_id, $message_input);
    print_r($result);
} catch (Exception $e) {
    echo 'Exception when calling AIApi->sendMessage: ', $e->getMessage(), PHP_EOL;
}
```

### Parameters

| Name | Type | Description  | Notes |
| ------------- | ------------- | ------------- | ------------- |
| **project_id** | **string**|  | |
| **session_id** | **string**|  | |
| **message_input** | [**\Riders\Pulse\Model\MessageInput**](../Model/MessageInput.md)|  | |

### Return type

[**\Riders\Pulse\Model\Session**](../Model/Session.md)

### Authorization

[bearerAuth](../../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: `application/json`
- **Accept**: `application/json`, `application/problem+json`

[[Back to top]](#) [[Back to API list]](../../README.md#endpoints)
[[Back to Model list]](../../README.md#models)
[[Back to README]](../../README.md)
