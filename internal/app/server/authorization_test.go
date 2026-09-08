package server

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/labstack/echo/v5"

	appupdate "github.com/damonto/sigmo/internal/app/update"
)

type fakeAuthorization struct {
	start      func(context.Context) error
	authorized bool
}

func (a *fakeAuthorization) Start(ctx context.Context) error      { return a.start(ctx) }
func (a *fakeAuthorization) Authorized() bool                     { return a.authorized }
func (a *fakeAuthorization) Licensee() *appupdate.Licensee        { return nil }
func (a *fakeAuthorization) RegisterActivationRoutes(*echo.Group) {}
func (a *fakeAuthorization) RegisterStatusRoute(*echo.Group)      {}
func (a *fakeAuthorization) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func TestStartAuthorization(t *testing.T) {
	startupErr := errors.New("authorization configuration is invalid")
	tests := []struct {
		name       string
		noProvider bool
		authorized bool
		startErr   error
		cancel     bool
		want       bool
		wantErr    error
	}{
		{name: "public edition", noProvider: true, want: true},
		{name: "authorized", authorized: true, want: true},
		{name: "activation required"},
		{name: "permanent startup failure", startErr: startupErr, wantErr: startupErr},
		{name: "failure with stale authorization", authorized: true, startErr: startupErr, wantErr: startupErr},
		{name: "stopped during validation", cancel: true, startErr: startupErr, wantErr: context.Canceled},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			var authorization Authorization
			if !tt.noProvider {
				authorization = &fakeAuthorization{
					authorized: tt.authorized,
					start: func(context.Context) error {
						if tt.cancel {
							cancel()
						}
						return tt.startErr
					},
				}
			}
			got, err := startAuthorization(ctx, authorization)
			if got != tt.want || !errors.Is(err, tt.wantErr) {
				t.Fatalf("startAuthorization() = %v, %v; want %v, %v", got, err, tt.want, tt.wantErr)
			}
		})
	}
}

func TestAuthorizationWaitKeepsNetworkRecoveryRunning(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		var wg sync.WaitGroup
		defer func() {
			cancel()
			wg.Wait()
		}()
		networkReady := make(chan struct{})
		networkStopped := make(chan struct{})
		wg.Go(func() {
			defer close(networkStopped)
			time.Sleep(5 * time.Second)
			close(networkReady)
			<-ctx.Done()
		})
		authorization := &fakeAuthorization{
			authorized: true,
			start: func(ctx context.Context) error {
				select {
				case <-networkReady:
					return nil
				case <-ctx.Done():
					return ctx.Err()
				}
			},
		}
		if ok, err := startAuthorization(ctx, authorization); err != nil || !ok {
			t.Fatalf("startAuthorization() = %v, %v", ok, err)
		}
		select {
		case <-networkStopped:
			t.Fatal("successful authorization stopped network recovery")
		default:
		}
		cancel()
		wg.Wait()
	})
}

func TestConfirmUpdateHealthy(t *testing.T) {
	tests := []struct {
		name   string
		cancel bool
		want   int32
	}{
		{name: "stable process", want: 1},
		{name: "stopped process", cancel: true, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			if tt.cancel {
				cancel()
			} else {
				defer cancel()
			}
			var calls atomic.Int32
			confirmUpdateHealthyAfter(ctx, time.Millisecond, func() error {
				calls.Add(1)
				return nil
			})
			if got := calls.Load(); got != tt.want {
				t.Fatalf("mark healthy calls = %d, want %d", got, tt.want)
			}
		})
	}
}
