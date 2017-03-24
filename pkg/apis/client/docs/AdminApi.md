# \AdminApi

All URIs are relative to *https://virtserver.swaggerhub.com/nebtex/Menshend/1.0.0*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AdminDeleteService**](AdminApi.md#AdminDeleteService) | **Delete** /v1/adminServices/{id} | 
[**AdminGetService**](AdminApi.md#AdminGetService) | **Get** /v1/adminServices/{id} | 
[**AdminListService**](AdminApi.md#AdminListService) | **Get** /v1/adminServices | 
[**AdminSaveService**](AdminApi.md#AdminSaveService) | **Put** /v1/adminServices/{id} | 


# **AdminDeleteService**
> interface{} AdminDeleteService($id)



delete service


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **string**| service id | 

### Return type

[**interface{}**](interface{}.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **AdminGetService**
> AdminService AdminGetService($id)



returns all the available information of the service


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **string**| service id | 

### Return type

[**AdminService**](AdminService.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **AdminListService**
> []AdminService AdminListService($subdomain)



returns all the available information of the service over all the roles


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **subdomain** | **string**| subdomain | 

### Return type

[**[]AdminService**](AdminService.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **AdminSaveService**
> AdminService AdminSaveService($id, $body)



create a new service


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **string**| service id | 
 **body** | [**AdminService**](AdminService.md)|  | 

### Return type

[**AdminService**](AdminService.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

