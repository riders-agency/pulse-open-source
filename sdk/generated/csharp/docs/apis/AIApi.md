# Riders.Pulse.Api.AIApi

All URIs are relative to *https://pulse.example.test*

| Method | HTTP request | Description |
|--------|--------------|-------------|
| [**ApproveProposal**](AIApi.md#approveproposal) | **POST** /api/v1/projects/{project_id}/ai/tool-calls/{tool_call_id}/approve | approveProposal |
| [**CancelSession**](AIApi.md#cancelsession) | **POST** /api/v1/projects/{project_id}/ai/employee/sessions/{session_id}/cancel | cancelSession |
| [**CreateSession**](AIApi.md#createsession) | **POST** /api/v1/projects/{project_id}/ai/employee/sessions | createSession |
| [**EditProposal**](AIApi.md#editproposal) | **PATCH** /api/v1/projects/{project_id}/ai/tool-calls/{tool_call_id} | editProposal |
| [**GetSession**](AIApi.md#getsession) | **GET** /api/v1/projects/{project_id}/ai/employee/sessions/{session_id} | getSession |
| [**RejectProposal**](AIApi.md#rejectproposal) | **POST** /api/v1/projects/{project_id}/ai/tool-calls/{tool_call_id}/reject | rejectProposal |
| [**SendMessage**](AIApi.md#sendmessage) | **POST** /api/v1/projects/{project_id}/ai/employee/sessions/{session_id}/messages | sendMessage |

<a id="approveproposal"></a>
# **ApproveProposal**
> Session ApproveProposal (string projectId, string toolCallId)

approveProposal


### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **projectId** | **string** |  |  |
| **toolCallId** | **string** |  |  |

### Return type

[**Session**](Session.md)

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

<a id="cancelsession"></a>
# **CancelSession**
> Session CancelSession (string projectId, string sessionId)

cancelSession


### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **projectId** | **string** |  |  |
| **sessionId** | **string** |  |  |

### Return type

[**Session**](Session.md)

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

<a id="createsession"></a>
# **CreateSession**
> Session CreateSession (string projectId, SessionInput sessionInput)

createSession


### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **projectId** | **string** |  |  |
| **sessionInput** | [**SessionInput**](SessionInput.md) |  |  |

### Return type

[**Session**](Session.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **201** | Success |  -  |
| **0** | API error |  -  |

[[Back to top]](#) [[Back to API list]](../../README.md#documentation-for-api-endpoints) [[Back to Model list]](../../README.md#documentation-for-models) [[Back to README]](../../README.md)

<a id="editproposal"></a>
# **EditProposal**
> Session EditProposal (string projectId, string toolCallId, ProposalPatch proposalPatch)

editProposal


### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **projectId** | **string** |  |  |
| **toolCallId** | **string** |  |  |
| **proposalPatch** | [**ProposalPatch**](ProposalPatch.md) |  |  |

### Return type

[**Session**](Session.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Success |  -  |
| **0** | API error |  -  |

[[Back to top]](#) [[Back to API list]](../../README.md#documentation-for-api-endpoints) [[Back to Model list]](../../README.md#documentation-for-models) [[Back to README]](../../README.md)

<a id="getsession"></a>
# **GetSession**
> Session GetSession (string projectId, string sessionId)

getSession


### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **projectId** | **string** |  |  |
| **sessionId** | **string** |  |  |

### Return type

[**Session**](Session.md)

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

<a id="rejectproposal"></a>
# **RejectProposal**
> Session RejectProposal (string projectId, string toolCallId, Rejection rejection)

rejectProposal


### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **projectId** | **string** |  |  |
| **toolCallId** | **string** |  |  |
| **rejection** | [**Rejection**](Rejection.md) |  |  |

### Return type

[**Session**](Session.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Success |  -  |
| **0** | API error |  -  |

[[Back to top]](#) [[Back to API list]](../../README.md#documentation-for-api-endpoints) [[Back to Model list]](../../README.md#documentation-for-models) [[Back to README]](../../README.md)

<a id="sendmessage"></a>
# **SendMessage**
> Session SendMessage (string projectId, string sessionId, MessageInput messageInput)

sendMessage


### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **projectId** | **string** |  |  |
| **sessionId** | **string** |  |  |
| **messageInput** | [**MessageInput**](MessageInput.md) |  |  |

### Return type

[**Session**](Session.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Success |  -  |
| **0** | API error |  -  |

[[Back to top]](#) [[Back to API list]](../../README.md#documentation-for-api-endpoints) [[Back to Model list]](../../README.md#documentation-for-models) [[Back to README]](../../README.md)
