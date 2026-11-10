# AIAPI

All URIs are relative to *https://pulse.example.test*

Method | HTTP request | Description
------------- | ------------- | -------------
[**approveProposal**](AIAPI.md#approveproposal) | **POST** /api/v1/projects/{project_id}/ai/tool-calls/{tool_call_id}/approve | approveProposal
[**cancelSession**](AIAPI.md#cancelsession) | **POST** /api/v1/projects/{project_id}/ai/employee/sessions/{session_id}/cancel | cancelSession
[**createSession**](AIAPI.md#createsession) | **POST** /api/v1/projects/{project_id}/ai/employee/sessions | createSession
[**editProposal**](AIAPI.md#editproposal) | **PATCH** /api/v1/projects/{project_id}/ai/tool-calls/{tool_call_id} | editProposal
[**getSession**](AIAPI.md#getsession) | **GET** /api/v1/projects/{project_id}/ai/employee/sessions/{session_id} | getSession
[**rejectProposal**](AIAPI.md#rejectproposal) | **POST** /api/v1/projects/{project_id}/ai/tool-calls/{tool_call_id}/reject | rejectProposal
[**sendMessage**](AIAPI.md#sendmessage) | **POST** /api/v1/projects/{project_id}/ai/employee/sessions/{session_id}/messages | sendMessage


# **approveProposal**
```swift
    open class func approveProposal(projectId: String, toolCallId: String, completion: @escaping (_ data: Session?, _ error: Error?) -> Void)
```

approveProposal

### Example
```swift
// The following code samples are still beta. For any issue, please report via http://github.com/OpenAPITools/openapi-generator/issues/new
import PulseSDK

let projectId = "projectId_example" // String |
let toolCallId = "toolCallId_example" // String |

// approveProposal
AIAPI.approveProposal(projectId: projectId, toolCallId: toolCallId) { (response, error) in
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
 **projectId** | **String** |  |
 **toolCallId** | **String** |  |

### Return type

[**Session**](Session.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **cancelSession**
```swift
    open class func cancelSession(projectId: String, sessionId: String, completion: @escaping (_ data: Session?, _ error: Error?) -> Void)
```

cancelSession

### Example
```swift
// The following code samples are still beta. For any issue, please report via http://github.com/OpenAPITools/openapi-generator/issues/new
import PulseSDK

let projectId = "projectId_example" // String |
let sessionId = "sessionId_example" // String |

// cancelSession
AIAPI.cancelSession(projectId: projectId, sessionId: sessionId) { (response, error) in
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
 **projectId** | **String** |  |
 **sessionId** | **String** |  |

### Return type

[**Session**](Session.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **createSession**
```swift
    open class func createSession(projectId: String, sessionInput: SessionInput, completion: @escaping (_ data: Session?, _ error: Error?) -> Void)
```

createSession

### Example
```swift
// The following code samples are still beta. For any issue, please report via http://github.com/OpenAPITools/openapi-generator/issues/new
import PulseSDK

let projectId = "projectId_example" // String |
let sessionInput = SessionInput(screenScope: "screenScope_example", screenEntityType: "screenEntityType_example", screenEntityId: "screenEntityId_example", pageContextJson: "TODO") // SessionInput |

// createSession
AIAPI.createSession(projectId: projectId, sessionInput: sessionInput) { (response, error) in
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
 **projectId** | **String** |  |
 **sessionInput** | [**SessionInput**](SessionInput.md) |  |

### Return type

[**Session**](Session.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **editProposal**
```swift
    open class func editProposal(projectId: String, toolCallId: String, proposalPatch: ProposalPatch, completion: @escaping (_ data: Session?, _ error: Error?) -> Void)
```

editProposal

### Example
```swift
// The following code samples are still beta. For any issue, please report via http://github.com/OpenAPITools/openapi-generator/issues/new
import PulseSDK

let projectId = "projectId_example" // String |
let toolCallId = "toolCallId_example" // String |
let proposalPatch = ProposalPatch(inputJson: "TODO", editReason: "editReason_example") // ProposalPatch |

// editProposal
AIAPI.editProposal(projectId: projectId, toolCallId: toolCallId, proposalPatch: proposalPatch) { (response, error) in
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
 **projectId** | **String** |  |
 **toolCallId** | **String** |  |
 **proposalPatch** | [**ProposalPatch**](ProposalPatch.md) |  |

### Return type

[**Session**](Session.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getSession**
```swift
    open class func getSession(projectId: String, sessionId: String, completion: @escaping (_ data: Session?, _ error: Error?) -> Void)
```

getSession

### Example
```swift
// The following code samples are still beta. For any issue, please report via http://github.com/OpenAPITools/openapi-generator/issues/new
import PulseSDK

let projectId = "projectId_example" // String |
let sessionId = "sessionId_example" // String |

// getSession
AIAPI.getSession(projectId: projectId, sessionId: sessionId) { (response, error) in
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
 **projectId** | **String** |  |
 **sessionId** | **String** |  |

### Return type

[**Session**](Session.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **rejectProposal**
```swift
    open class func rejectProposal(projectId: String, toolCallId: String, rejection: Rejection, completion: @escaping (_ data: Session?, _ error: Error?) -> Void)
```

rejectProposal

### Example
```swift
// The following code samples are still beta. For any issue, please report via http://github.com/OpenAPITools/openapi-generator/issues/new
import PulseSDK

let projectId = "projectId_example" // String |
let toolCallId = "toolCallId_example" // String |
let rejection = Rejection(reason: "reason_example") // Rejection |

// rejectProposal
AIAPI.rejectProposal(projectId: projectId, toolCallId: toolCallId, rejection: rejection) { (response, error) in
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
 **projectId** | **String** |  |
 **toolCallId** | **String** |  |
 **rejection** | [**Rejection**](Rejection.md) |  |

### Return type

[**Session**](Session.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **sendMessage**
```swift
    open class func sendMessage(projectId: String, sessionId: String, messageInput: MessageInput, completion: @escaping (_ data: Session?, _ error: Error?) -> Void)
```

sendMessage

### Example
```swift
// The following code samples are still beta. For any issue, please report via http://github.com/OpenAPITools/openapi-generator/issues/new
import PulseSDK

let projectId = "projectId_example" // String |
let sessionId = "sessionId_example" // String |
let messageInput = MessageInput(body: "body_example", pageContextJson: "TODO", clientContextJson: "TODO", activeToolCallId: "activeToolCallId_example", localDate: "localDate_example", timezone: "timezone_example") // MessageInput |

// sendMessage
AIAPI.sendMessage(projectId: projectId, sessionId: sessionId, messageInput: messageInput) { (response, error) in
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
 **projectId** | **String** |  |
 **sessionId** | **String** |  |
 **messageInput** | [**MessageInput**](MessageInput.md) |  |

### Return type

[**Session**](Session.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)
