# Session

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AiChatSessionId** | **string** |  |
**ProjectId** | **string** |  |
**Status** | **string** |  |
**Messages** | [**[]Message**](Message.md) |  |
**ToolCalls** | Pointer to [**[]ToolCall**](ToolCall.md) |  | [optional]

## Methods

### NewSession

`func NewSession(aiChatSessionId string, projectId string, status string, messages []Message, ) *Session`

NewSession instantiates a new Session object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSessionWithDefaults

`func NewSessionWithDefaults() *Session`

NewSessionWithDefaults instantiates a new Session object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAiChatSessionId

`func (o *Session) GetAiChatSessionId() string`

GetAiChatSessionId returns the AiChatSessionId field if non-nil, zero value otherwise.

### GetAiChatSessionIdOk

`func (o *Session) GetAiChatSessionIdOk() (*string, bool)`

GetAiChatSessionIdOk returns a tuple with the AiChatSessionId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAiChatSessionId

`func (o *Session) SetAiChatSessionId(v string)`

SetAiChatSessionId sets AiChatSessionId field to given value.


### GetProjectId

`func (o *Session) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *Session) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *Session) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.


### GetStatus

`func (o *Session) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *Session) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *Session) SetStatus(v string)`

SetStatus sets Status field to given value.


### GetMessages

`func (o *Session) GetMessages() []Message`

GetMessages returns the Messages field if non-nil, zero value otherwise.

### GetMessagesOk

`func (o *Session) GetMessagesOk() (*[]Message, bool)`

GetMessagesOk returns a tuple with the Messages field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessages

`func (o *Session) SetMessages(v []Message)`

SetMessages sets Messages field to given value.


### GetToolCalls

`func (o *Session) GetToolCalls() []ToolCall`

GetToolCalls returns the ToolCalls field if non-nil, zero value otherwise.

### GetToolCallsOk

`func (o *Session) GetToolCallsOk() (*[]ToolCall, bool)`

GetToolCallsOk returns a tuple with the ToolCalls field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetToolCalls

`func (o *Session) SetToolCalls(v []ToolCall)`

SetToolCalls sets ToolCalls field to given value.

### HasToolCalls

`func (o *Session) HasToolCalls() bool`

HasToolCalls returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
