package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
)

// Retool REST user-management endpoints (base lives in restClient.baseURL).
// Paths are addressed by the user `sid` (user_<uuid>), surfaced over REST as `id`.
const (
	usersEndpoint    = "/api/v2/users"    // POST create, GET ?email= lookup
	userByIDEndpoint = "/api/v2/users/%s" // PATCH (enable/disable) — %s = sid
)

// ErrUserAlreadyExists is the idempotent-create sentinel the lifecycle handlers tolerate.
var ErrUserAlreadyExists = errors.New("user already exists")

// RESTUser is the Retool REST representation of a user. `ID` is the stable `sid`
// (user_<uuid>); `LegacyID` is the Postgres `users.id` that baton keys `user:<int64>` on.
type RESTUser struct {
	ID        string `json:"id"`
	LegacyID  int64  `json:"legacy_id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Active    bool   `json:"active"`
	UserType  string `json:"user_type"`
}

// CreateUserParams are the fields required to provision a new Retool user.
type CreateUserParams struct {
	Email     string
	FirstName string
	LastName  string
	UserType  string
}

type createUserRequest struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	UserType  string `json:"user_type,omitempty"`
}

type userResponse struct {
	Success bool      `json:"success"`
	Data    *RESTUser `json:"data"`
}

type userListResponse struct {
	Success    bool        `json:"success"`
	Data       []*RESTUser `json:"data"`
	TotalCount int         `json:"total_count"`
	HasMore    bool        `json:"has_more"`
}

// RetoolErrorResponse satisfies uhttp.ErrorResponse so error bodies surface their message.
type RetoolErrorResponse struct {
	Success bool   `json:"success"`
	Msg     string `json:"message"`
}

func (e *RetoolErrorResponse) Message() string {
	return e.Msg
}

// doRequest is the single REST chokepoint: bearer auth, JSON in/out, and a typed error
// body. It returns rate-limit annotations (so the SDK can back off) and the HTTP status
// code so callers can branch on it (e.g. conflict/not-found) without leaking the
// *http.Response. Per repo convention it returns raw errors — the connector layer wraps.
func (r *restClient) doRequest(ctx context.Context, method, endpoint string, query url.Values, body, out interface{}) (annotations.Annotations, int, error) {
	// JoinPath preserves any base-URL path prefix (e.g. a reverse proxy at /retool).
	u := r.baseURL.JoinPath(endpoint)
	u.RawQuery = query.Encode()

	reqOpts := []uhttp.RequestOption{
		uhttp.WithBearerToken(r.token),
		uhttp.WithAcceptJSONHeader(),
	}
	if body != nil {
		reqOpts = append(reqOpts, uhttp.WithJSONBody(body))
	}

	req, err := r.httpClient.NewRequest(ctx, method, u, reqOpts...)
	if err != nil {
		return nil, 0, err
	}

	var errResp RetoolErrorResponse
	var ratelimitData v2.RateLimitDescription
	doOpts := []uhttp.DoOption{
		uhttp.WithErrorResponse(&errResp),
		uhttp.WithRatelimitData(&ratelimitData),
	}
	if out != nil {
		doOpts = append(doOpts, uhttp.WithJSONResponse(out))
	}

	resp, err := r.httpClient.Do(req, doOpts...)

	// Surface rate-limit data on every path (including 429/errors) so the caller can
	// merge it into the action/provisioning annotations and the SDK can back off.
	var annos annotations.Annotations
	annos.WithRateLimiting(&ratelimitData)

	if resp != nil {
		// uhttp already drained, closed, and replaced the body with a no-op closer;
		// closing again is harmless and satisfies the bodyclose linter.
		defer resp.Body.Close()
		return annos, resp.StatusCode, err
	}
	return annos, 0, err
}

// ValidateREST cheaply confirms the REST token + base URL work (used by Validate()).
func (c *Client) ValidateREST(ctx context.Context) (annotations.Annotations, error) {
	if c.rest == nil {
		return nil, nil
	}
	q := url.Values{}
	q.Set("limit", "1")
	annos, _, err := c.rest.doRequest(ctx, http.MethodGet, usersEndpoint, q, nil, nil)
	return annos, err
}

// CreateUser provisions a new Retool user. Returns ErrUserAlreadyExists on a duplicate
// email (HTTP 409) so the caller can resolve and return the existing account.
func (c *Client) CreateUser(ctx context.Context, params CreateUserParams) (*RESTUser, annotations.Annotations, error) {
	var out userResponse
	annos, status, err := c.rest.doRequest(ctx, http.MethodPost, usersEndpoint, nil, createUserRequest(params), &out)
	if err != nil {
		if status == http.StatusConflict {
			return nil, annos, ErrUserAlreadyExists
		}
		return nil, annos, err
	}
	if out.Data == nil {
		return nil, annos, fmt.Errorf("create user: empty response body")
	}
	return out.Data, annos, nil
}

// GetUserByEmail resolves a user by email via the server-side filter. Returns
// ErrUserNotFound when no user matches and an error if the match is ambiguous.
func (c *Client) GetUserByEmail(ctx context.Context, email string) (*RESTUser, annotations.Annotations, error) {
	q := url.Values{}
	q.Set("email", email)

	var out userListResponse
	annos, _, err := c.rest.doRequest(ctx, http.MethodGet, usersEndpoint, q, nil, &out)
	if err != nil {
		return nil, annos, err
	}

	switch len(out.Data) {
	case 0:
		return nil, annos, ErrUserNotFound
	case 1:
		return out.Data[0], annos, nil
	default:
		return nil, annos, fmt.Errorf("ambiguous email lookup for %q: %d users matched", email, len(out.Data))
	}
}

// patchOperation is a JSON-Patch operation as accepted by PATCH /api/v2/users/{id}.
type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

type patchUserRequest struct {
	Operations []patchOperation `json:"operations"`
}

// SetUserActive enables (active=true) or disables (active=false) a user by sid.
// Idempotent: patching to the current state succeeds. Retool has no hard delete —
// this deactivation/reactivation pair is the whole account lifecycle after create.
func (c *Client) SetUserActive(ctx context.Context, sid string, active bool) (annotations.Annotations, error) {
	body := patchUserRequest{
		Operations: []patchOperation{{Op: "replace", Path: "/active", Value: active}},
	}
	annos, status, err := c.rest.doRequest(ctx, http.MethodPatch, fmt.Sprintf(userByIDEndpoint, sid), nil, body, nil)
	if err != nil {
		if status == http.StatusNotFound {
			return annos, ErrUserNotFound
		}
		return annos, err
	}
	return annos, nil
}
