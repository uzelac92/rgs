package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"rgs/middleware"
	"rgs/observability"
	"rgs/services"
	"time"

	"go.uber.org/zap"
)

type SSEHandler struct {
	bus *services.EventBus
}

func NewSSEHandler(bus *services.EventBus) *SSEHandler {
	return &SSEHandler{bus: bus}
}

func (h *SSEHandler) Stream(w http.ResponseWriter, r *http.Request) {
	operator, ok := middleware.OperatorFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	lastID := r.URL.Query().Get("last_event_id")
	if lastID == "" {
		lastID = r.Header.Get("Last-Event-ID")
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	if lastID != "" {
		events := h.bus.GetBufferedEvents(operator.ID, lastID)
		for _, evt := range events {
			h.writeEvent(w, evt)
		}
		flusher.Flush()
	}

	sub := h.bus.Subscribe(operator.ID)
	defer h.bus.Unsubscribe(operator.ID, sub)

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return

		case evt := <-sub:
			h.writeEvent(w, evt)
			flusher.Flush()

		case <-ticker.C:
			_, err := fmt.Fprintf(w, ": ping\n\n")
			if err != nil {
				observability.Logger.Error("failed to write event ping", zap.Error(err))
				return
			}
			flusher.Flush()
		}
	}
}

func (h *SSEHandler) writeEvent(w http.ResponseWriter, evt services.SSEEvent) {
	data, _ := json.Marshal(evt.Data)

	_, err := fmt.Fprintf(w, "id: %s\n", evt.ID)
	if err != nil {
		observability.Logger.Error("error writing event ID", zap.Error(err))
		return
	}
	_, err = fmt.Fprintf(w, "event: %s\n", evt.EventType)
	if err != nil {
		observability.Logger.Error("error writing event type", zap.Error(err))
		return
	}
	_, err = fmt.Fprintf(w, "data: %s\n\n", data)
	if err != nil {
		observability.Logger.Error("error writing event data", zap.Error(err))
		return
	}
}
