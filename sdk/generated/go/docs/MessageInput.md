# MessageInput

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Body** | **string** |  |
**PageContextJson** | Pointer to **map[string]interface{}** |  | [optional]
**ClientContextJson** | Pointer to **map[string]interface{}** |  | [optional]
**ActiveToolCallId** | Pointer to **string** |  | [optional]
**LocalDate** | Pointer to **string** |  | [optional]
**Timezone** | Pointer to **string** |  | [optional]

## Methods

### NewMessageInput

`func NewMessageInput(body string, ) *MessageInput`

NewMessageInput instantiates a new MessageInput object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewMessageInputWithDefaults

`func NewMessageInputWithDefaults() *MessageInput`

NewMessageInputWithDefaults instantiates a new MessageInput object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBody

`func (o *MessageInput) GetBody() string`

GetBody returns the Body field if non-nil, zero value otherwise.

### GetBodyOk

`func (o *MessageInput) GetBodyOk() (*string, bool)`

GetBodyOk returns a tuple with the Body field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBody

`func (o *MessageInput) SetBody(v string)`

SetBody sets Body field to given value.


### GetPageContextJson

`func (o *MessageInput) GetPageContextJson() map[string]interface{}`

GetPageContextJson returns the PageContextJson field if non-nil, zero value otherwise.

### GetPageContextJsonOk

`func (o *MessageInput) GetPageContextJsonOk() (*map[string]interface{}, bool)`

GetPageContextJsonOk returns a tuple with the PageContextJson field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPageContextJson

`func (o *MessageInput) SetPageContextJson(v map[string]interface{})`

SetPageContextJson sets PageContextJson field to given value.

### HasPageContextJson

`func (o *MessageInput) HasPageContextJson() bool`

HasPageContextJson returns a boolean if a field has been set.

### GetClientContextJson

`func (o *MessageInput) GetClientContextJson() map[string]interface{}`

GetClientContextJson returns the ClientContextJson field if non-nil, zero value otherwise.

### GetClientContextJsonOk

`func (o *MessageInput) GetClientContextJsonOk() (*map[string]interface{}, bool)`

GetClientContextJsonOk returns a tuple with the ClientContextJson field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClientContextJson

`func (o *MessageInput) SetClientContextJson(v map[string]interface{})`

SetClientContextJson sets ClientContextJson field to given value.

### HasClientContextJson

`func (o *MessageInput) HasClientContextJson() bool`

HasClientContextJson returns a boolean if a field has been set.

### GetActiveToolCallId

`func (o *MessageInput) GetActiveToolCallId() string`

GetActiveToolCallId returns the ActiveToolCallId field if non-nil, zero value otherwise.

### GetActiveToolCallIdOk

`func (o *MessageInput) GetActiveToolCallIdOk() (*string, bool)`

GetActiveToolCallIdOk returns a tuple with the ActiveToolCallId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActiveToolCallId

`func (o *MessageInput) SetActiveToolCallId(v string)`

SetActiveToolCallId sets ActiveToolCallId field to given value.

### HasActiveToolCallId

`func (o *MessageInput) HasActiveToolCallId() bool`

HasActiveToolCallId returns a boolean if a field has been set.

### GetLocalDate

`func (o *MessageInput) GetLocalDate() string`

GetLocalDate returns the LocalDate field if non-nil, zero value otherwise.

### GetLocalDateOk

`func (o *MessageInput) GetLocalDateOk() (*string, bool)`

GetLocalDateOk returns a tuple with the LocalDate field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLocalDate

`func (o *MessageInput) SetLocalDate(v string)`

SetLocalDate sets LocalDate field to given value.

### HasLocalDate

`func (o *MessageInput) HasLocalDate() bool`

HasLocalDate returns a boolean if a field has been set.

### GetTimezone

`func (o *MessageInput) GetTimezone() string`

GetTimezone returns the Timezone field if non-nil, zero value otherwise.

### GetTimezoneOk

`func (o *MessageInput) GetTimezoneOk() (*string, bool)`

GetTimezoneOk returns a tuple with the Timezone field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimezone

`func (o *MessageInput) SetTimezone(v string)`

SetTimezone sets Timezone field to given value.

### HasTimezone

`func (o *MessageInput) HasTimezone() bool`

HasTimezone returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
