# pulse_sdk.api.AIApi

## Load the API package
```dart
import 'package:pulse_sdk/api.dart';
```

All URIs are relative to *https://pulse.example.test*

Method | HTTP request | Description
------------- | ------------- | -------------
[**approveProposal**](AIApi.md#approveproposal) | **POST** /api/v1/projects/{project_id}/ai/tool-calls/{tool_call_id}/approve | approveProposal
[**cancelSession**](AIApi.md#cancelsession) | **POST** /api/v1/projects/{project_id}/ai/employee/sessions/{session_id}/cancel | cancelSession
[**createSession**](AIApi.md#createsession) | **POST** /api/v1/projects/{project_id}/ai/employee/sessions | createSession
[**editProposal**](AIApi.md#editproposal) | **PATCH** /api/v1/projects/{project_id}/ai/tool-calls/{tool_call_id} | editProposal
[**getSession**](AIApi.md#getsession) | **GET** /api/v1/projects/{project_id}/ai/employee/sessions/{session_id} | getSession
[**rejectProposal**](AIApi.md#rejectproposal) | **POST** /api/v1/projects/{project_id}/ai/tool-calls/{tool_call_id}/reject | rejectProposal
[**sendMessage**](AIApi.md#sendmessage) | **POST** /api/v1/projects/{project_id}/ai/employee/sessions/{session_id}/messages | sendMessage


# **approveProposal**
> Session approveProposal(projectId, toolCallId)

approveProposal

### Example
```dart
import 'package:pulse_sdk/api.dart';
// TODO Configure HTTP Bearer authorization: bearerAuth
// Case 1. Use String Token
//defaultApiClient.getAuthentication<HttpBearerAuth>('bearerAuth').setAccessToken('YOUR_ACCESS_TOKEN');
// Case 2. Use Function which generate token.
// String yourTokenGeneratorFunction() { ... }
//defaultApiClient.getAuthentication<HttpBearerAuth>('bearerAuth').setAccessToken(yourTokenGeneratorFunction);

final api_instance = AIApi();
final projectId = projectId_example; // String |
final toolCallId = toolCallId_example; // String |

try {
    final result = api_instance.approveProposal(projectId, toolCallId);
    print(result);
} catch (e) {
    print('Exception when calling AIApi->approveProposal: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **projectId** | **String**|  |
 **toolCallId** | **String**|  |

### Return type

[**Session**](Session.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **cancelSession**
> Session cancelSession(projectId, sessionId)

cancelSession

### Example
```dart
import 'package:pulse_sdk/api.dart';
// TODO Configure HTTP Bearer authorization: bearerAuth
// Case 1. Use String Token
//defaultApiClient.getAuthentication<HttpBearerAuth>('bearerAuth').setAccessToken('YOUR_ACCESS_TOKEN');
// Case 2. Use Function which generate token.
// String yourTokenGeneratorFunction() { ... }
//defaultApiClient.getAuthentication<HttpBearerAuth>('bearerAuth').setAccessToken(yourTokenGeneratorFunction);

final api_instance = AIApi();
final projectId = projectId_example; // String |
final sessionId = sessionId_example; // String |

try {
    final result = api_instance.cancelSession(projectId, sessionId);
    print(result);
} catch (e) {
    print('Exception when calling AIApi->cancelSession: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **projectId** | **String**|  |
 **sessionId** | **String**|  |

### Return type

[**Session**](Session.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **createSession**
> Session createSession(projectId, sessionInput)

createSession

### Example
```dart
import 'package:pulse_sdk/api.dart';
// TODO Configure HTTP Bearer authorization: bearerAuth
// Case 1. Use String Token
//defaultApiClient.getAuthentication<HttpBearerAuth>('bearerAuth').setAccessToken('YOUR_ACCESS_TOKEN');
// Case 2. Use Function which generate token.
// String yourTokenGeneratorFunction() { ... }
//defaultApiClient.getAuthentication<HttpBearerAuth>('bearerAuth').setAccessToken(yourTokenGeneratorFunction);

final api_instance = AIApi();
final projectId = projectId_example; // String |
final sessionInput = SessionInput(); // SessionInput |

try {
    final result = api_instance.createSession(projectId, sessionInput);
    print(result);
} catch (e) {
    print('Exception when calling AIApi->createSession: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **projectId** | **String**|  |
 **sessionInput** | [**SessionInput**](SessionInput.md)|  |

### Return type

[**Session**](Session.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **editProposal**
> Session editProposal(projectId, toolCallId, proposalPatch)

editProposal

### Example
```dart
import 'package:pulse_sdk/api.dart';
// TODO Configure HTTP Bearer authorization: bearerAuth
// Case 1. Use String Token
//defaultApiClient.getAuthentication<HttpBearerAuth>('bearerAuth').setAccessToken('YOUR_ACCESS_TOKEN');
// Case 2. Use Function which generate token.
// String yourTokenGeneratorFunction() { ... }
//defaultApiClient.getAuthentication<HttpBearerAuth>('bearerAuth').setAccessToken(yourTokenGeneratorFunction);

final api_instance = AIApi();
final projectId = projectId_example; // String |
final toolCallId = toolCallId_example; // String |
final proposalPatch = ProposalPatch(); // ProposalPatch |

try {
    final result = api_instance.editProposal(projectId, toolCallId, proposalPatch);
    print(result);
} catch (e) {
    print('Exception when calling AIApi->editProposal: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **projectId** | **String**|  |
 **toolCallId** | **String**|  |
 **proposalPatch** | [**ProposalPatch**](ProposalPatch.md)|  |

### Return type

[**Session**](Session.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getSession**
> Session getSession(projectId, sessionId)

getSession

### Example
```dart
import 'package:pulse_sdk/api.dart';
// TODO Configure HTTP Bearer authorization: bearerAuth
// Case 1. Use String Token
//defaultApiClient.getAuthentication<HttpBearerAuth>('bearerAuth').setAccessToken('YOUR_ACCESS_TOKEN');
// Case 2. Use Function which generate token.
// String yourTokenGeneratorFunction() { ... }
//defaultApiClient.getAuthentication<HttpBearerAuth>('bearerAuth').setAccessToken(yourTokenGeneratorFunction);

final api_instance = AIApi();
final projectId = projectId_example; // String |
final sessionId = sessionId_example; // String |

try {
    final result = api_instance.getSession(projectId, sessionId);
    print(result);
} catch (e) {
    print('Exception when calling AIApi->getSession: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **projectId** | **String**|  |
 **sessionId** | **String**|  |

### Return type

[**Session**](Session.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **rejectProposal**
> Session rejectProposal(projectId, toolCallId, rejection)

rejectProposal

### Example
```dart
import 'package:pulse_sdk/api.dart';
// TODO Configure HTTP Bearer authorization: bearerAuth
// Case 1. Use String Token
//defaultApiClient.getAuthentication<HttpBearerAuth>('bearerAuth').setAccessToken('YOUR_ACCESS_TOKEN');
// Case 2. Use Function which generate token.
// String yourTokenGeneratorFunction() { ... }
//defaultApiClient.getAuthentication<HttpBearerAuth>('bearerAuth').setAccessToken(yourTokenGeneratorFunction);

final api_instance = AIApi();
final projectId = projectId_example; // String |
final toolCallId = toolCallId_example; // String |
final rejection = Rejection(); // Rejection |

try {
    final result = api_instance.rejectProposal(projectId, toolCallId, rejection);
    print(result);
} catch (e) {
    print('Exception when calling AIApi->rejectProposal: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **projectId** | **String**|  |
 **toolCallId** | **String**|  |
 **rejection** | [**Rejection**](Rejection.md)|  |

### Return type

[**Session**](Session.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **sendMessage**
> Session sendMessage(projectId, sessionId, messageInput)

sendMessage

### Example
```dart
import 'package:pulse_sdk/api.dart';
// TODO Configure HTTP Bearer authorization: bearerAuth
// Case 1. Use String Token
//defaultApiClient.getAuthentication<HttpBearerAuth>('bearerAuth').setAccessToken('YOUR_ACCESS_TOKEN');
// Case 2. Use Function which generate token.
// String yourTokenGeneratorFunction() { ... }
//defaultApiClient.getAuthentication<HttpBearerAuth>('bearerAuth').setAccessToken(yourTokenGeneratorFunction);

final api_instance = AIApi();
final projectId = projectId_example; // String |
final sessionId = sessionId_example; // String |
final messageInput = MessageInput(); // MessageInput |

try {
    final result = api_instance.sendMessage(projectId, sessionId, messageInput);
    print(result);
} catch (e) {
    print('Exception when calling AIApi->sendMessage: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **projectId** | **String**|  |
 **sessionId** | **String**|  |
 **messageInput** | [**MessageInput**](MessageInput.md)|  |

### Return type

[**Session**](Session.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)
