package a2a

import (
	"context"
	"fmt"

	a2alib "github.com/a2aproject/a2a-go/a2a"
	"github.com/a2aproject/a2a-go/a2asrv"
	"github.com/a2aproject/a2a-go/a2asrv/eventqueue"

	"github.com/owulveryck/ucp-merchant-test/internal/ucp"
)

// executor implements [a2asrv.AgentExecutor] by dispatching A2A messages
// to the merchant interface based on the "action" field in DataParts.
type executor struct {
	server *Server
}

// actionContext carries the parsed request context for an action handler.
type actionContext struct {
	userID    string
	country   ucp.Country
	contextID string
	data      map[string]any
	parts     a2alib.ContentParts
}

// Execute processes an incoming A2A message by extracting the action from
// DataParts and dispatching to the appropriate merchant method.
//
// The response is written to the event queue as a [*a2alib.Message], which
// the a2a-go framework returns directly as the message/send result. This
// produces a Message response (not a Task), matching the UCP A2A spec.
//
// For checkout actions, the result DataPart contains the checkout object
// wrapped in {"a2a.ucp.checkout": ...} per the UCP A2A Checkout Binding.
func (e *executor) Execute(ctx context.Context, reqCtx *a2asrv.RequestContext, queue eventqueue.Queue) error {
	action, data := extractAction(reqCtx.Message)
	if action == "" {
		msg := a2alib.NewMessageForTask(a2alib.MessageRoleAgent, reqCtx,
			a2alib.TextPart{Text: "Error: no action specified in DataPart"})
		return queue.Write(ctx, msg)
	}

	handler, ok := e.server.actionHandlers()[action]
	if !ok {
		msg := a2alib.NewMessageForTask(a2alib.MessageRoleAgent, reqCtx,
			a2alib.TextPart{Text: fmt.Sprintf("Error: unknown action: %s", action)})
		return queue.Write(ctx, msg)
	}

	ac := &actionContext{
		userID:    userIDFromContext(ctx),
		country:   userCountryFromContext(ctx),
		contextID: reqCtx.ContextID,
		data:      data,
		parts:     reqCtx.Message.Parts,
	}

	result, err := handler(ctx, ac)
	if err != nil {
		msg := a2alib.NewMessageForTask(a2alib.MessageRoleAgent, reqCtx,
			a2alib.TextPart{Text: "Error: " + err.Error()})
		return queue.Write(ctx, msg)
	}

	msg := a2alib.NewMessageForTask(a2alib.MessageRoleAgent, reqCtx,
		a2alib.DataPart{Data: result})
	return queue.Write(ctx, msg)
}

// Cancel rejects task cancellation since UCP actions are synchronous.
func (e *executor) Cancel(_ context.Context, _ *a2asrv.RequestContext, _ eventqueue.Queue) error {
	return a2alib.NewError(a2alib.ErrTaskNotCancelable, "task cancellation not supported")
}

// extractAction finds the first DataPart containing an "action" field
// and returns the action name and the full data map from that part.
func extractAction(msg *a2alib.Message) (string, map[string]any) {
	if msg == nil {
		return "", nil
	}
	for _, part := range msg.Parts {
		data, ok := asDataPart(part)
		if !ok {
			continue
		}
		if action, ok := data["action"].(string); ok {
			return action, data
		}
	}
	return "", nil
}
