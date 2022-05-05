package request

// RequestPool ...
var RequestPool *requestPool

func init() {
	RequestPool = newRequestPool()
}
