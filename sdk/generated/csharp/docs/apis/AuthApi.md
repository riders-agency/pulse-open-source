# Riders.Pulse.Api.AuthApi

All URIs are relative to *https://pulse.example.test*

| Method | HTTP request | Description |
|--------|--------------|-------------|
| [**GetIdentity**](AuthApi.md#getidentity) | **GET** /api/v1/auth/me | getIdentity |
| [**Login**](AuthApi.md#login) | **POST** /api/v1/auth/login | login |
| [**Logout**](AuthApi.md#logout) | **POST** /api/v1/auth/logout | logout |

<a id="getidentity"></a>
# **GetIdentity**
> Identity GetIdentity ()

getIdentity


### Parameters
This endpoint does not need any parameter.
### Return type

[**Identity**](Identity.md)

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

<a id="login"></a>
# **Login**
> TokenPair Login (LoginInput loginInput)

login


### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **loginInput** | [**LoginInput**](LoginInput.md) |  |  |

### Return type

[**TokenPair**](TokenPair.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Success |  -  |
| **0** | API error |  -  |

[[Back to top]](#) [[Back to API list]](../../README.md#documentation-for-api-endpoints) [[Back to Model list]](../../README.md#documentation-for-models) [[Back to README]](../../README.md)

<a id="logout"></a>
# **Logout**
> void Logout ()

logout


### Parameters
This endpoint does not need any parameter.
### Return type

void (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/problem+json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **204** | Success |  -  |
| **0** | API error |  -  |

[[Back to top]](#) [[Back to API list]](../../README.md#documentation-for-api-endpoints) [[Back to Model list]](../../README.md#documentation-for-models) [[Back to README]](../../README.md)
