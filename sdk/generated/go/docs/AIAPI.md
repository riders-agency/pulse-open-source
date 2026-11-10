# \AIAPI

All URIs are relative to *https://pulse.example.test*

Method | HTTP request | Description
------------- | ------------- | -------------
[**ApproveProposal**](AIAPI.md#ApproveProposal) | **Post** /api/v1/projects/{project_id}/ai/tool-calls/{tool_call_id}/approve | approveProposal
[**CancelSession**](AIAPI.md#CancelSession) | **Post** /api/v1/projects/{project_id}/ai/employee/sessions/{session_id}/cancel | cancelSession
[**CreateSession**](AIAPI.md#CreateSession) | **Post** /api/v1/projects/{project_id}/ai/employee/sessions | createSession
[**EditProposal**](AIAPI.md#EditProposal) | **Patch** /api/v1/projects/{project_id}/ai/tool-calls/{tool_call_id} | editProposal
[**GetSession**](AIAPI.md#GetSession) | **Get** /api/v1/projects/{project_id}/ai/employee/sessions/{session_id} | getSession
[**RejectProposal**](AIAPI.md#RejectProposal) | **Post** /api/v1/projects/{project_id}/ai/tool-calls/{tool_call_id}/reject | rejectProposal
[**SendMessage**](AIAPI.md#SendMessage) | **Post** /api/v1/projects/{project_id}/ai/employee/sessions/{session_id}/messages | sendMessage



## ApproveProposal

> Session ApproveProposal(ctx, projectId, toolCallId).Execute()

approveProposal

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
	projectId := "projectId_example" // string |
	toolCallId := "toolCallId_example" // string |

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AIAPI.ApproveProposal(context.Background(), projectId, toolCallId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AIAPI.ApproveProposal``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ApproveProposal`: Session
	fmt.Fprintf(os.Stdout, "Response from `AIAPI.ApproveProposal`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**projectId** | **string** |  |
**toolCallId** | **string** |  |

### Other Parameters

Other parameters are passed through a pointer to a apiApproveProposalRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**Session**](Session.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CancelSession

> Session CancelSession(ctx, projectId, sessionId).Execute()

cancelSession

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
	projectId := "projectId_example" // string |
	sessionId := "sessionId_example" // string |

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AIAPI.CancelSession(context.Background(), projectId, sessionId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AIAPI.CancelSession``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CancelSession`: Session
	fmt.Fprintf(os.Stdout, "Response from `AIAPI.CancelSession`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**projectId** | **string** |  |
**sessionId** | **string** |  |

### Other Parameters

Other parameters are passed through a pointer to a apiCancelSessionRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**Session**](Session.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateSession

> Session CreateSession(ctx, projectId).SessionInput(sessionInput).Execute()

createSession

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
	projectId := "projectId_example" // string |
	sessionInput := *openapiclient.NewSessionInput("ScreenScope_example", "ScreenEntityType_example", "ScreenEntityId_example") // SessionInput |

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AIAPI.CreateSession(context.Background(), projectId).SessionInput(sessionInput).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AIAPI.CreateSession``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateSession`: Session
	fmt.Fprintf(os.Stdout, "Response from `AIAPI.CreateSession`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**projectId** | **string** |  |

### Other Parameters

Other parameters are passed through a pointer to a apiCreateSessionRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **sessionInput** | [**SessionInput**](SessionInput.md) |  |

### Return type

[**Session**](Session.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## EditProposal

> Session EditProposal(ctx, projectId, toolCallId).ProposalPatch(proposalPatch).Execute()

editProposal

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
	projectId := "projectId_example" // string |
	toolCallId := "toolCallId_example" // string |
	proposalPatch := *openapiclient.NewProposalPatch(map[string]interface{}{"key": interface{}(123)}) // ProposalPatch |

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AIAPI.EditProposal(context.Background(), projectId, toolCallId).ProposalPatch(proposalPatch).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AIAPI.EditProposal``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `EditProposal`: Session
	fmt.Fprintf(os.Stdout, "Response from `AIAPI.EditProposal`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**projectId** | **string** |  |
**toolCallId** | **string** |  |

### Other Parameters

Other parameters are passed through a pointer to a apiEditProposalRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **proposalPatch** | [**ProposalPatch**](ProposalPatch.md) |  |

### Return type

[**Session**](Session.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetSession

> Session GetSession(ctx, projectId, sessionId).Execute()

getSession

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
	projectId := "projectId_example" // string |
	sessionId := "sessionId_example" // string |

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AIAPI.GetSession(context.Background(), projectId, sessionId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AIAPI.GetSession``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetSession`: Session
	fmt.Fprintf(os.Stdout, "Response from `AIAPI.GetSession`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**projectId** | **string** |  |
**sessionId** | **string** |  |

### Other Parameters

Other parameters are passed through a pointer to a apiGetSessionRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**Session**](Session.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## RejectProposal

> Session RejectProposal(ctx, projectId, toolCallId).Rejection(rejection).Execute()

rejectProposal

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
	projectId := "projectId_example" // string |
	toolCallId := "toolCallId_example" // string |
	rejection := *openapiclient.NewRejection() // Rejection |

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AIAPI.RejectProposal(context.Background(), projectId, toolCallId).Rejection(rejection).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AIAPI.RejectProposal``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `RejectProposal`: Session
	fmt.Fprintf(os.Stdout, "Response from `AIAPI.RejectProposal`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**projectId** | **string** |  |
**toolCallId** | **string** |  |

### Other Parameters

Other parameters are passed through a pointer to a apiRejectProposalRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **rejection** | [**Rejection**](Rejection.md) |  |

### Return type

[**Session**](Session.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SendMessage

> Session SendMessage(ctx, projectId, sessionId).MessageInput(messageInput).Execute()

sendMessage

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
	projectId := "projectId_example" // string |
	sessionId := "sessionId_example" // string |
	messageInput := *openapiclient.NewMessageInput("Body_example") // MessageInput |

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AIAPI.SendMessage(context.Background(), projectId, sessionId).MessageInput(messageInput).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AIAPI.SendMessage``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SendMessage`: Session
	fmt.Fprintf(os.Stdout, "Response from `AIAPI.SendMessage`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**projectId** | **string** |  |
**sessionId** | **string** |  |

### Other Parameters

Other parameters are passed through a pointer to a apiSendMessageRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **messageInput** | [**MessageInput**](MessageInput.md) |  |

### Return type

[**Session**](Session.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)
