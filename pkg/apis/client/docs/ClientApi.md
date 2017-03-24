# \ClientApi

All URIs are relative to *https://virtserver.swaggerhub.com/nebtex/Menshend/1.0.0*

Method | HTTP request | Description
------------- | ------------- | -------------
[**ListAvailableServices**](ClientApi.md#ListAvailableServices) | **Get** /v1/clientServices | 


# **ListAvailableServices**
> []ClientService ListAvailableServices($subdomain, $role)



list or search availables services


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **subdomain** | **string**| filter by subdomain | [optional] 
 **role** | **string**| filter by role | [optional] 

### Return type

[**[]ClientService**](ClientService.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

