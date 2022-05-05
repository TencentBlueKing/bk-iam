package handler

var (
	authRequestBodyPool                *requestBodyPool[authRequest]
	authByActionsRequestBodyPool       *requestBodyPool[authByActionsRequest]
	authByResourcesRequestBodyPool     *requestBodyPool[authByResourcesRequest]
	queryRequestBodyPool               *requestBodyPool[queryRequest]
	queryByActionsRequestBodyPool      *requestBodyPool[queryByActionsRequest]
	queryByExtResourcesRequestBodyPool *requestBodyPool[queryByExtResourcesRequest]
)

func init() {
	authRequestBodyPool = newRequestBodyPool[authRequest]()
	authByActionsRequestBodyPool = newRequestBodyPool[authByActionsRequest]()
	authByResourcesRequestBodyPool = newRequestBodyPool[authByResourcesRequest]()
	queryRequestBodyPool = newRequestBodyPool[queryRequest]()
	queryByActionsRequestBodyPool = newRequestBodyPool[queryByActionsRequest]()
	queryByExtResourcesRequestBodyPool = newRequestBodyPool[queryByExtResourcesRequest]()
}
