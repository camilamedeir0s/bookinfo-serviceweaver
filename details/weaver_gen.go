// Code generated by "weaver generate". DO NOT EDIT.
//go:build !ignoreWeaverGen

package details

import (
	"context"
	"errors"
	"fmt"
	"github.com/ServiceWeaver/weaver"
	"github.com/ServiceWeaver/weaver/runtime/codegen"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"reflect"
)

func init() {
	codegen.Register(codegen.Registration{
		Name:  "github.com/camilamedeir0s/bookinfo-serviceweaver/details/Details",
		Iface: reflect.TypeOf((*Details)(nil)).Elem(),
		Impl:  reflect.TypeOf(details{}),
		LocalStubFn: func(impl any, caller string, tracer trace.Tracer) any {
			return details_local_stub{impl: impl.(Details), tracer: tracer, getBookDetailsMetrics: codegen.MethodMetricsFor(codegen.MethodLabels{Caller: caller, Component: "github.com/camilamedeir0s/bookinfo-serviceweaver/details/Details", Method: "GetBookDetails", Remote: false, Generated: true})}
		},
		ClientStubFn: func(stub codegen.Stub, caller string) any {
			return details_client_stub{stub: stub, getBookDetailsMetrics: codegen.MethodMetricsFor(codegen.MethodLabels{Caller: caller, Component: "github.com/camilamedeir0s/bookinfo-serviceweaver/details/Details", Method: "GetBookDetails", Remote: true, Generated: true})}
		},
		ServerStubFn: func(impl any, addLoad func(uint64, float64)) codegen.Server {
			return details_server_stub{impl: impl.(Details), addLoad: addLoad}
		},
		ReflectStubFn: func(caller func(string, context.Context, []any, []any) error) any {
			return details_reflect_stub{caller: caller}
		},
		RefData: "",
	})
}

// weaver.InstanceOf checks.
var _ weaver.InstanceOf[Details] = (*details)(nil)

// weaver.Router checks.
var _ weaver.Unrouted = (*details)(nil)

// Local stub implementations.

type details_local_stub struct {
	impl                  Details
	tracer                trace.Tracer
	getBookDetailsMetrics *codegen.MethodMetrics
}

// Check that details_local_stub implements the Details interface.
var _ Details = (*details_local_stub)(nil)

func (s details_local_stub) GetBookDetails(ctx context.Context, a0 int, a1 map[string]string) (r0 BookDetails, err error) {
	// Update metrics.
	begin := s.getBookDetailsMetrics.Begin()
	defer func() { s.getBookDetailsMetrics.End(begin, err != nil, 0, 0) }()
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		// Create a child span for this method.
		ctx, span = s.tracer.Start(ctx, "details.Details.GetBookDetails", trace.WithSpanKind(trace.SpanKindInternal))
		defer func() {
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}
			span.End()
		}()
	}

	return s.impl.GetBookDetails(ctx, a0, a1)
}

// Client stub implementations.

type details_client_stub struct {
	stub                  codegen.Stub
	getBookDetailsMetrics *codegen.MethodMetrics
}

// Check that details_client_stub implements the Details interface.
var _ Details = (*details_client_stub)(nil)

func (s details_client_stub) GetBookDetails(ctx context.Context, a0 int, a1 map[string]string) (r0 BookDetails, err error) {
	// Update metrics.
	var requestBytes, replyBytes int
	begin := s.getBookDetailsMetrics.Begin()
	defer func() { s.getBookDetailsMetrics.End(begin, err != nil, requestBytes, replyBytes) }()

	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		// Create a child span for this method.
		ctx, span = s.stub.Tracer().Start(ctx, "details.Details.GetBookDetails", trace.WithSpanKind(trace.SpanKindClient))
	}

	defer func() {
		// Catch and return any panics detected during encoding/decoding/rpc.
		if err == nil {
			err = codegen.CatchPanics(recover())
			if err != nil {
				err = errors.Join(weaver.RemoteCallError, err)
			}
		}

		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()

	}()

	// Encode arguments.
	enc := codegen.NewEncoder()
	enc.Int(a0)
	serviceweaver_enc_map_string_string_219dd46d(enc, a1)
	var shardKey uint64

	// Call the remote method.
	requestBytes = len(enc.Data())
	var results []byte
	results, err = s.stub.Run(ctx, 0, enc.Data(), shardKey)
	replyBytes = len(results)
	if err != nil {
		err = errors.Join(weaver.RemoteCallError, err)
		return
	}

	// Decode the results.
	dec := codegen.NewDecoder(results)
	(&r0).WeaverUnmarshal(dec)
	err = dec.Error()
	return
}

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

type details_server_stub struct {
	impl    Details
	addLoad func(key uint64, load float64)
}

// Check that details_server_stub implements the codegen.Server interface.
var _ codegen.Server = (*details_server_stub)(nil)

// GetStubFn implements the codegen.Server interface.
func (s details_server_stub) GetStubFn(method string) func(ctx context.Context, args []byte) ([]byte, error) {
	switch method {
	case "GetBookDetails":
		return s.getBookDetails
	default:
		return nil
	}
}

func (s details_server_stub) getBookDetails(ctx context.Context, args []byte) (res []byte, err error) {
	// Catch and return any panics detected during encoding/decoding/rpc.
	defer func() {
		if err == nil {
			err = codegen.CatchPanics(recover())
		}
	}()

	// Decode arguments.
	dec := codegen.NewDecoder(args)
	var a0 int
	a0 = dec.Int()
	var a1 map[string]string
	a1 = serviceweaver_dec_map_string_string_219dd46d(dec)

	// TODO(rgrandl): The deferred function above will recover from panics in the
	// user code: fix this.
	// Call the local method.
	r0, appErr := s.impl.GetBookDetails(ctx, a0, a1)

	// Encode the results.
	enc := codegen.NewEncoder()
	(r0).WeaverMarshal(enc)
	enc.Error(appErr)
	return enc.Data(), nil
}

// Reflect stub implementations.

type details_reflect_stub struct {
	caller func(string, context.Context, []any, []any) error
}

// Check that details_reflect_stub implements the Details interface.
var _ Details = (*details_reflect_stub)(nil)

func (s details_reflect_stub) GetBookDetails(ctx context.Context, a0 int, a1 map[string]string) (r0 BookDetails, err error) {
	err = s.caller("GetBookDetails", ctx, []any{a0, a1}, []any{&r0})
	return
}

// AutoMarshal implementations.

var _ codegen.AutoMarshal = (*BookDetails)(nil)

type __is_BookDetails[T ~struct {
	weaver.AutoMarshal
	ID        int    "json:\"id\""
	Author    string "json:\"author\""
	Year      int    "json:\"year\""
	Type      string "json:\"type\""
	Pages     int    "json:\"pages\""
	Publisher string "json:\"publisher\""
	Language  string "json:\"language\""
	ISBN10    string "json:\"ISBN-10\""
	ISBN13    string "json:\"ISBN-13\""
}] struct{}

var _ __is_BookDetails[BookDetails]

func (x *BookDetails) WeaverMarshal(enc *codegen.Encoder) {
	if x == nil {
		panic(fmt.Errorf("BookDetails.WeaverMarshal: nil receiver"))
	}
	enc.Int(x.ID)
	enc.String(x.Author)
	enc.Int(x.Year)
	enc.String(x.Type)
	enc.Int(x.Pages)
	enc.String(x.Publisher)
	enc.String(x.Language)
	enc.String(x.ISBN10)
	enc.String(x.ISBN13)
}

func (x *BookDetails) WeaverUnmarshal(dec *codegen.Decoder) {
	if x == nil {
		panic(fmt.Errorf("BookDetails.WeaverUnmarshal: nil receiver"))
	}
	x.ID = dec.Int()
	x.Author = dec.String()
	x.Year = dec.Int()
	x.Type = dec.String()
	x.Pages = dec.Int()
	x.Publisher = dec.String()
	x.Language = dec.String()
	x.ISBN10 = dec.String()
	x.ISBN13 = dec.String()
}

// Encoding/decoding implementations.

func serviceweaver_enc_map_string_string_219dd46d(enc *codegen.Encoder, arg map[string]string) {
	if arg == nil {
		enc.Len(-1)
		return
	}
	enc.Len(len(arg))
	for k, v := range arg {
		enc.String(k)
		enc.String(v)
	}
}

func serviceweaver_dec_map_string_string_219dd46d(dec *codegen.Decoder) map[string]string {
	n := dec.Len()
	if n == -1 {
		return nil
	}
	res := make(map[string]string, n)
	var k string
	var v string
	for i := 0; i < n; i++ {
		k = dec.String()
		v = dec.String()
		res[k] = v
	}
	return res
}
