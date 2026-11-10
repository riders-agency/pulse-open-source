# \AuthAPI

All URIs are relative to *https://pulse.example.test*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetIdentity**](AuthAPI.md#GetIdentity) | **Get** /api/v1/auth/me | getIdentity
[**Login**](AuthAPI.md#Login) | **Post** /api/v1/auth/login | login
[**Logout**](AuthAPI.md#Logout) | **Post** /api/v1/auth/logout | logout



## GetIdentity

> Identity GetIdentity(ctx).Execute()

getIdentity

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

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AuthAPI.GetIdentity(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AuthAPI.GetIdentity``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetIdentity`: Identity
	fmt.Fprintf(os.Stdout, "Response from `AuthAPI.GetIdentity`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetIdentityRequest struct via the builder pattern


### Return type

[**Identity**](Identity.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## Login

> TokenPair Login(ctx).LoginInput(loginInput).Execute()

login

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
	loginInput := *openapiclient.NewLoginInput("Email_example", "Password_example") // LoginInput |

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AuthAPI.Login(context.Background()).LoginInput(loginInput).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AuthAPI.Login``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `Login`: TokenPair
	fmt.Fprintf(os.Stdout, "Response from `AuthAPI.Login`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiLoginRequest struct via the builder pattern


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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## Logout

> Logout(ctx).Execute()

logout

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

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.AuthAPI.Logout(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AuthAPI.Logout``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiLogoutRequest struct via the builder pattern


### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)
