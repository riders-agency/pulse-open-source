# AuthAPI

All URIs are relative to *https://pulse.example.test*

Method | HTTP request | Description
------------- | ------------- | -------------
[**getIdentity**](AuthAPI.md#getidentity) | **GET** /api/v1/auth/me | getIdentity
[**login**](AuthAPI.md#login) | **POST** /api/v1/auth/login | login
[**logout**](AuthAPI.md#logout) | **POST** /api/v1/auth/logout | logout


# **getIdentity**
```swift
    open class func getIdentity(completion: @escaping (_ data: Identity?, _ error: Error?) -> Void)
```

getIdentity

### Example
```swift
// The following code samples are still beta. For any issue, please report via http://github.com/OpenAPITools/openapi-generator/issues/new
import PulseSDK


// getIdentity
AuthAPI.getIdentity() { (response, error) in
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
```swift
    open class func login(loginInput: LoginInput, completion: @escaping (_ data: TokenPair?, _ error: Error?) -> Void)
```

login

### Example
```swift
// The following code samples are still beta. For any issue, please report via http://github.com/OpenAPITools/openapi-generator/issues/new
import PulseSDK

let loginInput = LoginInput(email: "email_example", password: "password_example") // LoginInput |

// login
AuthAPI.login(loginInput: loginInput) { (response, error) in
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
 **loginInput** | [**LoginInput**](LoginInput.md) |  |

### Return type

[**TokenPair**](TokenPair.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **logout**
```swift
    open class func logout(completion: @escaping (_ data: Void?, _ error: Error?) -> Void)
```

logout

### Example
```swift
// The following code samples are still beta. For any issue, please report via http://github.com/OpenAPITools/openapi-generator/issues/new
import PulseSDK


// logout
AuthAPI.logout() { (response, error) in
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
This endpoint does not need any parameter.

### Return type

Void (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)
