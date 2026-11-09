# SessionInput

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ScreenScope** | **string** |  |
**ScreenEntityType** | **string** |  |
**ScreenEntityId** | **string** |  |
**PageContextJson** | Pointer to **map[string]interface{}** |  | [optional]

## Methods

### NewSessionInput

`func NewSessionInput(screenScope string, screenEntityType string, screenEntityId string, ) *SessionInput`

NewSessionInput instantiates a new SessionInput object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSessionInputWithDefaults

`func NewSessionInputWithDefaults() *SessionInput`

NewSessionInputWithDefaults instantiates a new SessionInput object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetScreenScope

`func (o *SessionInput) GetScreenScope() string`

GetScreenScope returns the ScreenScope field if non-nil, zero value otherwise.

### GetScreenScopeOk

`func (o *SessionInput) GetScreenScopeOk() (*string, bool)`

GetScreenScopeOk returns a tuple with the ScreenScope field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScreenScope

`func (o *SessionInput) SetScreenScope(v string)`

SetScreenScope sets ScreenScope field to given value.


### GetScreenEntityType

`func (o *SessionInput) GetScreenEntityType() string`

GetScreenEntityType returns the ScreenEntityType field if non-nil, zero value otherwise.

### GetScreenEntityTypeOk

`func (o *SessionInput) GetScreenEntityTypeOk() (*string, bool)`

GetScreenEntityTypeOk returns a tuple with the ScreenEntityType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScreenEntityType

`func (o *SessionInput) SetScreenEntityType(v string)`

SetScreenEntityType sets ScreenEntityType field to given value.


### GetScreenEntityId

`func (o *SessionInput) GetScreenEntityId() string`

GetScreenEntityId returns the ScreenEntityId field if non-nil, zero value otherwise.

### GetScreenEntityIdOk

`func (o *SessionInput) GetScreenEntityIdOk() (*string, bool)`

GetScreenEntityIdOk returns a tuple with the ScreenEntityId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScreenEntityId

`func (o *SessionInput) SetScreenEntityId(v string)`

SetScreenEntityId sets ScreenEntityId field to given value.


### GetPageContextJson

`func (o *SessionInput) GetPageContextJson() map[string]interface{}`

GetPageContextJson returns the PageContextJson field if non-nil, zero value otherwise.

### GetPageContextJsonOk

`func (o *SessionInput) GetPageContextJsonOk() (*map[string]interface{}, bool)`

GetPageContextJsonOk returns a tuple with the PageContextJson field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPageContextJson

`func (o *SessionInput) SetPageContextJson(v map[string]interface{})`

SetPageContextJson sets PageContextJson field to given value.

### HasPageContextJson

`func (o *SessionInput) HasPageContextJson() bool`

HasPageContextJson returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
