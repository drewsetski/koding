package kloud

import (
	"errors"
	"fmt"
	"strings"
	"time"

	cfg "koding/kites/config"
	"koding/kites/config/configstore"
	"koding/kites/kloud/stack"
	"koding/klientctl/config"
	"koding/klientctl/ctlcli"

	"github.com/koding/kite"
	kitecfg "github.com/koding/kite/config"
	"github.com/koding/logging"
)

// TODO(rjeczalik): rename to kite

// Transport is an interface that abstracts underlying
// RPC round trip.
//
// Default implementation used in this package is
// a kiteTransport, but plain net/rpc can also be
// used.
type Transport interface {
	Connect(url string) (Transport, error)
	Call(method string, arg, reply interface{}) error
}

// DefaultLog is a logger used by Client with nil Log.
var DefaultLog logging.Logger = logging.NewCustom("endpoint-kloud", false)

// DefaultClient is a default client used by Cache, Kite,
// KiteConfig and Kloud functions.
var DefaultClient = &Client{
	Transport: &KiteTransport{},
}

// Client is responsible for communication with Kloud kite.
type Client struct {
	// Transport is used for RPC communication.
	//
	// Required.
	Transport Transport

	// WaitInterval is used on polling for events.
	//
	// If zero, 10s is used by default.
	WaitInterval time.Duration

	cache *cfg.Cache
}

func (c *Client) Cache() *cfg.Cache {
	if c.cache != nil {
		return c.cache
	}

	c.cache = cfg.NewCache(configstore.CacheOptions("kd"))
	ctlcli.CloseOnExit(c.cache)

	return c.cache
}

func (c *Client) Username() string {
	if kt, ok := c.Transport.(*KiteTransport); ok {
		return kt.kiteConfig().Username
	}
	return cfg.CurrentUser.Username
}

func (c *Client) Call(method string, arg, reply interface{}) error {
	return c.Transport.Call(method, arg, reply)
}

func (c *Client) Wait(event string) <-chan *stack.EventResponse {
	ch := make(chan *stack.EventResponse, 1)

	var arg stack.EventArg

	if i := strings.IndexRune(event, '-'); i != -1 {
		arg.Type = event[:i]
		arg.EventId = event[i+1:]
	}

	if arg.Type == "" || arg.EventId == "" {
		ch <- &stack.EventResponse{
			EventId: arg.EventId,
			Error:   newErr(errors.New("malformed event string")),
		}
		close(ch)

		return ch
	}

	go func() {
		last := -1
		defer close(ch)

		id := stack.EventArgs{arg}

		for {
			var events []stack.EventResponse

			if err := c.Call("event", id, &events); err != nil {
				ch <- &stack.EventResponse{
					EventId: arg.EventId,
					Error:   newErr(err),
				}
				return
			}

			if len(events) == 0 {
				ch <- &stack.EventResponse{
					EventId: arg.EventId,
					Error:   newErr(fmt.Errorf("%s is no longer in progress", arg.Type)),
				}
				return
			}

			var event *stack.EventResponse

			for _, e := range events {
				if e.Event == nil {
					continue
				}

				if e.Event.Percentage > last {
					last = e.Event.Percentage
					event = &e
					break
				}
			}

			if event != nil {
				if event.Event.Error != "" {
					event.Error = newErr(errors.New(event.Event.Error))
				}

				ch <- event

				if event.Error != nil || event.Event.Percentage >= 100 {
					return
				}
			}

			time.Sleep(c.waitInterval())
		}
	}()

	return ch
}

func (c *Client) waitInterval() time.Duration {
	if c.WaitInterval != 0 {
		return c.WaitInterval
	}
	return 10 * time.Second
}

func newErr(err error) *kite.Error {
	if e, ok := err.(*kite.Error); ok {
		return e
	}
	return &kite.Error{
		Type:    "endpoint/kloud",
		Message: err.Error(),
	}
}

// KiteTransport is a default transport that uses github.com/koding/kite
// for underlying communication.
//
// Zero value of KiteTransport tries to connect to Kloud and Kontrol
// endpoints defined in config.Konfig (read from konfig.bolt).
type KiteTransport struct {
	// Konfig is a Koding configuration to use when calling endpoints.
	//
	// If nil, global config.Konfig is going to be used instead.
	Konfig *cfg.Konfig

	// ClientURL is an remote kite endpoint to connect to.
	//
	// If empty, kloud's public endpoint is going to be used instead.
	ClientURL string

	// DialTimeout is a maximum time external kite is
	// going to be dialed for.
	//
	// If zero, 30s is going to be used instead.
	DialTimeout time.Duration

	// TellTimeout is a maximum time of kite's
	// request/response roundtrip.
	//
	// If zero, 60s is going to be used instead.
	TellTimeout time.Duration

	// Log is used for logging.
	//
	// If nil, DefaultLog is going to be used instead.
	Log logging.Logger

	k       *kite.Kite
	kCfg    *kitecfg.Config
	kClient *kite.Client
}

var (
	_ Transport       = (*KiteTransport)(nil)
	_ stack.Validator = (*KiteTransport)(nil)
)

func (kt *KiteTransport) Call(method string, arg, reply interface{}) error {
	k, err := kt.client()
	if err != nil {
		return err
	}

	r, err := k.TellWithTimeout(method, kt.tellTimeout(), arg)
	if err != nil {
		return err
	}

	if reply != nil {
		return r.Unmarshal(reply)
	}

	return nil
}

func (kt *KiteTransport) Connect(url string) (Transport, error) {
	k, err := kt.newClient(url)
	if err != nil {
		return nil, err
	}

	ktCopy := *kt
	ktCopy.kClient = k

	return &ktCopy, nil
}

func (kt *KiteTransport) SetKiteKey(kiteKey string) {
	if kt.kClient != nil {
		kt.kClient.Auth = &kite.Auth{
			Type: "kiteKey",
			Key:  kiteKey,
		}
	}
}

func (kt *KiteTransport) kite() *kite.Kite {
	if kt.k != nil {
		return kt.k
	}

	kt.k = kite.NewWithConfig(config.Name, config.KiteVersion, kt.kiteConfig())
	kt.k.Log = kt.log()

	return kt.k
}

func (kt *KiteTransport) kiteConfig() *kitecfg.Config {
	if kt.kCfg != nil {
		return kt.kCfg
	}

	kt.kCfg = kt.konfig().KiteConfig()
	kt.kCfg.KontrolURL = kt.konfig().Endpoints.Kontrol().Public.String()
	kt.kCfg.Environment = config.Environment

	return kt.kCfg
}

func (kt *KiteTransport) client() (*kite.Client, error) {
	if kt.kClient != nil {
		return kt.kClient, nil
	}

	c, err := kt.newClient(kt.clientURL())
	if err != nil {
		return nil, err
	}

	kt.kClient = c

	return kt.kClient, nil
}

func (kt *KiteTransport) newClient(url string) (*kite.Client, error) {
	k := kt.kite().NewClient(url)
	k.Reconnect = true

	if err := k.DialTimeout(kt.dialTimeout()); err != nil {
		return nil, err
	}

	if kitekey := kt.kiteConfig().KiteKey; kitekey != "" {
		k.Auth = &kite.Auth{
			Type: "kiteKey",
			Key:  kitekey,
		}
	}

	return k, nil
}

func (kt *KiteTransport) dialTimeout() time.Duration {
	if kt.DialTimeout != 0 {
		return kt.DialTimeout
	}
	return 30 * time.Second
}

func (kt *KiteTransport) tellTimeout() time.Duration {
	if kt.TellTimeout != 0 {
		return kt.TellTimeout
	}
	return 60 * time.Second
}

func (kt *KiteTransport) log() logging.Logger {
	if kt.Log != nil {
		return kt.Log
	}
	return DefaultLog
}

func (kt *KiteTransport) konfig() *cfg.Konfig {
	if kt.Konfig != nil {
		return kt.Konfig
	}
	return config.Konfig
}

func (kt *KiteTransport) clientURL() string {
	if kt.ClientURL != "" {
		return kt.ClientURL
	}

	return kt.konfig().Endpoints.Kloud().Public.String()
}

func (kt *KiteTransport) Valid() error {
	// In order to test whether we're able to authenticate with kloud
	// we need to call some kite method. For that purpose we
	// use builtin "kite.print" method with empty string, since
	// this is the only nop method available.
	return kt.Call("kite.print", "", nil)
}

func Cache() *cfg.Cache { return DefaultClient.Cache() }
func Username() string  { return DefaultClient.Username() }
func Call(method string, arg, reply interface{}) error {
	return DefaultClient.Call(method, arg, reply)
}
func Wait(event string) <-chan *stack.EventResponse {
	return DefaultClient.Wait(event)
}
