package zInterface

type IRouter interface {
	PreHandle(request IRequest)
	Handle(request IRequest)
	PostHandle(requset IRequest)
}
