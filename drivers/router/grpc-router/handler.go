package grpc_router

import (
	"github.com/eolinker/apinto/drivers/router/grpc-router/manager"
	"github.com/eolinker/apinto/service"
	grpc_context "github.com/eolinker/eosc/eocontext/grpc-context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/eolinker/eosc/eocontext"
)

var completeCaller = manager.NewCompleteCaller()

type grpcRouter struct {
	completeHandler eocontext.CompleteHandler

	routerName  string
	routerId    string
	serviceName string

	finisher eocontext.FinishHandler
	service  service.IService
	filters  eocontext.IChainPro
	disable  bool
}

func (h *grpcRouter) ServeHTTP(ctx eocontext.EoContext) {
	grpcContext, err := grpc_context.Assert(ctx)
	if err != nil {
		return
	}
	if h.disable {
		grpcContext.SetFinish(manager.NewErrHandler(status.Error(codes.Unavailable, "router is disable")))
		grpcContext.FastFinish()
		return
	}

	//Set Label
	ctx.SetLabel("api", h.routerName)
	ctx.SetLabel("api_id", h.routerId)
	ctx.SetLabel("service", h.serviceName)
	ctx.SetLabel("service_id", h.service.Id())
	ctx.SetLabel("ip", grpcContext.Request().RealIP())

	ctx.SetCompleteHandler(h.completeHandler)
	ctx.SetApp(h.service)
	ctx.SetBalance(h.service)
	ctx.SetUpstreamHostHandler(h.service)
	ctx.SetFinish(h.finisher)
	h.filters.Chain(ctx, completeCaller)
}
