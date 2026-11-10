# ToolCall

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AiToolCallId** | **string** |  |
**ToolKey** | **string** |  |
**Status** | **string** |  |
**RequiresApproval** | Pointer to **bool** |  | [optional]
**InputJson** | Pointer to **map[string]interface{}** |  | [optional]
**OutputJson** | Pointer to **map[string]interface{}** |  | [optional]

## Methods

### NewToolCall

`func NewToolCall(aiToolCallId string, toolKey string, status string, ) *ToolCall`

NewToolCall instantiates a new ToolCall object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewToolCallWithDefaults

`func NewToolCallWithDefaults() *ToolCall`

NewToolCallWithDefaults instantiates a new ToolCall object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAiToolCallId

`func (o *ToolCall) GetAiToolCallId() string`

GetAiToolCallId returns the AiToolCallId field if non-nil, zero value otherwise.

### GetAiToolCallIdOk

`func (o *ToolCall) GetAiToolCallIdOk() (*string, bool)`

GetAiToolCallIdOk returns a tuple with the AiToolCallId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAiToolCallId

`func (o *ToolCall) SetAiToolCallId(v string)`

SetAiToolCallId sets AiToolCallId field to given value.


### GetToolKey

`func (o *ToolCall) GetToolKey() string`

GetToolKey returns the ToolKey field if non-nil, zero value otherwise.

### GetToolKeyOk

`func (o *ToolCall) GetToolKeyOk() (*string, bool)`

GetToolKeyOk returns a tuple with the ToolKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetToolKey

`func (o *ToolCall) SetToolKey(v string)`

SetToolKey sets ToolKey field to given value.


### GetStatus

`func (o *ToolCall) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *ToolCall) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *ToolCall) SetStatus(v string)`

SetStatus sets Status field to given value.


### GetRequiresApproval

`func (o *ToolCall) GetRequiresApproval() bool`

GetRequiresApproval returns the RequiresApproval field if non-nil, zero value otherwise.

### GetRequiresApprovalOk

`func (o *ToolCall) GetRequiresApprovalOk() (*bool, bool)`

GetRequiresApprovalOk returns a tuple with the RequiresApproval field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequiresApproval

`func (o *ToolCall) SetRequiresApproval(v bool)`

SetRequiresApproval sets RequiresApproval field to given value.

### HasRequiresApproval

`func (o *ToolCall) HasRequiresApproval() bool`

HasRequiresApproval returns a boolean if a field has been set.

### GetInputJson

`func (o *ToolCall) GetInputJson() map[string]interface{}`

GetInputJson returns the InputJson field if non-nil, zero value otherwise.

### GetInputJsonOk

`func (o *ToolCall) GetInputJsonOk() (*map[string]interface{}, bool)`

GetInputJsonOk returns a tuple with the InputJson field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInputJson

`func (o *ToolCall) SetInputJson(v map[string]interface{})`

SetInputJson sets InputJson field to given value.

### HasInputJson

`func (o *ToolCall) HasInputJson() bool`

HasInputJson returns a boolean if a field has been set.

### GetOutputJson

`func (o *ToolCall) GetOutputJson() map[string]interface{}`

GetOutputJson returns the OutputJson field if non-nil, zero value otherwise.

### GetOutputJsonOk

`func (o *ToolCall) GetOutputJsonOk() (*map[string]interface{}, bool)`

GetOutputJsonOk returns a tuple with the OutputJson field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOutputJson

`func (o *ToolCall) SetOutputJson(v map[string]interface{})`

SetOutputJson sets OutputJson field to given value.

### HasOutputJson

`func (o *ToolCall) HasOutputJson() bool`

HasOutputJson returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
