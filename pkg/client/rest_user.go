package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/conductorone/baton-sdk/pkg/uhttp"
)

// Retool REST user-management endpoints (base lives in restClient.baseURL).
// Paths are addressed by the user `sid` (user_<uuid>), surfaced over REST as `id`.
const (
	usersEndpoint    = "/api/v2/users"    // POST create, GET ?email= lookup
	userByIDEndpoint = "/api/v2/users/%s" // PATCH (disable/enable), DELETE — %s = sid
)

// Sentinels for the idempotent/conflict states the lifecycle handlers tolerate.
var (
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrUserAlreadyDisabled = errors.New("user already disabled")
)

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
// body. It returns the raw *http.Response so callers can branch on status (e.g. delete
// idempotency). Per repo convention it returns raw errors — the connector layer wraps.
func (r *restClient) doRequest(ctx context.Context, method, path string, query url.Values, body, out interface{}) (*http.Response, error) {
	u := url.URL{
		Scheme:   r.baseURL.Scheme,
		Host:     r.baseURL.Host,
		Path:     path,
		RawQuery: query.Encode(),
	}

	reqOpts := []uhttp.RequestOption{
		uhttp.WithBearerToken(r.token),
		uhttp.WithAcceptJSONHeader(),
	}
	if body != nil {
		reqOpts = append(reqOpts, uhttp.WithJSONBody(body))
	}

	req, err := r.httpClient.NewRequest(ctx, method, &u, reqOpts...)
	if err != nil {
		return nil, err
	}

	var errResp RetoolErrorResponse
	doOpts := []uhttp.DoOption{uhttp.WithErrorResponse(&errResp)}
	if out != nil {
		doOpts = append(doOpts, uhttp.WithJSONResponse(out))
	}

	return r.httpClient.Do(req, doOpts...)
}

// ValidateREST cheaply confirms the REST token + base URL work (used by Validate()).
func (c *Client) ValidateREST(ctx context.Context) error {
	if c.rest == nil {
		return nil
	}
	q := url.Values{}
	q.Set("limit", "1")
	_, err := c.rest.doRequest(ctx, http.MethodGet, usersEndpoint, q, nil, nil)
	return err
}

// CreateUser provisions a new Retool user. Returns ErrUserAlreadyExists on a duplicate
// email (HTTP 409) so the caller can resolve and return the existing account.
func (c *Client) CreateUser(ctx context.Context, params CreateUserParams) (*RESTUser, error) {
	var out userResponse
	resp, err := c.rest.doRequest(ctx, http.MethodPost, usersEndpoint, nil, createUserRequest{
		Email:     params.Email,
		FirstName: params.FirstName,
		LastName:  params.LastName,
		UserType:  params.UserType,
	}, &out)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusConflict {
			return nil, ErrUserAlreadyExists
		}
		return nil, err
	}
	if out.Data == nil {
		return nil, fmt.Errorf("create user: empty response body")
	}
	return out.Data, nil
}

// GetUserByEmail resolves a user by email via the server-side filter. Returns
// ErrUserNotFound when no user matches and an error if the match is ambiguous.
func (c *Client) GetUserByEmail(ctx context.Context, email string) (*RESTUser, error) {
	q := url.Values{}
	q.Set("email", email)

	var out userListResponse
	_, err := c.rest.doRequest(ctx, http.MethodGet, usersEndpoint, q, nil, &out)
	if err != nil {
		return nil, err
	}

	switch len(out.Data) {
	case 0:
		return nil, ErrUserNotFound
	case 1:
		return out.Data[0], nil
	default:
		return nil, fmt.Errorf("ambiguous email lookup for %q: %d users matched", email, len(out.Data))
	}
}

// DeleteUser deprovisions a user by sid. Retool's DELETE is a soft delete (it deactivates
// the user). Idempotent: a missing user (404) or an already-deactivated one (422) both
// return their sentinels so the caller can treat them as success.
func (c *Client) DeleteUser(ctx context.Context, sid string) error {
	resp, err := c.rest.doRequest(ctx, http.MethodDelete, fmt.Sprintf(userByIDEndpoint, sid), nil, nil, nil)
	if err != nil {
		if resp != nil {
			switch resp.StatusCode {
			case http.StatusNotFound:
				return ErrUserNotFound
			case http.StatusUnprocessableEntity:
				return ErrUserAlreadyDisabled
			}
		}
		return err
	}
	return nil
}
