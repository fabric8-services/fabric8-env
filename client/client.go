package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/fabric8-services/fabric8-auth-client/auth"
	goaclient "github.com/goadesign/goa/client"
	"github.com/goadesign/goa/middleware/security/jwt"
	errs "github.com/pkg/errors"
)

func NewAuthClient(hostURL string) (*AuthClient, error) {
	u, err := url.Parse(hostURL)
	if err != nil {
		return nil, err
	}

	client := http.Client{}
	c := auth.New(&doer{
		target: goaclient.HTTPClientDoer(&client),
	})
	c.Host = u.Host
	c.Scheme = u.Scheme
	return &AuthClient{c}, nil
}

type doer struct {
	target goaclient.Doer
}

func (d *doer) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	token := jwt.ContextJWT(ctx)
	if token != nil {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.Raw))
	}
	return d.target.Do(ctx, req)
}

type AuthClient struct {
	*auth.Client
}

func (c *AuthClient) CheckSpaceScope(ctx context.Context, spaceID, requiredScope string) error {
	resp, err := c.Client.ScopesResource(ctx, auth.ScopesResourcePath(spaceID))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errs.Errorf("get space's scope failed with error '%s'", resp.Status)
	}

	defer resp.Body.Close()
	scopes, err := c.Client.DecodeResourceScopesData(resp)
	for _, scope := range scopes.Data {
		if requiredScope == scope.ID {
			return nil
		}
	}
	return errs.Errorf("user doesn't have '%s' permission on '%s' space", requiredScope, spaceID)
}
