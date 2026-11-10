# pulse_sdk.api.AuthApi

## Load the API package
```dart
import 'package:pulse_sdk/api.dart';
```

All URIs are relative to *https://pulse.example.test*

Method | HTTP request | Description
------------- | ------------- | -------------
[**getIdentity**](AuthApi.md#getidentity) | **GET** /api/v1/auth/me | getIdentity
[**login**](AuthApi.md#login) | **POST** /api/v1/auth/login | login
[**logout**](AuthApi.md#logout) | **POST** /api/v1/auth/logout | logout


# **getIdentity**
> Identity getIdentity()

getIdentity

### Example
```dart
import 'package:pulse_sdk/api.dart';
// TODO Configure HTTP Bearer authorization: bearerAuth
// Case 1. Use String Token
//defaultApiClient.getAuthentication<HttpBearerAuth>('bearerAuth').setAccessToken('YOUR_ACCESS_TOKEN');
// Case 2. Use Function which generate token.
// String yourTokenGeneratorFunction() { ... }
//defaultApiClient.getAuthentication<HttpBearerAuth>('bearerAuth').setAccessToken(yourTokenGeneratorFunction);

final api_instance = AuthApi();

try {
    final result = api_instance.getIdentity();
    print(result);
} catch (e) {
    print('Exception when calling AuthApi->getIdentity: $e\n');
}
```

### Parameters
This endpoint does not need any parameter.

### Return type

[**Identity**](Identity.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **login**
> TokenPair login(loginInput)

login

### Example
```dart
import 'package:pulse_sdk/api.dart';

final api_instance = AuthApi();
final loginInput = LoginInput(); // LoginInput |

try {
    final result = api_instance.login(loginInput);
    print(result);
} catch (e) {
    print('Exception when calling AuthApi->login: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **loginInput** | [**LoginInput**](LoginInput.md)|  |

### Return type

[**TokenPair**](TokenPair.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **logout**
> logout()

logout

### Example
```dart
import 'package:pulse_sdk/api.dart';
// TODO Configure HTTP Bearer authorization: bearerAuth
// Case 1. Use String Token
//defaultApiClient.getAuthentication<HttpBearerAuth>('bearerAuth').setAccessToken('YOUR_ACCESS_TOKEN');
// Case 2. Use Function which generate token.
// String yourTokenGeneratorFunction() { ... }
//defaultApiClient.getAuthentication<HttpBearerAuth>('bearerAuth').setAccessToken(yourTokenGeneratorFunction);

final api_instance = AuthApi();

try {
    api_instance.logout();
} catch (e) {
    print('Exception when calling AuthApi->logout: $e\n');
}
```

### Parameters
This endpoint does not need any parameter.

### Return type

void (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)
