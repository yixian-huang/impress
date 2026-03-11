package provider

import "context"

type CaptchaProvider interface {
	Verify(ctx context.Context, token string, remoteIP string) error
}

type NoopCaptchaProvider struct{}

func (p *NoopCaptchaProvider) Verify(ctx context.Context, token string, remoteIP string) error {
	return nil
}
