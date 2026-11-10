# TaskPage

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Tasks** | [**TaskList**](TaskList.md) |  |

## Methods

### NewTaskPage

`func NewTaskPage(tasks TaskList, ) *TaskPage`

NewTaskPage instantiates a new TaskPage object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTaskPageWithDefaults

`func NewTaskPageWithDefaults() *TaskPage`

NewTaskPageWithDefaults instantiates a new TaskPage object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetTasks

`func (o *TaskPage) GetTasks() TaskList`

GetTasks returns the Tasks field if non-nil, zero value otherwise.

### GetTasksOk

`func (o *TaskPage) GetTasksOk() (*TaskList, bool)`

GetTasksOk returns a tuple with the Tasks field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTasks

`func (o *TaskPage) SetTasks(v TaskList)`

SetTasks sets Tasks field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
