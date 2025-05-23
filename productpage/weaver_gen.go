// Code generated by "weaver generate". DO NOT EDIT.
//go:build !ignoreWeaverGen

package productpage

import (
	"context"
	"github.com/ServiceWeaver/weaver"
	"github.com/ServiceWeaver/weaver/runtime/codegen"
	"go.opentelemetry.io/otel/trace"
	"reflect"
)

func init() {
	codegen.Register(codegen.Registration{
		Name:      "github.com/ServiceWeaver/weaver/Main",
		Iface:     reflect.TypeOf((*weaver.Main)(nil)).Elem(),
		Impl:      reflect.TypeOf(Server{}),
		Listeners: []string{"productpage"},
		LocalStubFn: func(impl any, caller string, tracer trace.Tracer) any {
			return main_local_stub{impl: impl.(weaver.Main), tracer: tracer}
		},
		ClientStubFn: func(stub codegen.Stub, caller string) any { return main_client_stub{stub: stub} },
		ServerStubFn: func(impl any, addLoad func(uint64, float64)) codegen.Server {
			return main_server_stub{impl: impl.(weaver.Main), addLoad: addLoad}
		},
		ReflectStubFn: func(caller func(string, context.Context, []any, []any) error) any {
			return main_reflect_stub{caller: caller}
		},
		RefData: "⟦d8b12067:wEaVeReDgE:github.com/ServiceWeaver/weaver/Main→github.com/camilamedeir0s/bookinfo-serviceweaver/details/Details⟧\n⟦8404e27b:wEaVeReDgE:github.com/ServiceWeaver/weaver/Main→github.com/camilamedeir0s/bookinfo-serviceweaver/reviews/Reviews⟧\n⟦685fa2bc:wEaVeReDgE:github.com/ServiceWeaver/weaver/Main→github.com/camilamedeir0s/bookinfo-serviceweaver/ratings/Ratings⟧\n⟦3f83a94b:wEaVeRlIsTeNeRs:github.com/ServiceWeaver/weaver/Main→productpage⟧\n",
	})
}

// weaver.InstanceOf checks.
var _ weaver.InstanceOf[weaver.Main] = (*Server)(nil)

// weaver.Router checks.
var _ weaver.Unrouted = (*Server)(nil)

// Local stub implementations.

type main_local_stub struct {
	impl   weaver.Main
	tracer trace.Tracer
}

// Check that main_local_stub implements the weaver.Main interface.
var _ weaver.Main = (*main_local_stub)(nil)

// Client stub implementations.

type main_client_stub struct {
	stub codegen.Stub
}

// Check that main_client_stub implements the weaver.Main interface.
var _ weaver.Main = (*main_client_stub)(nil)

// Note that "weaver generate" will always generate the error message below.
// Everything is okay. The error message is only relevant if you see it when
// you run "go build" or "go run".
var _ codegen.LatestVersion = codegen.Version[[0][24]struct{}](`

ERROR: You generated this file with 'weaver generate' v0.24.3 (codegen
version v0.24.0). The generated code is incompatible with the version of the
github.com/ServiceWeaver/weaver module that you're using. The weaver module
version can be found in your go.mod file or by running the following command.

    go list -m github.com/ServiceWeaver/weaver

We recommend updating the weaver module and the 'weaver generate' command by
running the following.

    go get github.com/ServiceWeaver/weaver@latest
    go install github.com/ServiceWeaver/weaver/cmd/weaver@latest

Then, re-run 'weaver generate' and re-build your code. If the problem persists,
please file an issue at https://github.com/ServiceWeaver/weaver/issues.

`)

// Server stub implementations.

type main_server_stub struct {
	impl    weaver.Main
	addLoad func(key uint64, load float64)
}

// Check that main_server_stub implements the codegen.Server interface.
var _ codegen.Server = (*main_server_stub)(nil)

// GetStubFn implements the codegen.Server interface.
func (s main_server_stub) GetStubFn(method string) func(ctx context.Context, args []byte) ([]byte, error) {
	switch method {
	default:
		return nil
	}
}

// Reflect stub implementations.

type main_reflect_stub struct {
	caller func(string, context.Context, []any, []any) error
}

// Check that main_reflect_stub implements the weaver.Main interface.
var _ weaver.Main = (*main_reflect_stub)(nil)

