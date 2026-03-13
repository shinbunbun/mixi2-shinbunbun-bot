package handler

import (
	"context"
	"log/slog"

	constv1 "github.com/mixigroup/mixi2-application-sdk-go/gen/go/social/mixi/application/const/v1"
	modelv1 "github.com/mixigroup/mixi2-application-sdk-go/gen/go/social/mixi/application/model/v1"
)

type WebhookHandler struct {
	logger *slog.Logger
}

func NewWebhookHandler() *WebhookHandler {
	return &WebhookHandler{
		logger: slog.Default(),
	}
}

func (h *WebhookHandler) Handle(ctx context.Context, ev *modelv1.Event) error {
	switch ev.EventType {
	case constv1.EventType_EVENT_TYPE_POST_CREATED:
		h.logger.Info("received POST_CREATED event",
			slog.String("event_id", ev.EventId),
		)
	case constv1.EventType_EVENT_TYPE_CHAT_MESSAGE_RECEIVED:
		h.logger.Info("received CHAT_MESSAGE_RECEIVED event",
			slog.String("event_id", ev.EventId),
		)
	default:
		h.logger.Info("received event",
			slog.String("event_id", ev.EventId),
			slog.Int("event_type", int(ev.EventType)),
		)
	}
	return nil
}
