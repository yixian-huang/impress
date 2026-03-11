package provider

import "context"

type NotifyEvent struct {
	Type    string
	Subject string
	Body    string
	Meta    map[string]string
}

type NotifierProvider interface {
	Notify(ctx context.Context, event NotifyEvent) error
}
