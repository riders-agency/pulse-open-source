# \AiApi

All URIs are relative to *https://pulse.example.test*

Method | HTTP request | Description
------------- | ------------- | -------------
[**approve_proposal**](AiApi.md#approve_proposal) | **POST** /api/v1/projects/{project_id}/ai/tool-calls/{tool_call_id}/approve | approveProposal
[**cancel_session**](AiApi.md#cancel_session) | **POST** /api/v1/projects/{project_id}/ai/employee/sessions/{session_id}/cancel | cancelSession
[**create_session**](AiApi.md#create_session) | **POST** /api/v1/projects/{project_id}/ai/employee/sessions | createSession
[**edit_proposal**](AiApi.md#edit_proposal) | **PATCH** /api/v1/projects/{project_id}/ai/tool-calls/{tool_call_id} | editProposal
[**get_session**](AiApi.md#get_session) | **GET** /api/v1/projects/{project_id}/ai/employee/sessions/{session_id} | getSession
[**reject_proposal**](AiApi.md#reject_proposal) | **POST** /api/v1/projects/{project_id}/ai/tool-calls/{tool_call_id}/reject | rejectProposal
[**send_message**](AiApi.md#send_message) | **POST** /api/v1/projects/{project_id}/ai/employee/sessions/{session_id}/messages | sendMessage



## approve_proposal

> models::Session approve_proposal(project_id, tool_call_id)
approveProposal

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**project_id** | **String** |  | [required] |
**tool_call_id** | **String** |  | [required] |

### Return type

[**models::Session**](Session.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## cancel_session

> models::Session cancel_session(project_id, session_id)
cancelSession

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**project_id** | **String** |  | [required] |
**session_id** | **String** |  | [required] |

### Return type

[**models::Session**](Session.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## create_session

> models::Session create_session(project_id, session_input)
createSession

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**project_id** | **String** |  | [required] |
**session_input** | [**SessionInput**](SessionInput.md) |  | [required] |

### Return type

[**models::Session**](Session.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## edit_proposal

> models::Session edit_proposal(project_id, tool_call_id, proposal_patch)
editProposal

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**project_id** | **String** |  | [required] |
**tool_call_id** | **String** |  | [required] |
**proposal_patch** | [**ProposalPatch**](ProposalPatch.md) |  | [required] |

### Return type

[**models::Session**](Session.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## get_session

> models::Session get_session(project_id, session_id)
getSession

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**project_id** | **String** |  | [required] |
**session_id** | **String** |  | [required] |

### Return type

[**models::Session**](Session.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## reject_proposal

> models::Session reject_proposal(project_id, tool_call_id, rejection)
rejectProposal

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**project_id** | **String** |  | [required] |
**tool_call_id** | **String** |  | [required] |
**rejection** | [**Rejection**](Rejection.md) |  | [required] |

### Return type

[**models::Session**](Session.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## send_message

> models::Session send_message(project_id, session_id, message_input)
sendMessage

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**project_id** | **String** |  | [required] |
**session_id** | **String** |  | [required] |
**message_input** | [**MessageInput**](MessageInput.md) |  | [required] |

### Return type

[**models::Session**](Session.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)
