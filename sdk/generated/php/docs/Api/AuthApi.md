# Riders\Pulse\AuthApi



All URIs are relative to https://pulse.example.test, except if the operation defines another base path.

| Method | HTTP request | Description |
| ------------- | ------------- | ------------- |
| [**getIdentity()**](AuthApi.md#getIdentity) | **GET** /api/v1/auth/me | getIdentity |
| [**login()**](AuthApi.md#login) | **POST** /api/v1/auth/login | login |
| [**logout()**](AuthApi.md#logout) | **POST** /api/v1/auth/logout | logout |


## `getIdentity()`

```php
getIdentity(): \Riders\Pulse\Model\Identity
```

getIdentity

### Example

```php
<?php
require_once(__DIR__ . '/vendor/autoload.php');


// Configure Bearer authorization: bearerAuth
$config = Riders\Pulse\Configuration::getDefaultConfiguration()->setAccessToken('YOUR_ACCESS_TOKEN');


$apiInstance = new Riders\Pulse\Api\AuthApi(
    // If you want use custom http client, pass your client which implements `GuzzleHttp\ClientInterface`.
    // This is optional, `GuzzleHttp\Client` will be used as default.
    new GuzzleHttp\Client(),
    $config
);

try {
    $result = $apiInstance->getIdentity();
    print_r($result);
} catch (Exception $e) {
    echo 'Exception when calling AuthApi->getIdentity: ', $e->getMessage(), PHP_EOL;
}
```

### Parameters

This endpoint does not need any parameter.

### Return type

[**\Riders\Pulse\Model\Identity**](../Model/Identity.md)

### Authorization

[bearerAuth](../../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`, `application/problem+json`

[[Back to top]](#) [[Back to API list]](../../README.md#endpoints)
[[Back to Model list]](../../README.md#models)
[[Back to README]](../../README.md)

## `login()`

```php
login($login_input): \Riders\Pulse\Model\TokenPair
```

login

### Example

```php
<?php
require_once(__DIR__ . '/vendor/autoload.php');



$apiInstance = new Riders\Pulse\Api\AuthApi(
    // If you want use custom http client, pass your client which implements `GuzzleHttp\ClientInterface`.
    // This is optional, `GuzzleHttp\Client` will be used as default.
    new GuzzleHttp\Client()
);
$login_input = new \Riders\Pulse\Model\LoginInput(); // \Riders\Pulse\Model\LoginInput

try {
    $result = $apiInstance->login($login_input);
    print_r($result);
} catch (Exception $e) {
    echo 'Exception when calling AuthApi->login: ', $e->getMessage(), PHP_EOL;
}
```

### Parameters

| Name | Type | Description  | Notes |
| ------------- | ------------- | ------------- | ------------- |
| **login_input** | [**\Riders\Pulse\Model\LoginInput**](../Model/LoginInput.md)|  | |

### Return type

[**\Riders\Pulse\Model\TokenPair**](../Model/TokenPair.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: `application/json`
- **Accept**: `application/json`, `application/problem+json`

[[Back to top]](#) [[Back to API list]](../../README.md#endpoints)
[[Back to Model list]](../../README.md#models)
[[Back to README]](../../README.md)

## `logout()`

```php
logout()
```

logout

### Example

```php
<?php
require_once(__DIR__ . '/vendor/autoload.php');


// Configure Bearer authorization: bearerAuth
$config = Riders\Pulse\Configuration::getDefaultConfiguration()->setAccessToken('YOUR_ACCESS_TOKEN');


$apiInstance = new Riders\Pulse\Api\AuthApi(
    // If you want use custom http client, pass your client which implements `GuzzleHttp\ClientInterface`.
    // This is optional, `GuzzleHttp\Client` will be used as default.
    new GuzzleHttp\Client(),
    $config
);

try {
    $apiInstance->logout();
} catch (Exception $e) {
    echo 'Exception when calling AuthApi->logout: ', $e->getMessage(), PHP_EOL;
}
```

### Parameters

This endpoint does not need any parameter.

### Return type

void (empty response body)

### Authorization

[bearerAuth](../../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/problem+json`

[[Back to top]](#) [[Back to API list]](../../README.md#endpoints)
[[Back to Model list]](../../README.md#models)
[[Back to README]](../../README.md)
