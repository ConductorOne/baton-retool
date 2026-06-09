package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Client struct {
	db *pgxpool.Pool

	// REST surface for account provisioning/deprovisioning (CXH-1585). nil when the
	// connector is configured for sync-only (no API base URL + token).
	rest *restClient
}

// restClient holds the bearer-authenticated HTTP surface for the Retool REST API.
type restClient struct {
	httpClient *uhttp.BaseHttpClient
	baseURL    *url.URL
	token      string
}

// RESTEnabled reports whether the REST surface (account lifecycle) is configured.
func (c *Client) RESTEnabled() bool {
	return c.rest != nil
}

func (c *Client) ValidateConnection(ctx context.Context) error {
	err := c.db.Ping(ctx)
	if err != nil {
		return err
	}

	return nil
}

func New(ctx context.Context, dsn string, apiBaseURL string, apiToken string) (*Client, error) {
	l := ctxzap.Extract(ctx)

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	logger := &Logger{}
	config.ConnConfig.LogLevel = logger.Zap2PgxLogLevel(l.Level())
	config.ConnConfig.Logger = logger

	if config.ConnConfig.Database == "" {
		return nil, fmt.Errorf("must specify a database to connect to")
	}

	db, err := pgxpool.ConnectConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	c := &Client{
		db: db,
	}

	// Optional REST surface. Both fields are validated as required-together at the
	// config layer, so presence of one implies the other.
	if apiBaseURL != "" && apiToken != "" {
		rest, err := newRESTClient(ctx, apiBaseURL, apiToken)
		if err != nil {
			return nil, err
		}
		c.rest = rest
	}

	return c, nil
}

func newRESTClient(ctx context.Context, apiBaseURL string, apiToken string) (*restClient, error) {
	base, err := url.Parse(strings.TrimRight(apiBaseURL, "/"))
	if err != nil {
		return nil, fmt.Errorf("invalid retool-api-base-url: %w", err)
	}
	if base.Scheme == "" || base.Host == "" {
		return nil, fmt.Errorf("invalid retool-api-base-url %q: must include scheme and host", apiBaseURL)
	}

	httpClient, err := uhttp.NewBaseHttpClientWithContext(ctx, &http.Client{Timeout: 30 * time.Second})
	if err != nil {
		return nil, err
	}

	return &restClient{
		httpClient: httpClient,
		baseURL:    base,
		token:      apiToken,
	}, nil
}
