package handler

var (
	authRequestPool                *requestBodyPool[authRequest]
	authByActionsRequestPool       *requestBodyPool[authByActionsRequest]
	authByResourcesRequestPool     *requestBodyPool[authByResourcesRequest]
	queryRequestPool               *requestBodyPool[queryRequest]
	queryByActionsRequestPool      *requestBodyPool[queryByActionsRequest]
	queryByExtResourcesRequestPool *requestBodyPool[queryByExtResourcesRequest]
)

func init() {
	authRequestPool = newRequestBodyPool[authRequest]()
	authByActionsRequestPool = newRequestBodyPool[authByActionsRequest]()
	authByResourcesRequestPool = newRequestBodyPool[authByResourcesRequest]()
	queryRequestPool = newRequestBodyPool[queryRequest]()
	queryByActionsRequestPool = newRequestBodyPool[queryByActionsRequest]()
	queryByExtResourcesRequestPool = newRequestBodyPool[queryByExtResourcesRequest]()
}
