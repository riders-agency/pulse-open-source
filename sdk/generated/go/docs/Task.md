# Task

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**TaskId** | **string** |  |
**TaskKey** | Pointer to **string** |  | [optional]
**ProjectId** | Pointer to **string** |  | [optional]
**Title** | **string** |  |
**Description** | Pointer to **string** |  | [optional]
**Type** | Pointer to **string** |  | [optional]
**Priority** | Pointer to **string** |  | [optional]
**Severity** | Pointer to **string** |  | [optional]
**TaskWorkflowStatusId** | Pointer to **string** |  | [optional]
**DueDate** | Pointer to **string** |  | [optional]
**Revision** | Pointer to **int32** |  | [optional]
**Billable** | Pointer to **bool** |  | [optional]

## Methods

### NewTask

`func NewTask(taskId string, title string, ) *Task`

NewTask instantiates a new Task object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTaskWithDefaults

`func NewTaskWithDefaults() *Task`

NewTaskWithDefaults instantiates a new Task object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetTaskId

`func (o *Task) GetTaskId() string`

GetTaskId returns the TaskId field if non-nil, zero value otherwise.

### GetTaskIdOk

`func (o *Task) GetTaskIdOk() (*string, bool)`

GetTaskIdOk returns a tuple with the TaskId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTaskId

`func (o *Task) SetTaskId(v string)`

SetTaskId sets TaskId field to given value.


### GetTaskKey

`func (o *Task) GetTaskKey() string`

GetTaskKey returns the TaskKey field if non-nil, zero value otherwise.

### GetTaskKeyOk

`func (o *Task) GetTaskKeyOk() (*string, bool)`

GetTaskKeyOk returns a tuple with the TaskKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTaskKey

`func (o *Task) SetTaskKey(v string)`

SetTaskKey sets TaskKey field to given value.

### HasTaskKey

`func (o *Task) HasTaskKey() bool`

HasTaskKey returns a boolean if a field has been set.

### GetProjectId

`func (o *Task) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *Task) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *Task) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.

### HasProjectId

`func (o *Task) HasProjectId() bool`

HasProjectId returns a boolean if a field has been set.

### GetTitle

`func (o *Task) GetTitle() string`

GetTitle returns the Title field if non-nil, zero value otherwise.

### GetTitleOk

`func (o *Task) GetTitleOk() (*string, bool)`

GetTitleOk returns a tuple with the Title field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTitle

`func (o *Task) SetTitle(v string)`

SetTitle sets Title field to given value.


### GetDescription

`func (o *Task) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *Task) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *Task) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *Task) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetType

`func (o *Task) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *Task) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *Task) SetType(v string)`

SetType sets Type field to given value.

### HasType

`func (o *Task) HasType() bool`

HasType returns a boolean if a field has been set.

### GetPriority

`func (o *Task) GetPriority() string`

GetPriority returns the Priority field if non-nil, zero value otherwise.

### GetPriorityOk

`func (o *Task) GetPriorityOk() (*string, bool)`

GetPriorityOk returns a tuple with the Priority field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPriority

`func (o *Task) SetPriority(v string)`

SetPriority sets Priority field to given value.

### HasPriority

`func (o *Task) HasPriority() bool`

HasPriority returns a boolean if a field has been set.

### GetSeverity

`func (o *Task) GetSeverity() string`

GetSeverity returns the Severity field if non-nil, zero value otherwise.

### GetSeverityOk

`func (o *Task) GetSeverityOk() (*string, bool)`

GetSeverityOk returns a tuple with the Severity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSeverity

`func (o *Task) SetSeverity(v string)`

SetSeverity sets Severity field to given value.

### HasSeverity

`func (o *Task) HasSeverity() bool`

HasSeverity returns a boolean if a field has been set.

### GetTaskWorkflowStatusId

`func (o *Task) GetTaskWorkflowStatusId() string`

GetTaskWorkflowStatusId returns the TaskWorkflowStatusId field if non-nil, zero value otherwise.

### GetTaskWorkflowStatusIdOk

`func (o *Task) GetTaskWorkflowStatusIdOk() (*string, bool)`

GetTaskWorkflowStatusIdOk returns a tuple with the TaskWorkflowStatusId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTaskWorkflowStatusId

`func (o *Task) SetTaskWorkflowStatusId(v string)`

SetTaskWorkflowStatusId sets TaskWorkflowStatusId field to given value.

### HasTaskWorkflowStatusId

`func (o *Task) HasTaskWorkflowStatusId() bool`

HasTaskWorkflowStatusId returns a boolean if a field has been set.

### GetDueDate

`func (o *Task) GetDueDate() string`

GetDueDate returns the DueDate field if non-nil, zero value otherwise.

### GetDueDateOk

`func (o *Task) GetDueDateOk() (*string, bool)`

GetDueDateOk returns a tuple with the DueDate field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDueDate

`func (o *Task) SetDueDate(v string)`

SetDueDate sets DueDate field to given value.

### HasDueDate

`func (o *Task) HasDueDate() bool`

HasDueDate returns a boolean if a field has been set.

### GetRevision

`func (o *Task) GetRevision() int32`

GetRevision returns the Revision field if non-nil, zero value otherwise.

### GetRevisionOk

`func (o *Task) GetRevisionOk() (*int32, bool)`

GetRevisionOk returns a tuple with the Revision field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRevision

`func (o *Task) SetRevision(v int32)`

SetRevision sets Revision field to given value.

### HasRevision

`func (o *Task) HasRevision() bool`

HasRevision returns a boolean if a field has been set.

### GetBillable

`func (o *Task) GetBillable() bool`

GetBillable returns the Billable field if non-nil, zero value otherwise.

### GetBillableOk

`func (o *Task) GetBillableOk() (*bool, bool)`

GetBillableOk returns a tuple with the Billable field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBillable

`func (o *Task) SetBillable(v bool)`

SetBillable sets Billable field to given value.

### HasBillable

`func (o *Task) HasBillable() bool`

HasBillable returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
