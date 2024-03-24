package business

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/dgrijalva/jwt-go"
	"github.com/quiver-london/go-revolut/business/1.0/request"
)

type OAuthService struct {
	clientId   string
	privateKey *rsa.PrivateKey
	issuer     string
	sandbox    bool
}

func NewOAuth(clientId string, privateKey *rsa.PrivateKey, issuer string, sandbox bool) *OAuthService {
	return &OAuthService{
		clientId:   clientId,
		privateKey: privateKey,
		issuer:     issuer,
		sandbox:    sandbox,
	}
}

const (
	clientAssertionType = "urn:ietf:params:oauth:client-assertion-type:jwt-bearer"
	aud                 = "https://revolut.com"

	grant_type_authorization_code = "authorization_code"
	grant_type_refresh_token      = "refresh_token"
)

type OAuthResp struct {
	// the access token
	AccessToken string `json:"access_token"`
	// "bearer" means that this token is valid to access the API
	TokenType string `json:"token_type"`
	// token expiration time in seconds
	ExpiresIn int32 `json:"expires_in"`
	// A token to be used to request a new access token
	RefreshToken string `json:"refresh_token"`
}

type AuthorizationCodeResp struct {
	// the account ID
	Id string
	// the user authorisation code (if granted)
	Code string
}

// ExchangeAuthorisationCode: This endpoint is used to exchange an authorisation code with an access token.
// doc: https://revolut-engineering.github.io/api-docs/#business-api-business-api-oauth-get-authorisation-code
func (oa *OAuthService) ExchangeAuthorisationCode(code string) (*OAuthResp, error) {
	clientAssertion, err := oa.generateClientAssertion()
	if err != nil {
		return nil, err
	}

	resp, statusCode, err := request.New(request.Config{
		Method:  http.MethodPost,
		Url:     "https://b2b.revolut.com/api/1.0/auth/token",
		Sandbox: oa.sandbox,
		Body: url.Values{
			// "authorization_code"
			"grant_type": []string{grant_type_authorization_code},
			// an authorisation code
			"code": []string{code},
			// your app ID
			"client_id": []string{oa.clientId},
			// "urn:ietf:params:oauth:client-assertion-type:jwt-bearer"
			"client_assertion_type": []string{clientAssertionType},
			// Your generated JWT token
			"client_assertion": []string{clientAssertion},
		},
		ContentType: request.ContentType_APPLICATION_FORM,
	})
	if err != nil {
		return nil, err
	}

	if statusCode != http.StatusOK {
		return nil, errors.New(string(resp))
	}

	r := &OAuthResp{}
	if err := json.Unmarshal(resp, r); err != nil {
		return nil, err
	}

	return r, nil
}

// RefreshAccessToken: This endpoint is used to request a new user access token after the expiration date.
// doc: https://revolut-engineering.github.io/api-docs/#business-api-business-api-oauth-refresh-access-token
func (oa *OAuthService) RefreshAccessToken(refreshToken string) (*OAuthResp, error) {
	clientAssertion, err := oa.generateClientAssertion()
	if err != nil {
		return nil, err
	}

	resp, statusCode, err := request.New(request.Config{
		Method:  http.MethodPost,
		Url:     "https://b2b.revolut.com/api/1.0/auth/token",
		Sandbox: oa.sandbox,
		Body: url.Values{
			"grant_type":            []string{grant_type_refresh_token},
			"refresh_token":         []string{refreshToken},
			"client_id":             []string{oa.clientId},
			"client_assertion_type": []string{clientAssertionType},
			"client_assertion":      []string{clientAssertion},
		},
		ContentType: request.ContentType_APPLICATION_FORM,
	})
	if err != nil {
		return nil, err
	}

	if statusCode != http.StatusOK {
		return nil, errors.New(string(resp))
	}

	r := &OAuthResp{}
	if err := json.Unmarshal(resp, r); err != nil {
		return nil, err
	}

	return r, nil
}

// GetAuthorisationCode: Navigate the user to this address to request an authorisation code
// doc: https://revolut-engineering.github.io/api-docs/business-api/#oauth-get-authorisation-code
func (oa *OAuthService) GetAuthorisationCode(clientId, redirectUri string) ([]*AuthorizationCodeResp, error) {

	resp, statusCode, err := request.New(request.Config{
		Method: http.MethodGet,
		Url:    fmt.Sprintf("https://business.revolut.com/app-confirm?client_id=%s&redirect_uri%s", clientId, redirectUri),
		Body:   nil,
	})
	if err != nil {
		return nil, err
	}

	if statusCode != http.StatusOK {
		return nil, errors.New(string(resp))
	}

	var r []*AuthorizationCodeResp
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return r, nil
}

func (oa *OAuthService) generateClientAssertion() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256,
		jwt.MapClaims{
			"iss": oa.issuer,
			"aud": aud,
			"sub": oa.clientId,
		})

	signedToken, err := token.SignedString(oa.privateKey)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}
