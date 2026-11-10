# ProposalPatch

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**InputJson** | **map[string]interface{}** |  |
**EditReason** | Pointer to **string** |  | [optional]

## Methods

### NewProposalPatch

`func NewProposalPatch(inputJson map[string]interface{}, ) *ProposalPatch`

NewProposalPatch instantiates a new ProposalPatch object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProposalPatchWithDefaults

`func NewProposalPatchWithDefaults() *ProposalPatch`

NewProposalPatchWithDefaults instantiates a new ProposalPatch object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetInputJson

`func (o *ProposalPatch) GetInputJson() map[string]interface{}`

GetInputJson returns the InputJson field if non-nil, zero value otherwise.

### GetInputJsonOk

`func (o *ProposalPatch) GetInputJsonOk() (*map[string]interface{}, bool)`

GetInputJsonOk returns a tuple with the InputJson field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInputJson

`func (o *ProposalPatch) SetInputJson(v map[string]interface{})`

SetInputJson sets InputJson field to given value.


### GetEditReason

`func (o *ProposalPatch) GetEditReason() string`

GetEditReason returns the EditReason field if non-nil, zero value otherwise.

### GetEditReasonOk

`func (o *ProposalPatch) GetEditReasonOk() (*string, bool)`

GetEditReasonOk returns a tuple with the EditReason field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEditReason

`func (o *ProposalPatch) SetEditReason(v string)`

SetEditReason sets EditReason field to given value.

### HasEditReason

`func (o *ProposalPatch) HasEditReason() bool`

HasEditReason returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
