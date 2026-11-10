# TaskList

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Items** | [**[]Task**](Task.md) |  |
**Total** | **int32** |  |
**Limit** | Pointer to **int32** |  | [optional]
**Offset** | Pointer to **int32** |  | [optional]

## Methods

### NewTaskList

`func NewTaskList(items []Task, total int32, ) *TaskList`

NewTaskList instantiates a new TaskList object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTaskListWithDefaults

`func NewTaskListWithDefaults() *TaskList`

NewTaskListWithDefaults instantiates a new TaskList object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetItems

`func (o *TaskList) GetItems() []Task`

GetItems returns the Items field if non-nil, zero value otherwise.

### GetItemsOk

`func (o *TaskList) GetItemsOk() (*[]Task, bool)`

GetItemsOk returns a tuple with the Items field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetItems

`func (o *TaskList) SetItems(v []Task)`

SetItems sets Items field to given value.


### GetTotal

`func (o *TaskList) GetTotal() int32`

GetTotal returns the Total field if non-nil, zero value otherwise.

### GetTotalOk

`func (o *TaskList) GetTotalOk() (*int32, bool)`

GetTotalOk returns a tuple with the Total field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotal

`func (o *TaskList) SetTotal(v int32)`

SetTotal sets Total field to given value.


### GetLimit

`func (o *TaskList) GetLimit() int32`

GetLimit returns the Limit field if non-nil, zero value otherwise.

### GetLimitOk

`func (o *TaskList) GetLimitOk() (*int32, bool)`

GetLimitOk returns a tuple with the Limit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLimit

`func (o *TaskList) SetLimit(v int32)`

SetLimit sets Limit field to given value.

### HasLimit

`func (o *TaskList) HasLimit() bool`

HasLimit returns a boolean if a field has been set.

### GetOffset

`func (o *TaskList) GetOffset() int32`

GetOffset returns the Offset field if non-nil, zero value otherwise.

### GetOffsetOk

`func (o *TaskList) GetOffsetOk() (*int32, bool)`

GetOffsetOk returns a tuple with the Offset field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOffset

`func (o *TaskList) SetOffset(v int32)`

SetOffset sets Offset field to given value.

### HasOffset

`func (o *TaskList) HasOffset() bool`

HasOffset returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
