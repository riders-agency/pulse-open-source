# TaskInput

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ProjectId** | **string** |  |
**Title** | **string** |  |
**Description** | Pointer to **string** |  | [optional]
**Type** | Pointer to **string** |  | [optional]
**Priority** | Pointer to **string** |  | [optional]
**Severity** | Pointer to **string** |  | [optional]
**TaskWorkflowStatusId** | Pointer to **string** |  | [optional]
**ClientId** | Pointer to **string** |  | [optional]
**DueDate** | Pointer to **string** |  | [optional]
**Billable** | Pointer to **bool** |  | [optional]

## Methods

### NewTaskInput

`func NewTaskInput(projectId string, title string, ) *TaskInput`

NewTaskInput instantiates a new TaskInput object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTaskInputWithDefaults

`func NewTaskInputWithDefaults() *TaskInput`

NewTaskInputWithDefaults instantiates a new TaskInput object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetProjectId

`func (o *TaskInput) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *TaskInput) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *TaskInput) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.


### GetTitle

`func (o *TaskInput) GetTitle() string`

GetTitle returns the Title field if non-nil, zero value otherwise.

### GetTitleOk

`func (o *TaskInput) GetTitleOk() (*string, bool)`

GetTitleOk returns a tuple with the Title field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTitle

`func (o *TaskInput) SetTitle(v string)`

SetTitle sets Title field to given value.


### GetDescription

`func (o *TaskInput) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *TaskInput) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *TaskInput) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *TaskInput) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetType

`func (o *TaskInput) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *TaskInput) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *TaskInput) SetType(v string)`

SetType sets Type field to given value.

### HasType

`func (o *TaskInput) HasType() bool`

HasType returns a boolean if a field has been set.

### GetPriority

`func (o *TaskInput) GetPriority() string`

GetPriority returns the Priority field if non-nil, zero value otherwise.

### GetPriorityOk

`func (o *TaskInput) GetPriorityOk() (*string, bool)`

GetPriorityOk returns a tuple with the Priority field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPriority

`func (o *TaskInput) SetPriority(v string)`

SetPriority sets Priority field to given value.

### HasPriority

`func (o *TaskInput) HasPriority() bool`

HasPriority returns a boolean if a field has been set.

### GetSeverity

`func (o *TaskInput) GetSeverity() string`

GetSeverity returns the Severity field if non-nil, zero value otherwise.

### GetSeverityOk

`func (o *TaskInput) GetSeverityOk() (*string, bool)`

GetSeverityOk returns a tuple with the Severity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSeverity

`func (o *TaskInput) SetSeverity(v string)`

SetSeverity sets Severity field to given value.

### HasSeverity

`func (o *TaskInput) HasSeverity() bool`

HasSeverity returns a boolean if a field has been set.

### GetTaskWorkflowStatusId

`func (o *TaskInput) GetTaskWorkflowStatusId() string`

GetTaskWorkflowStatusId returns the TaskWorkflowStatusId field if non-nil, zero value otherwise.

### GetTaskWorkflowStatusIdOk

`func (o *TaskInput) GetTaskWorkflowStatusIdOk() (*string, bool)`

GetTaskWorkflowStatusIdOk returns a tuple with the TaskWorkflowStatusId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTaskWorkflowStatusId

`func (o *TaskInput) SetTaskWorkflowStatusId(v string)`

SetTaskWorkflowStatusId sets TaskWorkflowStatusId field to given value.

### HasTaskWorkflowStatusId

`func (o *TaskInput) HasTaskWorkflowStatusId() bool`

HasTaskWorkflowStatusId returns a boolean if a field has been set.

### GetClientId

`func (o *TaskInput) GetClientId() string`

GetClientId returns the ClientId field if non-nil, zero value otherwise.

### GetClientIdOk

`func (o *TaskInput) GetClientIdOk() (*string, bool)`

GetClientIdOk returns a tuple with the ClientId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClientId

`func (o *TaskInput) SetClientId(v string)`

SetClientId sets ClientId field to given value.

### HasClientId

`func (o *TaskInput) HasClientId() bool`

HasClientId returns a boolean if a field has been set.

### GetDueDate

`func (o *TaskInput) GetDueDate() string`

GetDueDate returns the DueDate field if non-nil, zero value otherwise.

### GetDueDateOk

`func (o *TaskInput) GetDueDateOk() (*string, bool)`

GetDueDateOk returns a tuple with the DueDate field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDueDate

`func (o *TaskInput) SetDueDate(v string)`

SetDueDate sets DueDate field to given value.

### HasDueDate

`func (o *TaskInput) HasDueDate() bool`

HasDueDate returns a boolean if a field has been set.

### GetBillable

`func (o *TaskInput) GetBillable() bool`

GetBillable returns the Billable field if non-nil, zero value otherwise.

### GetBillableOk

`func (o *TaskInput) GetBillableOk() (*bool, bool)`

GetBillableOk returns a tuple with the Billable field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBillable

`func (o *TaskInput) SetBillable(v bool)`

SetBillable sets Billable field to given value.

### HasBillable

`func (o *TaskInput) HasBillable() bool`

HasBillable returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
