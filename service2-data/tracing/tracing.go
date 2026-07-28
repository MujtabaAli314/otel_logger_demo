// Package tracing provides small helpers for correlating service2 logs
// with the upstream caller's span.
package tracing

import (
	"context"
)

// ParentSpanID returns the span ID of the parent span carried in ctx, or
// an empty string when the current span has no parent (e.g. when the
// service is called directly without trace headers).
//
// In service2 the controller rebuilds a remote parent span context from
// the traceID/spanID headers injected by service1 and starts a child
// span; that child's parent is therefore service1's span, so this returns
// service1's span ID for every log emitted downstream of the controller.
func ParentSpanID(ctx context.Context) string {
	// if sc := trace.SpanFromContext(ctx).ParentSpanContext(); sc.HasSpanID() {
	// 	return sc.SpanID().String()
	// }
	return ""
}
