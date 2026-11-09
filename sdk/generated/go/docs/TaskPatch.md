# TaskPatch

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Title** | Pointer to **string** |  | [optional]
**Description** | Pointer to **string** |  | [optional]
**Priority** | Pointer to **string** |  | [optional]
**DueDate** | Pointer to **string** |  | [optional]
**Billable** | Pointer to **bool** |  | [optional]
**ExpectedRevision** | Pointer to **int32** |  | [optional]

## Methods

### NewTaskPatch

`func NewTaskPatch() *TaskPatch`

NewTaskPatch instantiates a new TaskPatch object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTaskPatchWithDefaults

`func NewTaskPatchWithDefaults() *TaskPatch`

NewTaskPatchWithDefaults instantiates a new TaskPatch object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetTitle

`func (o *TaskPatch) GetTitle() string`

GetTitle returns the Title field if non-nil, zero value otherwise.

### GetTitleOk

`func (o *TaskPatch) GetTitleOk() (*string, bool)`

GetTitleOk returns a tuple with the Title field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTitle

`func (o *TaskPatch) SetTitle(v string)`

SetTitle sets Title field to given value.

### HasTitle

`func (o *TaskPatch) HasTitle() bool`

HasTitle returns a boolean if a field has been set.

### GetDescription

`func (o *TaskPatch) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *TaskPatch) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *TaskPatch) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *TaskPatch) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetPriority

`func (o *TaskPatch) GetPriority() string`

GetPriority returns the Priority field if non-nil, zero value otherwise.

### GetPriorityOk

`func (o *TaskPatch) GetPriorityOk() (*string, bool)`

GetPriorityOk returns a tuple with the Priority field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPriority

`func (o *TaskPatch) SetPriority(v string)`

SetPriority sets Priority field to given value.

### HasPriority

`func (o *TaskPatch) HasPriority() bool`

HasPriority returns a boolean if a field has been set.

### GetDueDate

`func (o *TaskPatch) GetDueDate() string`

GetDueDate returns the DueDate field if non-nil, zero value otherwise.

### GetDueDateOk

`func (o *TaskPatch) GetDueDateOk() (*string, bool)`

GetDueDateOk returns a tuple with the DueDate field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDueDate

`func (o *TaskPatch) SetDueDate(v string)`

SetDueDate sets DueDate field to given value.

### HasDueDate

`func (o *TaskPatch) HasDueDate() bool`

HasDueDate returns a boolean if a field has been set.

### GetBillable

`func (o *TaskPatch) GetBillable() bool`

GetBillable returns the Billable field if non-nil, zero value otherwise.

### GetBillableOk

`func (o *TaskPatch) GetBillableOk() (*bool, bool)`

GetBillableOk returns a tuple with the Billable field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBillable

`func (o *TaskPatch) SetBillable(v bool)`

SetBillable sets Billable field to given value.

### HasBillable

`func (o *TaskPatch) HasBillable() bool`

HasBillable returns a boolean if a field has been set.

### GetExpectedRevision

`func (o *TaskPatch) GetExpectedRevision() int32`

GetExpectedRevision returns the ExpectedRevision field if non-nil, zero value otherwise.

### GetExpectedRevisionOk

`func (o *TaskPatch) GetExpectedRevisionOk() (*int32, bool)`

GetExpectedRevisionOk returns a tuple with the ExpectedRevision field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExpectedRevision

`func (o *TaskPatch) SetExpectedRevision(v int32)`

SetExpectedRevision sets ExpectedRevision field to given value.

### HasExpectedRevision

`func (o *TaskPatch) HasExpectedRevision() bool`

HasExpectedRevision returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
