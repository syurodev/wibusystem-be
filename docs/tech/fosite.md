Documentation

There is an API documentation available at godoc.org/ory/fosite.

Scopes

Fosite has three strategies for matching scopes. You can replace the default scope strategy if you need a custom one by implementing fosite.ScopeStrategy.

Using the composer, setting a strategy is easy:

import "github.com/ory/fosite"

var config = &fosite.Config{
ScopeStrategy: fosite.HierarchicScopeStrategy,
}
Note: To issue refresh tokens with any of the grants, you need to include the offline scope in the OAuth2 request. This can be modified by the RefreshTokenScopes compose configuration. When set to an empty array, all grants will issue refresh tokens.

fosite.WildcardScopeStrategy

This is the default strategy, and the safest one. It is best explained by looking at some examples:

users.* matches users.read
users.* matches users.read.foo
users.read matches users.read
users does not match users.read
users.read.* does not match users.read
users.*.* does not match users.read
users.*.* matches users.read.own
users.*.* matches users.read.own.other
users.read.* matches users.read.own
users.read.* matches users.read.own.other
users.write.* does not match users.read.own
users.*.bar matches users.baz.bar
users.*.bar does not users.baz.baz.bar
To request users.*, a client must have exactly users.* as granted scope.

fosite.ExactScopeStrategy

This strategy is searching only for exact matches. It returns true iff the scope is granted.

fosite.HierarchicScopeStrategy

This strategy is deprecated, use it with care. Again, it is best explained by looking at some examples:

users matches users
users matches users.read
users matches users.read.own
users.read matches users.read
users.read matches users.read.own
users.read does not match users.write
users.read does not match users.write.own
Globalization

Fosite does not natively carry translations for error messages and hints, but offers an interface that allows the consumer to define catalog bundles and an implementation to translate. This is available through the MessageCatalog interface. The functions defined are self-explanatory. The DefaultMessageCatalog illustrates this. Compose config has been extended to take in an instance of the MessageCatalog.

Building translated files

There are three possible "message key" types:

Value of RFC6749Error.ErrorField: This is a string like invalid_request and correlates to most errors produced by Fosite.
Hint identifier passed into RFC6749Error.WithHintIDOrDefaultf: This func is not used extensively in Fosite but, in time, most WithHint and WithHintf will be replaced with this function.
Free text string format passed into RFC6749Error.WithHint and RFC6749Error.WithHintf: This function is used in Fosite and Hydra extensively and any message catalog implementation can use the format string parameter as the message key.
An example of a message catalog can be seen in the i18n_test.go.

Generating the en messages file

This is a WIP at the moment, but effectively any scripting language can be used to generate this. It would need to traverse all files in the source code and extract the possible message identifiers based on the different message key types.

Quickstart

Instantiating fosite by hand can be painful. Therefore we created a few convenience helpers available through the compose package. It is strongly encouraged to use these well tested composers.

In this very basic example, we will instantiate fosite with all OpenID Connect and OAuth2 handlers enabled. Please refer to the example app for more details.

This little code snippet sets up a full-blown OAuth2 and OpenID Connect example.

package main

import "github.com/ory/fosite"
import "github.com/ory/fosite/compose"
import "github.com/ory/fosite/storage"

// This is the example storage that contains:
// * an OAuth2 Client with id "my-client" and secrets "foobar" and "foobaz" capable of all oauth2 and open id connect grant and response types.
// * a User for the resource owner password credentials grant type with username "peter" and password "secret".
//
// You will most likely replace this with your own logic once you set up a real world application.
var storage = storage.NewExampleStore()

// This secret is being used to sign access and refresh tokens as well as
// authorization codes. It must be exactly 32 bytes long.
var secret = []byte("my super secret signing password")

privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
if err != nil {
panic("unable to create private key")
}

// check the api docs of fosite.Config for further configuration options
var config = &fosite.Config{
	AccessTokenLifespan: time.Minute * 30,
	GlobalSecret: secret,
	// ...
}

var oauth2Provider = compose.ComposeAllEnabled(config, storage, privateKey)

// The authorize endpoint is usually at "https://mydomain.com/oauth2/auth".
func authorizeHandlerFunc(rw http.ResponseWriter, req *http.Request) {
	// This context will be passed to all methods. It doesn't fulfill a real purpose in the standard library but could be used
	// to abort database lookups or similar things.
	ctx := req.Context()

	// Let's create an AuthorizeRequest object!
	// It will analyze the request and extract important information like scopes, response type and others.
	ar, err := oauth2Provider.NewAuthorizeRequest(ctx, req)
	if err != nil {
		oauth2Provider.WriteAuthorizeError(ctx, rw, ar, err)
		return
	}

	// Normally, this would be the place where you would check if the user is logged in and gives his consent.
	// We're simplifying things and just checking if the request includes a valid username and password
	if req.Form.Get("username") != "peter" {
		rw.Header().Set("Content-Type", "text/html;charset=UTF-8")
		rw.Write([]byte(`<h1>Login page</h1>`))
		rw.Write([]byte(`
			<p>Howdy! This is the log in page. For this example, it is enough to supply the username.</p>
			<form method="post">
				<input type="text" name="username" /> <small>try peter</small><br>
				<input type="submit">
			</form>
		`))
		return
	}

	// Now that the user is authorized, we set up a session. When validating / looking up tokens, we additionally get
	// the session. You can store anything you want in it.

	// The session will be persisted by the store and made available when e.g. validating tokens or handling token endpoint requests.
	// The default OAuth2 and OpenID Connect handlers require the session to implement a few methods. Apart from that, the
	// session struct can be anything you want it to be.
	mySessionData := &fosite.DefaultSession{
		Username: req.Form.Get("username"),
	}

	// It's also wise to check the requested scopes, e.g.:
	// if authorizeRequest.GetScopes().Has("admin") {
	//     http.Error(rw, "you're not allowed to do that", http.StatusForbidden)
	//     return
	// }

	// Now we need to get a response. This is the place where the AuthorizeEndpointHandlers kick in and start processing the request.
	// NewAuthorizeResponse is capable of running multiple response type handlers which in turn enables this library
	// to support open id connect.
	response, err := oauth2Provider.NewAuthorizeResponse(ctx, ar, mySessionData)
	if err != nil {
		oauth2Provider.WriteAuthorizeError(ctx, rw, ar, err)
		return
	}

	// Awesome, now we redirect back to the client redirect uri and pass along an authorize code
	oauth2Provider.WriteAuthorizeResponse(ctx, rw, ar, response)
}

// The token endpoint is usually at "https://mydomain.com/oauth2/token"
func tokenHandlerFunc(rw http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	// Create an empty session object that will be passed to storage implementation to populate (unmarshal) the session into.
	// By passing an empty session object as a "prototype" to the store, the store can use the underlying type to unmarshal the value into it.
	// For an example of storage implementation that takes advantage of that, see SQL Store (fosite_store_sql.go) from ory/Hydra project.
	mySessionData := new(fosite.DefaultSession)

	// This will create an access request object and iterate through the registered TokenEndpointHandlers to validate the request.
	accessRequest, err := oauth2Provider.NewAccessRequest(ctx, req, mySessionData)
	if err != nil {
		oauth2Provider.WriteAccessError(ctx, rw, accessRequest, err)
		return
	}

	if mySessionData.Username == "super-admin-guy" {
		// do something...
	}

	// Next we create a response for the access request. Again, we iterate through the TokenEndpointHandlers
	// and aggregate the result in response.
	response, err := oauth2Provider.NewAccessResponse(ctx, accessRequest)
	if err != nil {
		oauth2Provider.WriteAccessError(ctx, rw, accessRequest, err)
		return
	}

	// All done, send the response.
	oauth2Provider.WriteAccessResponse(ctx, rw, accessRequest, response)

	// The client has a valid access token now
}

func someResourceProviderHandlerFunc(rw http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	requiredScope := "blogposts.create"

	_, ar, err := oauth2Provider.IntrospectToken(ctx, fosite.AccessTokenFromRequest(req), fosite.AccessToken, new(fosite.DefaultSession), requiredScope)
	if err != nil {
		// ...
	}

	// If no error occurred the token + scope is valid and you have access to:
	// ar.GetClient().GetID(), ar.GetGrantedScopes(), ar.GetScopes(), ar.GetSession().UserID, ar.GetRequestedAt(), ...
}
Code Examples

Fosite provides integration tests as well as a http server example:

Fosite ships with an example app that runs in your browser: Example app.
If you want to check out how to enable specific handlers, check out the integration tests.
If you have working examples yourself, please share them with us!

Example Storage Implementation

Fosite does not ship a storage implementation. This is intended, because requirements vary with every environment. You can find a reference implementation at storage/memory.go. This storage fulfills requirements from all OAuth2 and OpenID Connect handlers.

Extensible handlers

OAuth2 is a framework. Fosite mimics this behaviour by enabling you to replace existing or create new OAuth2 handlers. Of course, fosite ships handlers for all OAuth2 and OpenID Connect flows.

Fosite OAuth2 Core Handlers implement the Client Credentials Grant, Resource Owner Password Credentials Grant, Implicit Grant, Authorization Code Grant, Refresh Token Grant
Fosite OpenID Connect Handlers implement the Authentication using the Authorization Code Flow, Authentication using the Implicit Flow, Authentication using the Hybrid Flow
This section is missing documentation and we welcome any contributions in that direction.

JWT Introspection

Please note that when using the OAuth2StatelessJWTIntrospectionFactory access token revocation is not possible.

Contribute

You need git and golang installed on your system.

go get -d github.com/ory/fosite
cd $GOPATH/src/github.com/ory/fosite
git status
git remote add myfork <url-to-your-fork>
go test ./...
Simple, right? Now you are ready to go! Make sure to run go test ./... often, detecting problems with your code rather sooner than later. Please read [CONTRIBUTE.md] before creating pull requests and issues.

Refresh mock objects

Run ./generate-mocks.sh in fosite's root directory or run the contents of [generate-mocks.sh] in a shell.

Hall of Fame

This place is reserved for the fearless bug hunters, reviewers and contributors (alphabetical order).

agtorre: contributions, participations.
danielchatfield: contributions, participations.
leetal: contributions, participations.
jrossiter: contributions, participations.
jrossiter: contributions, participations.
danilobuerger: contributions, participations.
Find out more about the author of Fosite and Hydra, and the Ory Company.

Collapse ▴
 Documentation ¶
Index ¶
Constants
Variables
func AccessTokenFromRequest(req *http.Request) string
func AddLocalizerToErr(catalog i18n.MessageCatalog, err error, requester Requester) error
func AddLocalizerToErrWithLang(catalog i18n.MessageCatalog, lang language.Tag, err error) error
func DefaultAudienceMatchingStrategy(haystack []string, needle []string) error
func EscapeJSONString(str string) string
func ExactAudienceMatchingStrategy(haystack []string, needle []string) error
func ExactScopeStrategy(haystack []string, needle string) bool
func GetAudiences(form url.Values) []string
func GetEffectiveLifespan(c Client, gt GrantType, tt TokenType, fallback time.Duration) time.Duration
func GetPostFormHTMLTemplate(ctx context.Context, f *Fosite) *template.Template
func HierarchicScopeStrategy(haystack []string, needle string) bool
func IsLocalhost(redirectURI *url.URL) bool
func IsRedirectURISecure(ctx context.Context, redirectURI *url.URL) bool
func IsRedirectURISecureStrict(ctx context.Context, redirectURI *url.URL) bool
func IsValidRedirectURI(redirectURI *url.URL) bool
func JKWKSFetcherWithDefaultTTL(ttl time.Duration) func(*DefaultJWKSFetcherStrategy)
func JWKSFetcherWithCache(cache *ristretto.Cache[string, *jose.JSONWebKeySet]) func(*DefaultJWKSFetcherStrategy)
func JWKSFetcherWithHTTPClient(client *retryablehttp.Client) func(*DefaultJWKSFetcherStrategy)
func JWKSFetcherWithHTTPClientSource(clientSourceFunc func(ctx context.Context) *retryablehttp.Client) func(*DefaultJWKSFetcherStrategy)
func MatchRedirectURIWithClientRedirectURIs(rawurl string, client Client) (*url.URL, error)
func NewContext() context.Context
func RemoveEmpty(args []string) (ret []string)
func StringInSlice(needle string, haystack []string) bool
func URLSetFragment(source *url.URL, fragment url.Values)DEPRECATED
func WildcardScopeStrategy(matchers []string, needle string) bool
func WriteAuthorizeFormPostResponse(redirectURL string, parameters url.Values, template *template.Template, ...)
type AccessRequest
func NewAccessRequest(session Session) *AccessRequest
func (a *AccessRequest) GetGrantTypes() Arguments
type AccessRequester
type AccessResponder
type AccessResponse
func NewAccessResponse() *AccessResponse
func (a *AccessResponse) GetAccessToken() string
func (a *AccessResponse) GetExtra(key string) interface{}
func (a *AccessResponse) GetTokenType() string
func (a *AccessResponse) SetAccessToken(token string)
func (a *AccessResponse) SetExpiresIn(expiresIn time.Duration)
func (a *AccessResponse) SetExtra(key string, value interface{})
func (a *AccessResponse) SetScopes(scopes Arguments)
func (a *AccessResponse) SetTokenType(name string)
func (a *AccessResponse) ToMap() map[string]interface{}
type AccessTokenIssuerProvider
type AccessTokenLifespanProvider
type AllowedPromptValuesProvider
type AllowedPromptsProvider
type Arguments
func (r Arguments) Exact(name string) boolDEPRECATED
func (r Arguments) ExactOne(name string) bool
func (r Arguments) Has(items ...string) bool
func (r Arguments) HasOneOf(items ...string) bool
func (r Arguments) Matches(items ...string) bool
func (r Arguments) MatchesExact(items ...string) bool
type AudienceMatchingStrategy
type AudienceStrategyProvider
type AuthorizeCodeLifespanProvider
type AuthorizeEndpointHandler
type AuthorizeEndpointHandlers
func (a *AuthorizeEndpointHandlers) Append(h AuthorizeEndpointHandler)
type AuthorizeEndpointHandlersProvider
type AuthorizeRequest
func NewAuthorizeRequest() *AuthorizeRequest
func (d *AuthorizeRequest) DidHandleAllResponseTypes() bool
func (d *AuthorizeRequest) GetDefaultResponseMode() ResponseModeType
func (d *AuthorizeRequest) GetRedirectURI() *url.URL
func (d *AuthorizeRequest) GetResponseMode() ResponseModeType
func (d *AuthorizeRequest) GetResponseTypes() Arguments
func (d *AuthorizeRequest) GetState() string
func (d *AuthorizeRequest) IsRedirectURIValid() bool
func (d *AuthorizeRequest) SetDefaultResponseMode(defaultResponseMode ResponseModeType)
func (d *AuthorizeRequest) SetResponseTypeHandled(name string)
type AuthorizeRequester
type AuthorizeResponder
type AuthorizeResponse
func NewAuthorizeResponse() *AuthorizeResponse
func (a *AuthorizeResponse) AddHeader(key, value string)
func (a *AuthorizeResponse) AddParameter(key, value string)
func (a *AuthorizeResponse) GetCode() string
func (a *AuthorizeResponse) GetHeader() http.Header
func (a *AuthorizeResponse) GetParameters() url.Values
type BCrypt
func (b *BCrypt) Compare(ctx context.Context, hash, data []byte) error
func (b *BCrypt) Hash(ctx context.Context, data []byte) ([]byte, error)
type BCryptCostProvider
type Client
type ClientAuthenticationStrategy
type ClientAuthenticationStrategyProvider
type ClientLifespanConfig
type ClientManager
type ClientWithCustomTokenLifespans
type ClientWithSecretRotation
type Config
func (c *Config) EnforcePushedAuthorize(ctx context.Context) bool
func (c *Config) GetAccessTokenIssuer(ctx context.Context) string
func (c *Config) GetAccessTokenLifespan(_ context.Context) time.Duration
func (c *Config) GetAllowedPrompts(_ context.Context) []string
func (c *Config) GetAudienceStrategy(_ context.Context) AudienceMatchingStrategy
func (c *Config) GetAuthorizeCodeLifespan(_ context.Context) time.Duration
func (c *Config) GetAuthorizeEndpointHandlers(ctx context.Context) AuthorizeEndpointHandlers
func (c *Config) GetBCryptCost(_ context.Context) int
func (c *Config) GetClientAuthenticationStrategy(_ context.Context) ClientAuthenticationStrategy
func (c *Config) GetDisableRefreshTokenValidation(_ context.Context) bool
func (c *Config) GetEnablePKCEPlainChallengeMethod(ctx context.Context) bool
func (c *Config) GetEnforcePKCE(ctx context.Context) bool
func (c *Config) GetEnforcePKCEForPublicClients(ctx context.Context) bool
func (c *Config) GetFormPostHTMLTemplate(ctx context.Context) *template.Template
func (c *Config) GetGlobalSecret(ctx context.Context) ([]byte, error)
func (c *Config) GetGrantTypeJWTBearerCanSkipClientAuth(ctx context.Context) bool
func (c *Config) GetGrantTypeJWTBearerIDOptional(ctx context.Context) bool
func (c *Config) GetGrantTypeJWTBearerIssuedDateOptional(ctx context.Context) bool
func (c *Config) GetHMACHasher(ctx context.Context) func() hash.Hash
func (c *Config) GetHTTPClient(ctx context.Context) *retryablehttp.Client
func (c *Config) GetIDTokenIssuer(ctx context.Context) string
func (c *Config) GetIDTokenLifespan(_ context.Context) time.Duration
func (c *Config) GetJWKSFetcherStrategy(_ context.Context) JWKSFetcherStrategy
func (c *Config) GetJWTMaxDuration(_ context.Context) time.Duration
func (c *Config) GetJWTScopeField(ctx context.Context) jwt.JWTScopeFieldEnum
func (c *Config) GetMessageCatalog(ctx context.Context) i18n.MessageCatalog
func (c *Config) GetMinParameterEntropy(_ context.Context) int
func (c *Config) GetOmitRedirectScopeParam(ctx context.Context) bool
func (c *Config) GetPushedAuthorizeContextLifespan(ctx context.Context) time.Duration
func (c *Config) GetPushedAuthorizeEndpointHandlers(ctx context.Context) PushedAuthorizeEndpointHandlers
func (c *Config) GetPushedAuthorizeRequestURIPrefix(ctx context.Context) string
func (c *Config) GetRedirectSecureChecker(_ context.Context) func(context.Context, *url.URL) bool
func (c *Config) GetRefreshTokenLifespan(_ context.Context) time.Duration
func (c *Config) GetRefreshTokenScopes(_ context.Context) []string
func (c *Config) GetResponseModeHandlerExtension(ctx context.Context) ResponseModeHandler
func (c *Config) GetRevocationHandlers(ctx context.Context) RevocationHandlers
func (c *Config) GetRotatedGlobalSecrets(ctx context.Context) ([][]byte, error)
func (c *Config) GetSanitationWhiteList(ctx context.Context) []string
func (c *Config) GetScopeStrategy(_ context.Context) ScopeStrategy
func (c *Config) GetSecretsHasher(ctx context.Context) Hasher
func (c *Config) GetSendDebugMessagesToClients(ctx context.Context) bool
func (c *Config) GetTokenEndpointHandlers(ctx context.Context) TokenEndpointHandlers
func (c *Config) GetTokenEntropy(_ context.Context) int
func (c *Config) GetTokenIntrospectionHandlers(ctx context.Context) TokenIntrospectionHandlers
func (c *Config) GetTokenURLs(ctx context.Context) []string
func (c *Config) GetUseLegacyErrorFormat(ctx context.Context) bool
func (c *Config) GetVerifiableCredentialsNonceLifespan(_ context.Context) time.Duration
type Configurator
type ContextKey
type DefaultClient
func (c *DefaultClient) GetAudience() Arguments
func (c *DefaultClient) GetGrantTypes() Arguments
func (c *DefaultClient) GetHashedSecret() []byte
func (c *DefaultClient) GetID() string
func (c *DefaultClient) GetRedirectURIs() []string
func (c *DefaultClient) GetResponseTypes() Arguments
func (c *DefaultClient) GetRotatedHashes() [][]byte
func (c *DefaultClient) GetScopes() Arguments
func (c *DefaultClient) IsPublic() bool
type DefaultClientWithCustomTokenLifespans
func (c *DefaultClientWithCustomTokenLifespans) GetEffectiveLifespan(gt GrantType, tt TokenType, fallback time.Duration) time.Duration
func (c *DefaultClientWithCustomTokenLifespans) GetTokenLifespans() *ClientLifespanConfig
func (c *DefaultClientWithCustomTokenLifespans) SetTokenLifespans(lifespans *ClientLifespanConfig)
type DefaultJWKSFetcherStrategy
func (s *DefaultJWKSFetcherStrategy) Resolve(ctx context.Context, location string, ignoreCache bool) (*jose.JSONWebKeySet, error)
func (s *DefaultJWKSFetcherStrategy) WaitForCache()
type DefaultOpenIDConnectClient
func (c *DefaultOpenIDConnectClient) GetJSONWebKeys() *jose.JSONWebKeySet
func (c *DefaultOpenIDConnectClient) GetJSONWebKeysURI() string
func (c *DefaultOpenIDConnectClient) GetRequestObjectSigningAlgorithm() string
func (c *DefaultOpenIDConnectClient) GetRequestURIs() []string
func (c *DefaultOpenIDConnectClient) GetTokenEndpointAuthMethod() string
func (c *DefaultOpenIDConnectClient) GetTokenEndpointAuthSigningAlgorithm() string
type DefaultResponseModeClient
func (c *DefaultResponseModeClient) GetResponseModes() []ResponseModeType
type DefaultResponseModeHandler
func NewDefaultResponseModeHandler() *DefaultResponseModeHandler
func (d *DefaultResponseModeHandler) ResponseModes() ResponseModeTypes
func (d *DefaultResponseModeHandler) WriteAuthorizeError(ctx context.Context, rw http.ResponseWriter, ar AuthorizeRequester, err error)
func (d *DefaultResponseModeHandler) WriteAuthorizeResponse(ctx context.Context, rw http.ResponseWriter, ar AuthorizeRequester, ...)
type DefaultSession
func (s *DefaultSession) Clone() Session
func (s *DefaultSession) GetExpiresAt(key TokenType) time.Time
func (s *DefaultSession) GetExtraClaims() map[string]interface{}
func (s *DefaultSession) GetSubject() string
func (s *DefaultSession) GetUsername() string
func (s *DefaultSession) SetExpiresAt(key TokenType, exp time.Time)
func (s *DefaultSession) SetSubject(subject string)
type DisableRefreshTokenValidationProvider
type EnablePKCEPlainChallengeMethodProvider
type EnforcePKCEForPublicClientsProvider
type EnforcePKCEProvider
type ExtraClaimsSession
type FormPostHTMLTemplateProvider
type Fosite
func NewOAuth2Provider(s Storage, c Configurator) *Fosite
func (f *Fosite) AuthenticateClient(ctx context.Context, r *http.Request, form url.Values) (Client, error)
func (f *Fosite) DefaultClientAuthenticationStrategy(ctx context.Context, r *http.Request, form url.Values) (Client, error)
func (f *Fosite) GetMinParameterEntropy(ctx context.Context) int
func (f *Fosite) IntrospectToken(ctx context.Context, token string, tokenUse TokenUse, session Session, ...) (_ TokenUse, _ AccessRequester, err error)
func (f *Fosite) NewAccessRequest(ctx context.Context, r *http.Request, session Session) (_ AccessRequester, err error)
func (f *Fosite) NewAccessResponse(ctx context.Context, requester AccessRequester) (_ AccessResponder, err error)
func (f *Fosite) NewAuthorizeRequest(ctx context.Context, r *http.Request) (_ AuthorizeRequester, err error)
func (f *Fosite) NewAuthorizeResponse(ctx context.Context, ar AuthorizeRequester, session Session) (_ AuthorizeResponder, err error)
func (f *Fosite) NewIntrospectionRequest(ctx context.Context, r *http.Request, session Session) (_ IntrospectionResponder, err error)
func (f *Fosite) NewPushedAuthorizeRequest(ctx context.Context, r *http.Request) (_ AuthorizeRequester, err error)
func (f *Fosite) NewPushedAuthorizeResponse(ctx context.Context, ar AuthorizeRequester, session Session) (_ PushedAuthorizeResponder, err error)
func (f *Fosite) NewRevocationRequest(ctx context.Context, r *http.Request) (err error)
func (f *Fosite) ParseResponseMode(ctx context.Context, r *http.Request, request *AuthorizeRequest) error
func (f *Fosite) ResponseModeHandler(ctx context.Context) ResponseModeHandler
func (f *Fosite) WriteAccessError(ctx context.Context, rw http.ResponseWriter, req AccessRequester, err error)
func (f *Fosite) WriteAccessResponse(ctx context.Context, rw http.ResponseWriter, requester AccessRequester, ...)
func (f *Fosite) WriteAuthorizeError(ctx context.Context, rw http.ResponseWriter, ar AuthorizeRequester, err error)
func (f *Fosite) WriteAuthorizeResponse(ctx context.Context, rw http.ResponseWriter, ar AuthorizeRequester, ...)
func (f *Fosite) WriteIntrospectionError(ctx context.Context, rw http.ResponseWriter, err error)
func (f *Fosite) WriteIntrospectionResponse(ctx context.Context, rw http.ResponseWriter, r IntrospectionResponder)
func (f *Fosite) WritePushedAuthorizeError(ctx context.Context, rw http.ResponseWriter, ar AuthorizeRequester, err error)
func (f *Fosite) WritePushedAuthorizeResponse(ctx context.Context, rw http.ResponseWriter, ar AuthorizeRequester, ...)
func (f *Fosite) WriteRevocationResponse(ctx context.Context, rw http.ResponseWriter, err error)
type G11NContext
type GetJWTMaxDurationProvider
type GetSecretsHashingProvider
type GlobalSecretProvider
type GrantType
type GrantTypeJWTBearerCanSkipClientAuthProvider
type GrantTypeJWTBearerIDOptionalProvider
type GrantTypeJWTBearerIssuedDateOptionalProvider
type HMACHashingProvider
type HTTPClientProvider
type Hasher
type IDTokenIssuerProvider
type IDTokenLifespanProvider
type IntrospectionResponder
type IntrospectionResponse
func (r *IntrospectionResponse) GetAccessRequester() AccessRequester
func (r *IntrospectionResponse) GetAccessTokenType() string
func (r *IntrospectionResponse) GetTokenUse() TokenUse
func (r *IntrospectionResponse) IsActive() bool
type JWKSFetcherStrategy
func NewDefaultJWKSFetcherStrategy(opts ...func(*DefaultJWKSFetcherStrategy)) JWKSFetcherStrategy
type JWKSFetcherStrategyProvider
type JWTScopeFieldProvider
type MessageCatalogProvider
type MinParameterEntropyProvider
type OAuth2Provider
type OmitRedirectScopeParamProvider
type OpenIDConnectClient
type PARStorage
type PushedAuthorizeEndpointHandler
type PushedAuthorizeEndpointHandlers
func (a *PushedAuthorizeEndpointHandlers) Append(h PushedAuthorizeEndpointHandler)
type PushedAuthorizeRequestConfigProvider
type PushedAuthorizeRequestHandlersProvider
type PushedAuthorizeResponder
type PushedAuthorizeResponse
func (a *PushedAuthorizeResponse) AddHeader(key, value string)
func (a *PushedAuthorizeResponse) GetExpiresIn() int
func (a *PushedAuthorizeResponse) GetExtra(key string) interface{}
func (a *PushedAuthorizeResponse) GetHeader() http.Header
func (a *PushedAuthorizeResponse) GetRequestURI() string
func (a *PushedAuthorizeResponse) SetExpiresIn(seconds int)
func (a *PushedAuthorizeResponse) SetExtra(key string, value interface{})
func (a *PushedAuthorizeResponse) SetRequestURI(requestURI string)
func (a *PushedAuthorizeResponse) ToMap() map[string]interface{}
type RFC6749Error
func ErrorToRFC6749Error(err error) *RFC6749Error
func (e *RFC6749Error) Cause() error
func (e *RFC6749Error) Debug() string
func (e RFC6749Error) Error() string
func (e *RFC6749Error) GetDescription() string
func (e RFC6749Error) Is(err error) bool
func (e RFC6749Error) MarshalJSON() ([]byte, error)
func (e *RFC6749Error) Reason() string
func (e *RFC6749Error) RequestID() string
func (e *RFC6749Error) Sanitize() *RFC6749ErrorDEPRECATED
func (e *RFC6749Error) StackTrace() (trace errors.StackTrace)
func (e *RFC6749Error) Status() string
func (e *RFC6749Error) StatusCode() int
func (e *RFC6749Error) ToValues() url.Values
func (e *RFC6749Error) UnmarshalJSON(b []byte) error
func (e RFC6749Error) Unwrap() error
func (e *RFC6749Error) WithDebug(debug string) *RFC6749Error
func (e *RFC6749Error) WithDebugf(debug string, args ...interface{}) *RFC6749Error
func (e *RFC6749Error) WithDescription(description string) *RFC6749Error
func (e *RFC6749Error) WithExposeDebug(exposeDebug bool) *RFC6749Error
func (e *RFC6749Error) WithHint(hint string) *RFC6749Error
func (e *RFC6749Error) WithHintIDOrDefaultf(ID string, def string, args ...interface{}) *RFC6749Error
func (e *RFC6749Error) WithHintTranslationID(ID string) *RFC6749Error
func (e *RFC6749Error) WithHintf(hint string, args ...interface{}) *RFC6749Error
func (e RFC6749Error) WithLegacyFormat(useLegacyFormat bool) *RFC6749Error
func (e *RFC6749Error) WithLocalizer(catalog i18n.MessageCatalog, lang language.Tag) *RFC6749Error
func (e *RFC6749Error) WithTrace(err error) *RFC6749Error
func (e RFC6749Error) WithWrap(cause error) *RFC6749Error
func (e *RFC6749Error) Wrap(err error)
type RFC6749ErrorJson
type RedirectSecureCheckerProvider
type RefreshTokenLifespanProvider
type RefreshTokenScopesProvider
type Request
func NewRequest() *Request
func (a *Request) AppendRequestedAudience(audience string)
func (a *Request) AppendRequestedScope(scope string)
func (a *Request) GetClient() Client
func (a *Request) GetGrantedAudience() Arguments
func (a *Request) GetGrantedScopes() Arguments
func (a *Request) GetID() string
func (a *Request) GetLang() language.Tag
func (a *Request) GetRequestForm() url.Values
func (a *Request) GetRequestedAt() time.Time
func (a *Request) GetRequestedAudience() (audience Arguments)
func (a *Request) GetRequestedScopes() Arguments
func (a *Request) GetSession() Session
func (a *Request) GrantAudience(audience string)
func (a *Request) GrantScope(scope string)
func (a *Request) Merge(request Requester)
func (a *Request) Sanitize(allowedParameters []string) Requester
func (a *Request) SetID(id string)
func (a *Request) SetRequestedAudience(s Arguments)
func (a *Request) SetRequestedScopes(s Arguments)
func (a *Request) SetSession(session Session)
type Requester
type ResponseModeClient
type ResponseModeHandler
type ResponseModeHandlerExtensionProvider
type ResponseModeType
type ResponseModeTypes
func (rs ResponseModeTypes) Has(item ResponseModeType) bool
type RevocationHandler
type RevocationHandlers
func (t *RevocationHandlers) Append(h RevocationHandler)
type RevocationHandlersProvider
type RotatedGlobalSecretsProvider
type SanitationAllowedProvider
type ScopeStrategy
type ScopeStrategyProvider
type SendDebugMessagesToClientsProvider
type Session
type Storage
type TokenEndpointHandler
type TokenEndpointHandlers
func (t *TokenEndpointHandlers) Append(h TokenEndpointHandler)
type TokenEndpointHandlersProvider
type TokenEntropyProvider
type TokenIntrospectionHandlers
func (t *TokenIntrospectionHandlers) Append(h TokenIntrospector)
type TokenIntrospectionHandlersProvider
type TokenIntrospector
type TokenType
type TokenURLProvider
type TokenUse
type UseLegacyErrorFormatProvider
type VerifiableCredentialsNonceLifespanProvider
Constants ¶
View Source
const (
	ResponseModeDefault  = ResponseModeType("")
	ResponseModeFormPost = ResponseModeType("form_post")
	ResponseModeQuery    = ResponseModeType("query")
	ResponseModeFragment = ResponseModeType("fragment")
)
View Source
const (
	RequestContextKey           = ContextKey("request")
	AccessRequestContextKey     = ContextKey("accessRequest")
	AccessResponseContextKey    = ContextKey("accessResponse")
	AuthorizeRequestContextKey  = ContextKey("authorizeRequest")
	AuthorizeResponseContextKey = ContextKey("authorizeResponse")
	// PushedAuthorizeResponseContextKey is the response context
	PushedAuthorizeResponseContextKey = ContextKey("pushedAuthorizeResponse")
)
View Source
const (
	AccessToken   TokenType = "access_token"
	RefreshToken  TokenType = "refresh_token"
	AuthorizeCode TokenType = "authorize_code"
	IDToken       TokenType = "id_token"
	// PushedAuthorizeRequestContext represents the PAR context object
	PushedAuthorizeRequestContext TokenType = "par_context"

	GrantTypeImplicit          GrantType = "implicit"
	GrantTypeRefreshToken      GrantType = "refresh_token"
	GrantTypeAuthorizationCode GrantType = "authorization_code"
	GrantTypePassword          GrantType = "password"
	GrantTypeClientCredentials GrantType = "client_credentials"
	GrantTypeJWTBearer         GrantType = "urn:ietf:params:oauth:grant-type:jwt-bearer" //nolint:gosec // this is not a hardcoded credential

	BearerAccessToken string = "bearer"
)
View Source
const (
	ErrorPARNotSupported           = "The OAuth 2.0 provider does not support Pushed Authorization Requests"
	DebugPARStorageInvalid         = "'PARStorage' not implemented"
	DebugPARConfigMissing          = "'PushedAuthorizeRequestConfigProvider' not implemented"
	DebugPARRequestsHandlerMissing = "'PushedAuthorizeRequestHandlersProvider' not implemented"
)
View Source
const DefaultBCryptWorkFactor = 12
View Source
const MinParameterEntropy = 8
Variables ¶
View Source
var (
	// ErrInvalidatedAuthorizeCode is an error indicating that an authorization code has been
	// used previously.
	ErrInvalidatedAuthorizeCode = stderr.New("Authorization code has ben invalidated")
	// ErrSerializationFailure is an error indicating that the transactional capable storage could not guarantee
	// consistency of Update & Delete operations on the same rows between multiple sessions.
	ErrSerializationFailure = &RFC6749Error{
		ErrorField:       errUnknownErrorName,
		DescriptionField: "The request could not be completed because another request is competing for the same resource.",
		CodeField:        http.StatusConflict,
	}
	ErrUnknownRequest = &RFC6749Error{
		ErrorField:       errUnknownErrorName,
		DescriptionField: "The handler is not responsible for this request.",
		CodeField:        http.StatusBadRequest,
	}
	ErrRequestForbidden = &RFC6749Error{
		ErrorField:       errRequestForbidden,
		DescriptionField: "The request is not allowed.",
		HintField:        "You are not allowed to perform this action.",
		CodeField:        http.StatusForbidden,
	}
	ErrInvalidRequest = &RFC6749Error{
		ErrorField:       errInvalidRequestName,
		DescriptionField: "The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed.",
		HintField:        "Make sure that the various parameters are correct, be aware of case sensitivity and trim your parameters. Make sure that the client you are using has exactly whitelisted the redirect_uri you specified.",
		CodeField:        http.StatusBadRequest,
	}
	ErrUnauthorizedClient = &RFC6749Error{
		ErrorField:       errUnauthorizedClientName,
		DescriptionField: "The client is not authorized to request a token using this method.",
		HintField:        "Make sure that client id and secret are correctly specified and that the client exists.",
		CodeField:        http.StatusBadRequest,
	}
	ErrAccessDenied = &RFC6749Error{
		ErrorField:       errAccessDeniedName,
		DescriptionField: "The resource owner or authorization server denied the request.",
		HintField:        "Make sure that the request you are making is valid. Maybe the credential or request parameters you are using are limited in scope or otherwise restricted.",
		CodeField:        http.StatusForbidden,
	}
	ErrUnsupportedResponseType = &RFC6749Error{
		ErrorField:       errUnsupportedResponseTypeName,
		DescriptionField: "The authorization server does not support obtaining a token using this method.",
		CodeField:        http.StatusBadRequest,
	}
	ErrUnsupportedResponseMode = &RFC6749Error{
		ErrorField:       errUnsupportedResponseModeName,
		DescriptionField: "The authorization server does not support obtaining a response using this response mode.",
		CodeField:        http.StatusBadRequest,
	}
	ErrInvalidScope = &RFC6749Error{
		ErrorField:       errInvalidScopeName,
		DescriptionField: "The requested scope is invalid, unknown, or malformed.",
		CodeField:        http.StatusBadRequest,
	}
	ErrServerError = &RFC6749Error{
		ErrorField:       errServerErrorName,
		DescriptionField: "The authorization server encountered an unexpected condition that prevented it from fulfilling the request.",
		CodeField:        http.StatusInternalServerError,
	}
	ErrTemporarilyUnavailable = &RFC6749Error{
		ErrorField:       errTemporarilyUnavailableName,
		DescriptionField: "The authorization server is currently unable to handle the request due to a temporary overloading or maintenance of the server.",
		CodeField:        http.StatusServiceUnavailable,
	}
	ErrUnsupportedGrantType = &RFC6749Error{
		ErrorField:       errUnsupportedGrantTypeName,
		DescriptionField: "The authorization grant type is not supported by the authorization server.",
		CodeField:        http.StatusBadRequest,
	}
	ErrInvalidGrant = &RFC6749Error{
		ErrorField:       errInvalidGrantName,
		DescriptionField: "The provided authorization grant (e.g., authorization code, resource owner credentials) or refresh token is invalid, expired, revoked, does not match the redirection URI used in the authorization request, or was issued to another client.",
		CodeField:        http.StatusBadRequest,
	}
	ErrInvalidClient = &RFC6749Error{
		ErrorField:       errInvalidClientName,
		DescriptionField: "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method).",
		CodeField:        http.StatusUnauthorized,
	}
	ErrInvalidState = &RFC6749Error{
		ErrorField:       errInvalidStateName,
		DescriptionField: "The state is missing or does not have enough characters and is therefore considered too weak.",
		CodeField:        http.StatusBadRequest,
	}
	ErrMisconfiguration = &RFC6749Error{
		ErrorField:       errMisconfigurationName,
		DescriptionField: "The request failed because of an internal error that is probably caused by misconfiguration.",
		CodeField:        http.StatusInternalServerError,
	}
	ErrInsufficientEntropy = &RFC6749Error{
		ErrorField:       errInsufficientEntropyName,
		DescriptionField: "The request used a security parameter (e.g., anti-replay, anti-csrf) with insufficient entropy.",
		CodeField:        http.StatusBadRequest,
	}
	ErrNotFound = &RFC6749Error{
		ErrorField:       errNotFoundName,
		DescriptionField: "Could not find the requested resource(s).",
		CodeField:        http.StatusNotFound,
	}
	ErrRequestUnauthorized = &RFC6749Error{
		ErrorField:       errRequestUnauthorizedName,
		DescriptionField: "The request could not be authorized.",
		HintField:        "Check that you provided valid credentials in the right format.",
		CodeField:        http.StatusUnauthorized,
	}
	ErrTokenSignatureMismatch = &RFC6749Error{
		ErrorField:       errTokenSignatureMismatchName,
		DescriptionField: "Token signature mismatch.",
		HintField:        "Check that you provided  a valid token in the right format.",
		CodeField:        http.StatusBadRequest,
	}
	ErrInvalidTokenFormat = &RFC6749Error{
		ErrorField:       errInvalidTokenFormatName,
		DescriptionField: "Invalid token format.",
		HintField:        "Check that you provided a valid token in the right format.",
		CodeField:        http.StatusBadRequest,
	}
	ErrTokenExpired = &RFC6749Error{
		ErrorField:       errTokenExpiredName,
		DescriptionField: "Token expired.",
		HintField:        "The token expired.",
		CodeField:        http.StatusUnauthorized,
	}
	ErrScopeNotGranted = &RFC6749Error{
		ErrorField:       errScopeNotGrantedName,
		DescriptionField: "The token was not granted the requested scope.",
		HintField:        "The resource owner did not grant the requested scope.",
		CodeField:        http.StatusForbidden,
	}
	ErrTokenClaim = &RFC6749Error{
		ErrorField:       errTokenClaimName,
		DescriptionField: "The token failed validation due to a claim mismatch.",
		HintField:        "One or more token claims failed validation.",
		CodeField:        http.StatusUnauthorized,
	}
	ErrInactiveToken = &RFC6749Error{
		ErrorField:       errTokenInactiveName,
		DescriptionField: "Token is inactive because it is malformed, expired or otherwise invalid.",
		HintField:        "Token validation failed.",
		CodeField:        http.StatusUnauthorized,
	}
	ErrLoginRequired = &RFC6749Error{
		ErrorField:       errLoginRequired,
		DescriptionField: "The Authorization Server requires End-User authentication.",
		CodeField:        http.StatusBadRequest,
	}
	ErrInteractionRequired = &RFC6749Error{
		DescriptionField: "The Authorization Server requires End-User interaction of some form to proceed.",
		ErrorField:       errInteractionRequired,
		CodeField:        http.StatusBadRequest,
	}
	ErrConsentRequired = &RFC6749Error{
		DescriptionField: "The Authorization Server requires End-User consent.",
		ErrorField:       errConsentRequired,
		CodeField:        http.StatusBadRequest,
	}
	ErrRequestNotSupported = &RFC6749Error{
		DescriptionField: "The OP does not support use of the request parameter.",
		ErrorField:       errRequestNotSupportedName,
		CodeField:        http.StatusBadRequest,
	}
	ErrRequestURINotSupported = &RFC6749Error{
		DescriptionField: "The OP does not support use of the request_uri parameter.",
		ErrorField:       errRequestURINotSupportedName,
		CodeField:        http.StatusBadRequest,
	}
	ErrRegistrationNotSupported = &RFC6749Error{
		DescriptionField: "The OP does not support use of the registration parameter.",
		ErrorField:       errRegistrationNotSupportedName,
		CodeField:        http.StatusBadRequest,
	}
	ErrInvalidRequestURI = &RFC6749Error{
		DescriptionField: "The request_uri in the Authorization Request returns an error or contains invalid data.",
		ErrorField:       errInvalidRequestURI,
		CodeField:        http.StatusBadRequest,
	}
	ErrInvalidRequestObject = &RFC6749Error{
		DescriptionField: "The request parameter contains an invalid Request Object.",
		ErrorField:       errInvalidRequestObject,
		CodeField:        http.StatusBadRequest,
	}
	ErrJTIKnown = &RFC6749Error{
		DescriptionField: "The jti was already used.",
		ErrorField:       errJTIKnownName,
		CodeField:        http.StatusBadRequest,
	}
)
View Source
var DefaultFormPostTemplate = template.Must(template.New("form_post").Parse(`<html>
   <head>
      <title>Submit This Form</title>
   </head>
   <body onload="javascript:document.forms[0].submit()">
      <form method="post" action="{{ .RedirURL }}">
         {{ range $key,$value := .Parameters }}
            {{ range $parameter:= $value}}
		      <input type="hidden" name="{{$key}}" value="{{$parameter}}"/>
            {{end}}
         {{ end }}
      </form>
   </body>
</html>`))
Functions ¶
func AccessTokenFromRequest ¶
added in v0.2.0
func AccessTokenFromRequest(req *http.Request) string
func AddLocalizerToErr ¶
added in v0.41.0
func AddLocalizerToErr(catalog i18n.MessageCatalog, err error, requester Requester) error
AddLocalizerToErr augments the error object with the localizer based on the language set in the requester object. This is primarily required for response writers like introspection that do not take in the requester in the Write* function that produces the translated message. See - WriteIntrospectionError, for example.

func AddLocalizerToErrWithLang ¶
added in v0.41.0
func AddLocalizerToErrWithLang(catalog i18n.MessageCatalog, lang language.Tag, err error) error
AddLocalizerToErrWithLang augments the error object with the localizer based on the language passed in. This is primarily required for response writers like introspection that do not take in the requester in the Write* function that produces the translated message. See - WriteIntrospectionError, for example.

func DefaultAudienceMatchingStrategy ¶
added in v0.27.0
func DefaultAudienceMatchingStrategy(haystack []string, needle []string) error
func EscapeJSONString ¶
added in v0.34.0
func EscapeJSONString(str string) string
EscapeJSONString does a poor man's JSON encoding. Useful when we do not want to use full JSON encoding because we just had an error doing the JSON encoding. The characters that MUST be escaped: quotation mark, reverse solidus, and the control characters (U+0000 through U+001F). See: https://tools.ietf.org/html/std90#section-7

func ExactAudienceMatchingStrategy ¶
added in v0.36.0
func ExactAudienceMatchingStrategy(haystack []string, needle []string) error
ExactAudienceMatchingStrategy does not assume that audiences are URIs, but compares strings as-is and does matching with exact string comparison. It requires that all strings in "needle" are present in "haystack". Use this strategy when your audience values are not URIs (e.g., you use client IDs for audience and they are UUIDs or random strings).

func ExactScopeStrategy ¶
added in v0.17.1
func ExactScopeStrategy(haystack []string, needle string) bool
func GetAudiences ¶
added in v0.36.0
func GetAudiences(form url.Values) []string
GetAudiences allows audiences to be provided as repeated "audience" form parameter, or as a space-delimited "audience" form parameter if it is not repeated. RFC 8693 in section 2.1 specifies that multiple audience values should be multiple query parameters, while RFC 6749 says that that request parameter must not be included more than once (and thus why we use space-delimited value). This function tries to satisfy both. If "audience" form parameter is repeated, we do not split the value by space.

func GetEffectiveLifespan ¶
added in v0.43.0
func GetEffectiveLifespan(c Client, gt GrantType, tt TokenType, fallback time.Duration) time.Duration
GetEffectiveLifespan either maps GrantType x TokenType to the client's configured lifespan, or returns the fallback value.

func GetPostFormHTMLTemplate ¶
added in v0.36.0
func GetPostFormHTMLTemplate(ctx context.Context, f *Fosite) *template.Template
func HierarchicScopeStrategy ¶
added in v0.2.0
func HierarchicScopeStrategy(haystack []string, needle string) bool
func IsLocalhost ¶
added in v0.29.3
func IsLocalhost(redirectURI *url.URL) bool
func IsRedirectURISecure ¶
func IsRedirectURISecure(ctx context.Context, redirectURI *url.URL) bool
func IsRedirectURISecureStrict ¶
added in v0.35.1
func IsRedirectURISecureStrict(ctx context.Context, redirectURI *url.URL) bool
IsRedirectURISecureStrict is stricter than IsRedirectURISecure and it does not allow custom-scheme URLs because they can be hijacked for native apps. Use claimed HTTPS redirects instead. See discussion in https://github.com/ory/fosite/pull/489.

func IsValidRedirectURI ¶
func IsValidRedirectURI(redirectURI *url.URL) bool
IsValidRedirectURI validates a redirect_uri as specified in:

* https://tools.ietf.org/html/rfc6749#section-3.1.2

The redirection endpoint URI MUST be an absolute URI as defined by [RFC3986] Section 4.3.
The endpoint URI MUST NOT include a fragment component.
https://tools.ietf.org/html/rfc3986#section-4.3 absolute-URI = scheme ":" hier-part [ "?" query ]
https://tools.ietf.org/html/rfc6819#section-5.1.1
func JKWKSFetcherWithDefaultTTL ¶
added in v0.43.0
func JKWKSFetcherWithDefaultTTL(ttl time.Duration) func(*DefaultJWKSFetcherStrategy)
JKWKSFetcherWithDefaultTTL sets the default TTL for the cache.

func JWKSFetcherWithCache ¶
added in v0.43.0
func JWKSFetcherWithCache(cache *ristretto.Cache[string, *jose.JSONWebKeySet]) func(*DefaultJWKSFetcherStrategy)
JWKSFetcherWithCache sets the cache to use.

func JWKSFetcherWithHTTPClient ¶
added in v0.43.0
func JWKSFetcherWithHTTPClient(client *retryablehttp.Client) func(*DefaultJWKSFetcherStrategy)
JWKSFetcherWithHTTPClient sets the HTTP client to use.

func JWKSFetcherWithHTTPClientSource ¶
added in v0.43.0
func JWKSFetcherWithHTTPClientSource(clientSourceFunc func(ctx context.Context) *retryablehttp.Client) func(*DefaultJWKSFetcherStrategy)
JWKSFetcherWithHTTPClientSource sets the HTTP client source function to use.

func MatchRedirectURIWithClientRedirectURIs ¶
func MatchRedirectURIWithClientRedirectURIs(rawurl string, client Client) (*url.URL, error)
MatchRedirectURIWithClientRedirectURIs if the given uri is a registered redirect uri. Does not perform uri validation.

Considered specifications

https://tools.ietf.org/html/rfc6749#section-3.1.2.3 If multiple redirection URIs have been registered, if only part of the redirection URI has been registered, or if no redirection URI has been registered, the client MUST include a redirection URI with the authorization request using the "redirect_uri" request parameter.

When a redirection URI is included in an authorization request, the authorization server MUST compare and match the value received against at least one of the registered redirection URIs (or URI components) as defined in [RFC3986] Section 6, if any redirection URIs were registered. If the client registration included the full redirection URI, the authorization server MUST compare the two URIs using simple string comparison as defined in [RFC3986] Section 6.2.1.

* https://tools.ietf.org/html/rfc6819#section-4.4.1.7

The authorization server may also enforce the usage and validation of pre-registered redirect URIs (see Section 5.2.3.5). This will allow for early recognition of authorization "code" disclosure to counterfeit clients.
The attacker will need to use another redirect URI for its authorization process rather than the target web site because it needs to intercept the flow. So, if the authorization server associates the authorization "code" with the redirect URI of a particular end-user authorization and validates this redirect URI with the redirect URI passed to the token's endpoint, such an attack is detected (see Section 5.2.4.5).
func NewContext ¶
func NewContext() context.Context
func RemoveEmpty ¶
added in v0.32.1
func RemoveEmpty(args []string) (ret []string)
func StringInSlice ¶
func StringInSlice(needle string, haystack []string) bool
StringInSlice returns true if needle exists in haystack

func
URLSetFragment
DEPRECATED
added in v0.36.0
func WildcardScopeStrategy ¶
added in v0.11.0
func WildcardScopeStrategy(matchers []string, needle string) bool
func WriteAuthorizeFormPostResponse ¶
added in v0.36.0
func WriteAuthorizeFormPostResponse(redirectURL string, parameters url.Values, template *template.Template, rw io.Writer)
Types ¶
type AccessRequest ¶
type AccessRequest struct {
	GrantTypes       Arguments `json:"grantTypes" gorethink:"grantTypes"`
	HandledGrantType Arguments `json:"handledGrantType" gorethink:"handledGrantType"`

	Request
}
func NewAccessRequest ¶
func NewAccessRequest(session Session) *AccessRequest
func (*AccessRequest) GetGrantTypes ¶
func (a *AccessRequest) GetGrantTypes() Arguments
type AccessRequester ¶
type AccessRequester interface {
	// GetGrantType returns the requests grant type.
	GetGrantTypes() (grantTypes Arguments)

	Requester
}
AccessRequester is a token endpoint's request context.

type AccessResponder ¶
type AccessResponder interface {
	// SetExtra sets a key value pair for the access response.
	SetExtra(key string, value interface{})

	// GetExtra returns a key's value.
	GetExtra(key string) interface{}

	SetExpiresIn(time.Duration)

	SetScopes(scopes Arguments)

	// SetAccessToken sets the responses mandatory access token.
	SetAccessToken(token string)

	// SetTokenType set's the responses mandatory token type
	SetTokenType(tokenType string)

	// SetAccessToken returns the responses access token.
	GetAccessToken() (token string)

	// GetTokenType returns the responses token type.
	GetTokenType() (token string)

	// ToMap converts the response to a map.
	ToMap() map[string]interface{}
}
AccessResponder is a token endpoint's response.

type AccessResponse ¶
type AccessResponse struct {
	Extra       map[string]interface{}
	AccessToken string
	TokenType   string
}
func NewAccessResponse ¶
func NewAccessResponse() *AccessResponse
func (*AccessResponse) GetAccessToken ¶
func (a *AccessResponse) GetAccessToken() string
func (*AccessResponse) GetExtra ¶
func (a *AccessResponse) GetExtra(key string) interface{}
func (*AccessResponse) GetTokenType ¶
func (a *AccessResponse) GetTokenType() string
func (*AccessResponse) SetAccessToken ¶
func (a *AccessResponse) SetAccessToken(token string)
func (*AccessResponse) SetExpiresIn ¶
func (a *AccessResponse) SetExpiresIn(expiresIn time.Duration)
func (*AccessResponse) SetExtra ¶
func (a *AccessResponse) SetExtra(key string, value interface{})
func (*AccessResponse) SetScopes ¶
func (a *AccessResponse) SetScopes(scopes Arguments)
func (*AccessResponse) SetTokenType ¶
func (a *AccessResponse) SetTokenType(name string)
func (*AccessResponse) ToMap ¶
func (a *AccessResponse) ToMap() map[string]interface{}
type AccessTokenIssuerProvider ¶
added in v0.43.0
type AccessTokenIssuerProvider interface {
	// GetAccessTokenIssuer returns the access token issuer.
	GetAccessTokenIssuer(ctx context.Context) string
}
AccessTokenIssuerProvider returns the provider for configuring the JWT issuer.

type AccessTokenLifespanProvider ¶
added in v0.43.0
type AccessTokenLifespanProvider interface {
	// GetAccessTokenLifespan returns the access token lifespan.
	GetAccessTokenLifespan(ctx context.Context) time.Duration
}
AccessTokenLifespanProvider returns the provider for configuring the access token lifespan.

type AllowedPromptValuesProvider ¶
added in v0.43.0
type AllowedPromptValuesProvider interface {
	// GetAllowedPromptValues returns the allowed prompt values.
	GetAllowedPromptValues(ctx context.Context) int
}
AllowedPromptValuesProvider returns the provider for configuring the allowed prompt values.

type AllowedPromptsProvider ¶
added in v0.43.0
type AllowedPromptsProvider interface {
	// GetAllowedPrompts returns the allowed prompts.
	GetAllowedPrompts(ctx context.Context) []string
}
AllowedPromptsProvider returns the provider for configuring the allowed prompts.

type Arguments ¶
type Arguments []string
func (Arguments)
Exact
DEPRECATED
func (Arguments) ExactOne ¶
added in v0.30.3
func (r Arguments) ExactOne(name string) bool
ExactOne checks, by string case, that a single argument equals the provided string.

func (Arguments) Has ¶
func (r Arguments) Has(items ...string) bool
Has checks, in a case-insensitive manner, that all of the items provided exists in arguments.

func (Arguments) HasOneOf ¶
added in v0.15.6
func (r Arguments) HasOneOf(items ...string) bool
HasOneOf checks, in a case-insensitive manner, that one of the items provided exists in arguments.

func (Arguments) Matches ¶
func (r Arguments) Matches(items ...string) bool
Matches performs an case-insensitive, out-of-order check that the items provided exist and equal all of the args in arguments. Note:

Providing a list that includes duplicate string-case items will return not matched.
func (Arguments) MatchesExact ¶
added in v0.30.3
func (r Arguments) MatchesExact(items ...string) bool
MatchesExact checks, by order and string case, that the items provided equal those in arguments.

type AudienceMatchingStrategy ¶
added in v0.27.0
type AudienceMatchingStrategy func(haystack []string, needle []string) error
type AudienceStrategyProvider ¶
added in v0.43.0
type AudienceStrategyProvider interface {
	// GetAudienceStrategy returns the audience strategy.
	GetAudienceStrategy(ctx context.Context) AudienceMatchingStrategy
}
AudienceStrategyProvider returns the provider for configuring the audience strategy.

type AuthorizeCodeLifespanProvider ¶
added in v0.43.0
type AuthorizeCodeLifespanProvider interface {
	// GetAuthorizeCodeLifespan returns the authorization code lifespan.
	GetAuthorizeCodeLifespan(ctx context.Context) time.Duration
}
AuthorizeCodeLifespanProvider returns the provider for configuring the authorization code lifespan.

type AuthorizeEndpointHandler ¶
type AuthorizeEndpointHandler interface {
	// HandleAuthorizeRequest handles an authorize endpoint request. To extend the handler's capabilities, the http request
	// is passed along, if further information retrieval is required. If the handler feels that he is not responsible for
	// the authorize request, he must return nil and NOT modify session nor responder neither requester.
	//
	// The following spec is a good example of what HandleAuthorizeRequest should do.
	// * https://tools.ietf.org/html/rfc6749#section-3.1.1
	//   response_type REQUIRED.
	//   The value MUST be one of "code" for requesting an
	//   authorization code as described by Section 4.1.1, "token" for
	//   requesting an access token (implicit grant) as described by
	//   Section 4.2.1, or a registered extension value as described by Section 8.4.
	HandleAuthorizeEndpointRequest(ctx context.Context, requester AuthorizeRequester, responder AuthorizeResponder) error
}
type AuthorizeEndpointHandlers ¶
type AuthorizeEndpointHandlers []AuthorizeEndpointHandler
AuthorizeEndpointHandlers is a list of AuthorizeEndpointHandler

func (*AuthorizeEndpointHandlers) Append ¶
func (a *AuthorizeEndpointHandlers) Append(h AuthorizeEndpointHandler)
Append adds an AuthorizeEndpointHandler to this list. Ignores duplicates based on reflect.TypeOf.

type AuthorizeEndpointHandlersProvider ¶
added in v0.43.0
type AuthorizeEndpointHandlersProvider interface {
	// GetAuthorizeEndpointHandlers returns the authorize endpoint handlers.
	GetAuthorizeEndpointHandlers(ctx context.Context) AuthorizeEndpointHandlers
}
AuthorizeEndpointHandlersProvider returns the provider for configuring the authorize endpoint handlers.

type AuthorizeRequest ¶
type AuthorizeRequest struct {
	ResponseTypes        Arguments        `json:"responseTypes" gorethink:"responseTypes"`
	RedirectURI          *url.URL         `json:"redirectUri" gorethink:"redirectUri"`
	State                string           `json:"state" gorethink:"state"`
	HandledResponseTypes Arguments        `json:"handledResponseTypes" gorethink:"handledResponseTypes"`
	ResponseMode         ResponseModeType `json:"ResponseModes" gorethink:"ResponseModes"`
	DefaultResponseMode  ResponseModeType `json:"DefaultResponseMode" gorethink:"DefaultResponseMode"`

	Request
}
AuthorizeRequest is an implementation of AuthorizeRequester

func NewAuthorizeRequest ¶
func NewAuthorizeRequest() *AuthorizeRequest
func (*AuthorizeRequest) DidHandleAllResponseTypes ¶
func (d *AuthorizeRequest) DidHandleAllResponseTypes() bool
func (*AuthorizeRequest) GetDefaultResponseMode ¶
added in v0.36.0
func (d *AuthorizeRequest) GetDefaultResponseMode() ResponseModeType
func (*AuthorizeRequest) GetRedirectURI ¶
func (d *AuthorizeRequest) GetRedirectURI() *url.URL
func (*AuthorizeRequest) GetResponseMode ¶
added in v0.36.0
func (d *AuthorizeRequest) GetResponseMode() ResponseModeType
func (*AuthorizeRequest) GetResponseTypes ¶
func (d *AuthorizeRequest) GetResponseTypes() Arguments
func (*AuthorizeRequest) GetState ¶
func (d *AuthorizeRequest) GetState() string
func (*AuthorizeRequest) IsRedirectURIValid ¶
func (d *AuthorizeRequest) IsRedirectURIValid() bool
func (*AuthorizeRequest) SetDefaultResponseMode ¶
added in v0.36.0
func (d *AuthorizeRequest) SetDefaultResponseMode(defaultResponseMode ResponseModeType)
func (*AuthorizeRequest) SetResponseTypeHandled ¶
func (d *AuthorizeRequest) SetResponseTypeHandled(name string)
type AuthorizeRequester ¶
type AuthorizeRequester interface {
	// GetResponseTypes returns the requested response types
	GetResponseTypes() (responseTypes Arguments)

	// SetResponseTypeHandled marks a response_type (e.g. token or code) as handled indicating that the response type
	// is supported.
	SetResponseTypeHandled(responseType string)

	// DidHandleAllResponseTypes returns if all requested response types have been handled correctly
	DidHandleAllResponseTypes() (didHandle bool)

	// GetRedirectURI returns the requested redirect URI
	GetRedirectURI() (redirectURL *url.URL)

	// IsRedirectURIValid returns false if the redirect is not rfc-conform (i.e. missing client, not on white list,
	// or malformed)
	IsRedirectURIValid() (isValid bool)

	// GetState returns the request's state.
	GetState() (state string)

	// GetResponseMode returns response_mode of the authorization request
	GetResponseMode() ResponseModeType

	// SetDefaultResponseMode sets default response mode for a response type in a flow
	SetDefaultResponseMode(responseMode ResponseModeType)

	// GetDefaultResponseMode gets default response mode for a response type in a flow
	GetDefaultResponseMode() ResponseModeType

	Requester
}
AuthorizeRequester is an authorize endpoint's request context.

type AuthorizeResponder ¶
type AuthorizeResponder interface {
	// GetCode returns the response's authorize code if set.
	GetCode() string

	// GetHeader returns the response's header
	GetHeader() (header http.Header)

	// AddHeader adds an header key value pair to the response
	AddHeader(key, value string)

	// GetParameters returns the response's parameters
	GetParameters() (query url.Values)

	// AddParameter adds key value pair to the response
	AddParameter(key, value string)
}
AuthorizeResponder is an authorization endpoint's response.

type AuthorizeResponse ¶
type AuthorizeResponse struct {
	Header     http.Header
	Parameters url.Values
	// contains filtered or unexported fields
}
AuthorizeResponse is an implementation of AuthorizeResponder

func NewAuthorizeResponse ¶
func NewAuthorizeResponse() *AuthorizeResponse
func (*AuthorizeResponse) AddHeader ¶
func (a *AuthorizeResponse) AddHeader(key, value string)
func (*AuthorizeResponse) AddParameter ¶
added in v0.36.0
func (a *AuthorizeResponse) AddParameter(key, value string)
func (*AuthorizeResponse) GetCode ¶
func (a *AuthorizeResponse) GetCode() string
func (*AuthorizeResponse) GetHeader ¶
func (a *AuthorizeResponse) GetHeader() http.Header
func (*AuthorizeResponse) GetParameters ¶
added in v0.36.0
func (a *AuthorizeResponse) GetParameters() url.Values
type BCrypt ¶
added in v0.4.0
type BCrypt struct {
	Config interface {
		BCryptCostProvider
	}
}
BCrypt implements the Hasher interface by using BCrypt.

func (*BCrypt) Compare ¶
added in v0.4.0
func (b *BCrypt) Compare(ctx context.Context, hash, data []byte) error
func (*BCrypt) Hash ¶
added in v0.4.0
func (b *BCrypt) Hash(ctx context.Context, data []byte) ([]byte, error)
type BCryptCostProvider ¶
added in v0.43.0
type BCryptCostProvider interface {
	// GetBCryptCost returns the BCrypt  hash cost.
	GetBCryptCost(ctx context.Context) int
}
BCryptCostProvider returns the provider for configuring the BCrypt hash cost.

type Client ¶
type Client interface {
	// GetID returns the client ID.
	GetID() string

	// GetHashedSecret returns the hashed secret as it is stored in the store.
	GetHashedSecret() []byte

	// GetRedirectURIs returns the client's allowed redirect URIs.
	GetRedirectURIs() []string

	// GetGrantTypes returns the client's allowed grant types.
	GetGrantTypes() Arguments

	// GetResponseTypes returns the client's allowed response types.
	// All allowed combinations of response types have to be listed, each combination having
	// response types of the combination separated by a space.
	GetResponseTypes() Arguments

	// GetScopes returns the scopes this client is allowed to request.
	GetScopes() Arguments

	// IsPublic returns true, if this client is marked as public.
	IsPublic() bool

	// GetAudience returns the allowed audience(s) for this client.
	GetAudience() Arguments
}
Client represents a client or an app.

type ClientAuthenticationStrategy ¶
added in v0.38.0
type ClientAuthenticationStrategy func(context.Context, *http.Request, url.Values) (Client, error)
ClientAuthenticationStrategy provides a method signature for authenticating a client request

type ClientAuthenticationStrategyProvider ¶
added in v0.43.0
type ClientAuthenticationStrategyProvider interface {
	// GetClientAuthenticationStrategy returns the client authentication strategy.
	GetClientAuthenticationStrategy(ctx context.Context) ClientAuthenticationStrategy
}
ClientAuthenticationStrategyProvider returns the provider for configuring the client authentication strategy.

type ClientLifespanConfig ¶
added in v0.43.0
type ClientLifespanConfig struct {
	AuthorizationCodeGrantAccessTokenLifespan  *time.Duration `json:"authorization_code_grant_access_token_lifespan"`
	AuthorizationCodeGrantIDTokenLifespan      *time.Duration `json:"authorization_code_grant_id_token_lifespan"`
	AuthorizationCodeGrantRefreshTokenLifespan *time.Duration `json:"authorization_code_grant_refresh_token_lifespan"`
	ClientCredentialsGrantAccessTokenLifespan  *time.Duration `json:"client_credentials_grant_access_token_lifespan"`
	ImplicitGrantAccessTokenLifespan           *time.Duration `json:"implicit_grant_access_token_lifespan"`
	ImplicitGrantIDTokenLifespan               *time.Duration `json:"implicit_grant_id_token_lifespan"`
	JwtBearerGrantAccessTokenLifespan          *time.Duration `json:"jwt_bearer_grant_access_token_lifespan"`
	PasswordGrantAccessTokenLifespan           *time.Duration `json:"password_grant_access_token_lifespan"`
	PasswordGrantRefreshTokenLifespan          *time.Duration `json:"password_grant_refresh_token_lifespan"`
	RefreshTokenGrantIDTokenLifespan           *time.Duration `json:"refresh_token_grant_id_token_lifespan"`
	RefreshTokenGrantAccessTokenLifespan       *time.Duration `json:"refresh_token_grant_access_token_lifespan"`
	RefreshTokenGrantRefreshTokenLifespan      *time.Duration `json:"refresh_token_grant_refresh_token_lifespan"`
}
ClientLifespanConfig holds default lifespan configuration for the different token types that may be issued for the client. This configuration takes precedence over fosite's instance-wide default lifespan, but it may be overridden by a session's expires_at claim.

The OIDC Hybrid grant type inherits token lifespan configuration from the implicit grant.

type ClientManager ¶
type ClientManager interface {
	// GetClient loads the client by its ID or returns an error
	// if the client does not exist or another error occurred.
	GetClient(ctx context.Context, id string) (Client, error)
	// ClientAssertionJWTValid returns an error if the JTI is
	// known or the DB check failed and nil if the JTI is not known.
	ClientAssertionJWTValid(ctx context.Context, jti string) error
	// SetClientAssertionJWT marks a JTI as known for the given
	// expiry time. Before inserting the new JTI, it will clean
	// up any existing JTIs that have expired as those tokens can
	// not be replayed due to the expiry.
	SetClientAssertionJWT(ctx context.Context, jti string, exp time.Time) error
}
ClientManager defines the (persistent) manager interface for clients.

type ClientWithCustomTokenLifespans ¶
added in v0.43.0
type ClientWithCustomTokenLifespans interface {
	// GetEffectiveLifespan either maps GrantType x TokenType to the client's configured lifespan, or returns the fallback value.
	GetEffectiveLifespan(gt GrantType, tt TokenType, fallback time.Duration) time.Duration
}
type ClientWithSecretRotation ¶
added in v0.41.0
type ClientWithSecretRotation interface {
	Client
	// GetRotatedHashes returns a slice of hashed secrets used for secrets rotation.
	GetRotatedHashes() [][]byte
}
ClientWithSecretRotation extends Client interface by a method providing a slice of rotated secrets.

type Config ¶
added in v0.43.0
type Config struct {
	// AccessTokenLifespan sets how long an access token is going to be valid. Defaults to one hour.
	AccessTokenLifespan time.Duration

	// VerifiableCredentialsNonceLifespan sets how long a verifiable credentials nonce is going to be valid. Defaults to one hour.
	VerifiableCredentialsNonceLifespan time.Duration

	// RefreshTokenLifespan sets how long a refresh token is going to be valid. Defaults to 30 days. Set to -1 for
	// refresh tokens that never expire.
	RefreshTokenLifespan time.Duration

	// AuthorizeCodeLifespan sets how long an authorize code is going to be valid. Defaults to fifteen minutes.
	AuthorizeCodeLifespan time.Duration

	// IDTokenLifespan sets the default id token lifetime. Defaults to one hour.
	IDTokenLifespan time.Duration

	// IDTokenIssuer sets the default issuer of the ID Token.
	IDTokenIssuer string

	// HashCost sets the cost of the password hashing cost. Defaults to 12.
	HashCost int

	// DisableRefreshTokenValidation sets the introspection endpoint to disable refresh token validation.
	DisableRefreshTokenValidation bool

	// SendDebugMessagesToClients if set to true, includes error debug messages in response payloads. Be aware that sensitive
	// data may be exposed, depending on your implementation of Fosite. Such sensitive data might include database error
	// codes or other information. Proceed with caution!
	SendDebugMessagesToClients bool

	// ScopeStrategy sets the scope strategy that should be supported, for example fosite.WildcardScopeStrategy.
	ScopeStrategy ScopeStrategy

	// AudienceMatchingStrategy sets the audience matching strategy that should be supported, defaults to fosite.DefaultsAudienceMatchingStrategy.
	AudienceMatchingStrategy AudienceMatchingStrategy

	// EnforcePKCE, if set to true, requires clients to perform authorize code flows with PKCE. Defaults to false.
	EnforcePKCE bool

	// EnforcePKCEForPublicClients requires only public clients to use PKCE with the authorize code flow. Defaults to false.
	EnforcePKCEForPublicClients bool

	// EnablePKCEPlainChallengeMethod sets whether or not to allow the plain challenge method (S256 should be used whenever possible, plain is really discouraged). Defaults to false.
	EnablePKCEPlainChallengeMethod bool

	// AllowedPromptValues sets which OpenID Connect prompt values the server supports. Defaults to []string{"login", "none", "consent", "select_account"}.
	AllowedPromptValues []string

	// TokenURL is the the URL of the Authorization Server's Token Endpoint. If the authorization server is intended
	// to be compatible with the private_key_jwt client authentication method (see http://openid.net/specs/openid-connect-core-1_0.html#CodeFlowAuth),
	// this value MUST be set.
	TokenURL string

	// JWKSFetcherStrategy is responsible for fetching JSON Web Keys from remote URLs. This is required when the private_key_jwt
	// client authentication method is used. Defaults to fosite.DefaultJWKSFetcherStrategy.
	JWKSFetcherStrategy JWKSFetcherStrategy

	// TokenEntropy indicates the entropy of the random string, used as the "message" part of the HMAC token.
	// Defaults to 32.
	TokenEntropy int

	// RedirectSecureChecker is a function that returns true if the provided URL can be securely used as a redirect URL.
	RedirectSecureChecker func(context.Context, *url.URL) bool

	// RefreshTokenScopes defines which OAuth scopes will be given refresh tokens during the authorization code grant exchange. This defaults to "offline" and "offline_access". When set to an empty array, all exchanges will be given refresh tokens.
	RefreshTokenScopes []string

	// MinParameterEntropy controls the minimum size of state and nonce parameters. Defaults to fosite.MinParameterEntropy.
	MinParameterEntropy int

	// UseLegacyErrorFormat controls whether the legacy error format (with `error_debug`, `error_hint`, ...)
	// should be used or not.
	UseLegacyErrorFormat bool

	// GrantTypeJWTBearerCanSkipClientAuth indicates, if client authentication can be skipped, when using jwt as assertion.
	GrantTypeJWTBearerCanSkipClientAuth bool

	// GrantTypeJWTBearerIDOptional indicates, if jti (JWT ID) claim required or not in JWT.
	GrantTypeJWTBearerIDOptional bool

	// GrantTypeJWTBearerIssuedDateOptional indicates, if "iat" (issued at) claim required or not in JWT.
	GrantTypeJWTBearerIssuedDateOptional bool

	// GrantTypeJWTBearerMaxDuration sets the maximum time after JWT issued date, during which the JWT is considered valid.
	GrantTypeJWTBearerMaxDuration time.Duration

	// ClientAuthenticationStrategy indicates the Strategy to authenticate client requests
	ClientAuthenticationStrategy ClientAuthenticationStrategy

	// ResponseModeHandlerExtension provides a handler for custom response modes
	ResponseModeHandlerExtension ResponseModeHandler

	// MessageCatalog is the message bundle used for i18n
	MessageCatalog i18n.MessageCatalog

	// FormPostHTMLTemplate sets html template for rendering the authorization response when the request has response_mode=form_post.
	FormPostHTMLTemplate *template.Template

	// OmitRedirectScopeParam indicates whether the "scope" parameter should be omitted from the redirect URL.
	OmitRedirectScopeParam bool

	// SanitationWhiteList is a whitelist of form values that are required by the token endpoint. These values
	// are safe for storage in a database (cleartext).
	SanitationWhiteList []string

	// JWTScopeClaimKey defines the claim key to be used to set the scope in. Valid fields are "scope" or "scp" or both.
	JWTScopeClaimKey jwt.JWTScopeFieldEnum

	// AccessTokenIssuer is the issuer to be used when generating access tokens.
	AccessTokenIssuer string

	// ClientSecretsHasher is the hasher used to hash OAuth2 Client Secrets.
	ClientSecretsHasher Hasher

	// HTTPClient is the HTTP client to use for requests.
	HTTPClient *retryablehttp.Client

	// AuthorizeEndpointHandlers is a list of handlers that are called before the authorization endpoint is served.
	AuthorizeEndpointHandlers AuthorizeEndpointHandlers

	// TokenEndpointHandlers is a list of handlers that are called before the token endpoint is served.
	TokenEndpointHandlers TokenEndpointHandlers

	// TokenIntrospectionHandlers is a list of handlers that are called before the token introspection endpoint is served.
	TokenIntrospectionHandlers TokenIntrospectionHandlers

	// RevocationHandlers is a list of handlers that are called before the revocation endpoint is served.
	RevocationHandlers RevocationHandlers

	// PushedAuthorizeEndpointHandlers is a list of handlers that are called before the PAR endpoint is served.
	PushedAuthorizeEndpointHandlers PushedAuthorizeEndpointHandlers

	// GlobalSecret is the global secret used to sign and verify signatures.
	GlobalSecret []byte

	// RotatedGlobalSecrets is a list of global secrets that are used to verify signatures.
	RotatedGlobalSecrets [][]byte

	// HMACHasher is the hasher used to generate HMAC signatures.
	HMACHasher func() hash.Hash

	// PushedAuthorizeRequestURIPrefix is the URI prefix for the PAR request_uri.
	// This is defaulted to 'urn:ietf:params:oauth:request_uri:'.
	PushedAuthorizeRequestURIPrefix string

	// PushedAuthorizeContextLifespan is the lifespan of the PAR context
	PushedAuthorizeContextLifespan time.Duration

	// IsPushedAuthorizeEnforced enforces pushed authorization request for /authorize
	IsPushedAuthorizeEnforced bool
}
func (*Config) EnforcePushedAuthorize ¶
added in v0.43.0
func (c *Config) EnforcePushedAuthorize(ctx context.Context) bool
EnforcePushedAuthorize indicates if PAR is enforced. In this mode, a client cannot pass authorize parameters at the 'authorize' endpoint. The 'authorize' endpoint must contain the PAR request_uri.

func (*Config) GetAccessTokenIssuer ¶
added in v0.43.0
func (c *Config) GetAccessTokenIssuer(ctx context.Context) string
func (*Config) GetAccessTokenLifespan ¶
added in v0.43.0
func (c *Config) GetAccessTokenLifespan(_ context.Context) time.Duration
GetAccessTokenLifespan returns how long an access token should be valid. Defaults to one hour.

func (*Config) GetAllowedPrompts ¶
added in v0.43.0
func (c *Config) GetAllowedPrompts(_ context.Context) []string
func (*Config) GetAudienceStrategy ¶
added in v0.43.0
func (c *Config) GetAudienceStrategy(_ context.Context) AudienceMatchingStrategy
GetAudienceStrategy returns the scope strategy to be used. Defaults to glob scope strategy.

func (*Config) GetAuthorizeCodeLifespan ¶
added in v0.43.0
func (c *Config) GetAuthorizeCodeLifespan(_ context.Context) time.Duration
GetAuthorizeCodeLifespan returns how long an authorize code should be valid. Defaults to one fifteen minutes.

func (*Config) GetAuthorizeEndpointHandlers ¶
added in v0.43.0
func (c *Config) GetAuthorizeEndpointHandlers(ctx context.Context) AuthorizeEndpointHandlers
func (*Config) GetBCryptCost ¶
added in v0.43.0
func (c *Config) GetBCryptCost(_ context.Context) int
GetBCryptCost returns the bcrypt cost factor. Defaults to 12.

func (*Config) GetClientAuthenticationStrategy ¶
added in v0.43.0
func (c *Config) GetClientAuthenticationStrategy(_ context.Context) ClientAuthenticationStrategy
GetClientAuthenticationStrategy returns the configured client authentication strategy. Defaults to nil. Note that on a nil strategy `fosite.Fosite` fallbacks to its default client authentication strategy `fosite.Fosite.DefaultClientAuthenticationStrategy`

func (*Config) GetDisableRefreshTokenValidation ¶
added in v0.43.0
func (c *Config) GetDisableRefreshTokenValidation(_ context.Context) bool
GetDisableRefreshTokenValidation returns whether to disable the validation of the refresh token.

func (*Config) GetEnablePKCEPlainChallengeMethod ¶
added in v0.43.0
func (c *Config) GetEnablePKCEPlainChallengeMethod(ctx context.Context) bool
GetEnablePKCEPlainChallengeMethod returns whether or not to allow the plain challenge method (S256 should be used whenever possible, plain is really discouraged).

func (*Config) GetEnforcePKCE ¶
added in v0.43.0
func (c *Config) GetEnforcePKCE(ctx context.Context) bool
GetEnforcePKCE If set to true, public clients must use PKCE.

func (*Config) GetEnforcePKCEForPublicClients ¶
added in v0.43.0
func (c *Config) GetEnforcePKCEForPublicClients(ctx context.Context) bool
GetEnforcePKCEForPublicClients returns the value of EnforcePKCEForPublicClients.

func (*Config) GetFormPostHTMLTemplate ¶
added in v0.43.0
func (c *Config) GetFormPostHTMLTemplate(ctx context.Context) *template.Template
func (*Config) GetGlobalSecret ¶
added in v0.43.0
func (c *Config) GetGlobalSecret(ctx context.Context) ([]byte, error)
func (*Config) GetGrantTypeJWTBearerCanSkipClientAuth ¶
added in v0.43.0
func (c *Config) GetGrantTypeJWTBearerCanSkipClientAuth(ctx context.Context) bool
GetGrantTypeJWTBearerCanSkipClientAuth returns the GrantTypeJWTBearerCanSkipClientAuth field.

func (*Config) GetGrantTypeJWTBearerIDOptional ¶
added in v0.43.0
func (c *Config) GetGrantTypeJWTBearerIDOptional(ctx context.Context) bool
GetGrantTypeJWTBearerIDOptional returns the GrantTypeJWTBearerIDOptional field.

func (*Config) GetGrantTypeJWTBearerIssuedDateOptional ¶
added in v0.43.0
func (c *Config) GetGrantTypeJWTBearerIssuedDateOptional(ctx context.Context) bool
GetGrantTypeJWTBearerIssuedDateOptional returns the GrantTypeJWTBearerIssuedDateOptional field.

func (*Config) GetHMACHasher ¶
added in v0.43.0
func (c *Config) GetHMACHasher(ctx context.Context) func() hash.Hash
func (*Config) GetHTTPClient ¶
added in v0.43.0
func (c *Config) GetHTTPClient(ctx context.Context) *retryablehttp.Client
func (*Config) GetIDTokenIssuer ¶
added in v0.43.0
func (c *Config) GetIDTokenIssuer(ctx context.Context) string
func (*Config) GetIDTokenLifespan ¶
added in v0.43.0
func (c *Config) GetIDTokenLifespan(_ context.Context) time.Duration
GetIDTokenLifespan returns how long an id token should be valid. Defaults to one hour.

func (*Config) GetJWKSFetcherStrategy ¶
added in v0.43.0
func (c *Config) GetJWKSFetcherStrategy(_ context.Context) JWKSFetcherStrategy
GetJWKSFetcherStrategy returns the JWKSFetcherStrategy.

func (*Config) GetJWTMaxDuration ¶
added in v0.43.0
func (c *Config) GetJWTMaxDuration(_ context.Context) time.Duration
GetJWTMaxDuration specified the maximum amount of allowed `exp` time for a JWT. It compares the time with the JWT's `exp` time if the JWT time is larger, will cause the JWT to be invalid.

Defaults to a day.

func (*Config) GetJWTScopeField ¶
added in v0.43.0
func (c *Config) GetJWTScopeField(ctx context.Context) jwt.JWTScopeFieldEnum
func (*Config) GetMessageCatalog ¶
added in v0.43.0
func (c *Config) GetMessageCatalog(ctx context.Context) i18n.MessageCatalog
func (*Config) GetMinParameterEntropy ¶
added in v0.43.0
func (c *Config) GetMinParameterEntropy(_ context.Context) int
GetMinParameterEntropy returns MinParameterEntropy if set. Defaults to fosite.MinParameterEntropy.

func (*Config) GetOmitRedirectScopeParam ¶
added in v0.43.0
func (c *Config) GetOmitRedirectScopeParam(ctx context.Context) bool
func (*Config) GetPushedAuthorizeContextLifespan ¶
added in v0.43.0
func (c *Config) GetPushedAuthorizeContextLifespan(ctx context.Context) time.Duration
GetPushedAuthorizeContextLifespan is the lifespan of the short-lived PAR context.

func (*Config) GetPushedAuthorizeEndpointHandlers ¶
added in v0.43.0
func (c *Config) GetPushedAuthorizeEndpointHandlers(ctx context.Context) PushedAuthorizeEndpointHandlers
GetPushedAuthorizeEndpointHandlers returns the handlers.

func (*Config) GetPushedAuthorizeRequestURIPrefix ¶
added in v0.43.0
func (c *Config) GetPushedAuthorizeRequestURIPrefix(ctx context.Context) string
GetPushedAuthorizeRequestURIPrefix is the request URI prefix. This is usually 'urn:ietf:params:oauth:request_uri:'.

func (*Config) GetRedirectSecureChecker ¶
added in v0.43.0
func (c *Config) GetRedirectSecureChecker(_ context.Context) func(context.Context, *url.URL) bool
GetRedirectSecureChecker returns the checker to check if redirect URI is secure. Defaults to fosite.IsRedirectURISecure.

func (*Config) GetRefreshTokenLifespan ¶
added in v0.43.0
func (c *Config) GetRefreshTokenLifespan(_ context.Context) time.Duration
GetRefreshTokenLifespan sets how long a refresh token is going to be valid. Defaults to 30 days. Set to -1 for refresh tokens that never expire.

func (*Config) GetRefreshTokenScopes ¶
added in v0.43.0
func (c *Config) GetRefreshTokenScopes(_ context.Context) []string
GetRefreshTokenScopes returns which scopes will provide refresh tokens.

func (*Config) GetResponseModeHandlerExtension ¶
added in v0.43.0
func (c *Config) GetResponseModeHandlerExtension(ctx context.Context) ResponseModeHandler
func (*Config) GetRevocationHandlers ¶
added in v0.43.0
func (c *Config) GetRevocationHandlers(ctx context.Context) RevocationHandlers
func (*Config) GetRotatedGlobalSecrets ¶
added in v0.43.0
func (c *Config) GetRotatedGlobalSecrets(ctx context.Context) ([][]byte, error)
func (*Config) GetSanitationWhiteList ¶
added in v0.43.0
func (c *Config) GetSanitationWhiteList(ctx context.Context) []string
GetSanitationWhiteList returns a list of allowed form values that are required by the token endpoint. These values are safe for storage in a database (cleartext).

func (*Config) GetScopeStrategy ¶
added in v0.43.0
func (c *Config) GetScopeStrategy(_ context.Context) ScopeStrategy
GetScopeStrategy returns the scope strategy to be used. Defaults to glob scope strategy.

func (*Config) GetSecretsHasher ¶
added in v0.43.0
func (c *Config) GetSecretsHasher(ctx context.Context) Hasher
func (*Config) GetSendDebugMessagesToClients ¶
added in v0.43.0
func (c *Config) GetSendDebugMessagesToClients(ctx context.Context) bool
func (*Config) GetTokenEndpointHandlers ¶
added in v0.43.0
func (c *Config) GetTokenEndpointHandlers(ctx context.Context) TokenEndpointHandlers
func (*Config) GetTokenEntropy ¶
added in v0.43.0
func (c *Config) GetTokenEntropy(_ context.Context) int
GetTokenEntropy returns the entropy of the "message" part of a HMAC Token. Defaults to 32.

func (*Config) GetTokenIntrospectionHandlers ¶
added in v0.43.0
func (c *Config) GetTokenIntrospectionHandlers(ctx context.Context) TokenIntrospectionHandlers
func (*Config) GetTokenURLs ¶
added in v0.45.0
func (c *Config) GetTokenURLs(ctx context.Context) []string
func (*Config) GetUseLegacyErrorFormat ¶
added in v0.43.0
func (c *Config) GetUseLegacyErrorFormat(ctx context.Context) bool
func (*Config) GetVerifiableCredentialsNonceLifespan ¶
added in v0.45.0
func (c *Config) GetVerifiableCredentialsNonceLifespan(_ context.Context) time.Duration
GetNonceLifespan returns how long a nonce should be valid. Defaults to one hour.

type Configurator ¶
added in v0.43.0
type Configurator interface {
	IDTokenIssuerProvider
	IDTokenLifespanProvider
	AllowedPromptsProvider
	EnforcePKCEProvider
	EnforcePKCEForPublicClientsProvider
	EnablePKCEPlainChallengeMethodProvider
	GrantTypeJWTBearerCanSkipClientAuthProvider
	GrantTypeJWTBearerIDOptionalProvider
	GrantTypeJWTBearerIssuedDateOptionalProvider
	GetJWTMaxDurationProvider
	AudienceStrategyProvider
	ScopeStrategyProvider
	RedirectSecureCheckerProvider
	OmitRedirectScopeParamProvider
	SanitationAllowedProvider
	JWTScopeFieldProvider
	AccessTokenIssuerProvider
	DisableRefreshTokenValidationProvider
	RefreshTokenScopesProvider
	AccessTokenLifespanProvider
	RefreshTokenLifespanProvider
	VerifiableCredentialsNonceLifespanProvider
	AuthorizeCodeLifespanProvider
	TokenEntropyProvider
	RotatedGlobalSecretsProvider
	GlobalSecretProvider
	JWKSFetcherStrategyProvider
	HTTPClientProvider
	ScopeStrategyProvider
	AudienceStrategyProvider
	MinParameterEntropyProvider
	HMACHashingProvider
	ClientAuthenticationStrategyProvider
	ResponseModeHandlerExtensionProvider
	SendDebugMessagesToClientsProvider
	JWKSFetcherStrategyProvider
	ClientAuthenticationStrategyProvider
	ResponseModeHandlerExtensionProvider
	MessageCatalogProvider
	FormPostHTMLTemplateProvider
	TokenURLProvider
	GetSecretsHashingProvider
	AuthorizeEndpointHandlersProvider
	TokenEndpointHandlersProvider
	TokenIntrospectionHandlersProvider
	RevocationHandlersProvider
	UseLegacyErrorFormatProvider
}
type ContextKey ¶
added in v0.40.0
type ContextKey string
type DefaultClient ¶
type DefaultClient struct {
	ID             string   `json:"id"`
	Secret         []byte   `json:"client_secret,omitempty"`
	RotatedSecrets [][]byte `json:"rotated_secrets,omitempty"`
	RedirectURIs   []string `json:"redirect_uris"`
	GrantTypes     []string `json:"grant_types"`
	ResponseTypes  []string `json:"response_types"`
	Scopes         []string `json:"scopes"`
	Audience       []string `json:"audience"`
	Public         bool     `json:"public"`
}
DefaultClient is a simple default implementation of the Client interface.

func (*DefaultClient) GetAudience ¶
added in v0.27.0
func (c *DefaultClient) GetAudience() Arguments
func (*DefaultClient) GetGrantTypes ¶
func (c *DefaultClient) GetGrantTypes() Arguments
func (*DefaultClient) GetHashedSecret ¶
func (c *DefaultClient) GetHashedSecret() []byte
func (*DefaultClient) GetID ¶
func (c *DefaultClient) GetID() string
func (*DefaultClient) GetRedirectURIs ¶
func (c *DefaultClient) GetRedirectURIs() []string
func (*DefaultClient) GetResponseTypes ¶
func (c *DefaultClient) GetResponseTypes() Arguments
func (*DefaultClient) GetRotatedHashes ¶
added in v0.41.0
func (c *DefaultClient) GetRotatedHashes() [][]byte
func (*DefaultClient) GetScopes ¶
added in v0.2.0
func (c *DefaultClient) GetScopes() Arguments
func (*DefaultClient) IsPublic ¶
added in v0.4.0
func (c *DefaultClient) IsPublic() bool
type DefaultClientWithCustomTokenLifespans ¶
added in v0.43.0
type DefaultClientWithCustomTokenLifespans struct {
	*DefaultClient
	TokenLifespans *ClientLifespanConfig `json:"token_lifespans"`
}
func (*DefaultClientWithCustomTokenLifespans) GetEffectiveLifespan ¶
added in v0.43.0
func (c *DefaultClientWithCustomTokenLifespans) GetEffectiveLifespan(gt GrantType, tt TokenType, fallback time.Duration) time.Duration
GetEffectiveLifespan either maps GrantType x TokenType to the client's configured lifespan, or returns the fallback value.

func (*DefaultClientWithCustomTokenLifespans) GetTokenLifespans ¶
added in v0.43.0
func (c *DefaultClientWithCustomTokenLifespans) GetTokenLifespans() *ClientLifespanConfig
func (*DefaultClientWithCustomTokenLifespans) SetTokenLifespans ¶
added in v0.43.0
func (c *DefaultClientWithCustomTokenLifespans) SetTokenLifespans(lifespans *ClientLifespanConfig)
type DefaultJWKSFetcherStrategy ¶
added in v0.21.0
type DefaultJWKSFetcherStrategy struct {
	// contains filtered or unexported fields
}
DefaultJWKSFetcherStrategy is a default implementation of the JWKSFetcherStrategy interface.

func (*DefaultJWKSFetcherStrategy) Resolve ¶
added in v0.21.0
func (s *DefaultJWKSFetcherStrategy) Resolve(ctx context.Context, location string, ignoreCache bool) (*jose.JSONWebKeySet, error)
Resolve returns the JSON Web Key Set, or an error if something went wrong. The forceRefresh, if true, forces the strategy to fetch the key from the remote. If forceRefresh is false, the strategy may use a caching strategy to fetch the key.

func (*DefaultJWKSFetcherStrategy) WaitForCache ¶
added in v0.43.0
func (s *DefaultJWKSFetcherStrategy) WaitForCache()
type DefaultOpenIDConnectClient ¶
added in v0.21.0
type DefaultOpenIDConnectClient struct {
	*DefaultClient
	JSONWebKeysURI                    string              `json:"jwks_uri"`
	JSONWebKeys                       *jose.JSONWebKeySet `json:"jwks"`
	TokenEndpointAuthMethod           string              `json:"token_endpoint_auth_method"`
	RequestURIs                       []string            `json:"request_uris"`
	RequestObjectSigningAlgorithm     string              `json:"request_object_signing_alg"`
	TokenEndpointAuthSigningAlgorithm string              `json:"token_endpoint_auth_signing_alg"`
}
func (*DefaultOpenIDConnectClient) GetJSONWebKeys ¶
added in v0.21.0
func (c *DefaultOpenIDConnectClient) GetJSONWebKeys() *jose.JSONWebKeySet
func (*DefaultOpenIDConnectClient) GetJSONWebKeysURI ¶
added in v0.21.0
func (c *DefaultOpenIDConnectClient) GetJSONWebKeysURI() string
func (*DefaultOpenIDConnectClient) GetRequestObjectSigningAlgorithm ¶
added in v0.21.0
func (c *DefaultOpenIDConnectClient) GetRequestObjectSigningAlgorithm() string
func (*DefaultOpenIDConnectClient) GetRequestURIs ¶
added in v0.21.0
func (c *DefaultOpenIDConnectClient) GetRequestURIs() []string
func (*DefaultOpenIDConnectClient) GetTokenEndpointAuthMethod ¶
added in v0.21.0
func (c *DefaultOpenIDConnectClient) GetTokenEndpointAuthMethod() string
func (*DefaultOpenIDConnectClient) GetTokenEndpointAuthSigningAlgorithm ¶
added in v0.21.0
func (c *DefaultOpenIDConnectClient) GetTokenEndpointAuthSigningAlgorithm() string
type DefaultResponseModeClient ¶
added in v0.36.0
type DefaultResponseModeClient struct {
	*DefaultClient
	ResponseModes []ResponseModeType `json:"response_modes"`
}
func (*DefaultResponseModeClient) GetResponseModes ¶
added in v0.36.0
func (c *DefaultResponseModeClient) GetResponseModes() []ResponseModeType
type DefaultResponseModeHandler ¶
added in v0.41.0
type DefaultResponseModeHandler struct{}
func NewDefaultResponseModeHandler ¶
added in v0.43.0
func NewDefaultResponseModeHandler() *DefaultResponseModeHandler
func (*DefaultResponseModeHandler) ResponseModes ¶
added in v0.41.0
func (d *DefaultResponseModeHandler) ResponseModes() ResponseModeTypes
func (*DefaultResponseModeHandler) WriteAuthorizeError ¶
added in v0.41.0
func (d *DefaultResponseModeHandler) WriteAuthorizeError(ctx context.Context, rw http.ResponseWriter, ar AuthorizeRequester, err error)
func (*DefaultResponseModeHandler) WriteAuthorizeResponse ¶
added in v0.41.0
func (d *DefaultResponseModeHandler) WriteAuthorizeResponse(ctx context.Context, rw http.ResponseWriter, ar AuthorizeRequester, resp AuthorizeResponder)
type DefaultSession ¶
added in v0.5.0
type DefaultSession struct {
	ExpiresAt map[TokenType]time.Time `json:"expires_at"`
	Username  string                  `json:"username"`
	Subject   string                  `json:"subject"`
	Extra     map[string]interface{}  `json:"extra"`
}
DefaultSession is a default implementation of the session interface.

func (*DefaultSession) Clone ¶
added in v0.6.0
func (s *DefaultSession) Clone() Session
func (*DefaultSession) GetExpiresAt ¶
added in v0.5.0
func (s *DefaultSession) GetExpiresAt(key TokenType) time.Time
func (*DefaultSession) GetExtraClaims ¶
added in v0.40.0
func (s *DefaultSession) GetExtraClaims() map[string]interface{}
GetExtraClaims implements ExtraClaimsSession for DefaultSession. The returned value can be modified in-place.

func (*DefaultSession) GetSubject ¶
added in v0.5.0
func (s *DefaultSession) GetSubject() string
func (*DefaultSession) GetUsername ¶
added in v0.5.0
func (s *DefaultSession) GetUsername() string
func (*DefaultSession) SetExpiresAt ¶
added in v0.5.0
func (s *DefaultSession) SetExpiresAt(key TokenType, exp time.Time)
func (*DefaultSession) SetSubject ¶
added in v0.37.0
func (s *DefaultSession) SetSubject(subject string)
type DisableRefreshTokenValidationProvider ¶
added in v0.43.0
type DisableRefreshTokenValidationProvider interface {
	// GetDisableRefreshTokenValidation returns the disable refresh token validation flag.
	GetDisableRefreshTokenValidation(ctx context.Context) bool
}
DisableRefreshTokenValidationProvider returns the provider for configuring the refresh token validation.

type EnablePKCEPlainChallengeMethodProvider ¶
added in v0.43.0
type EnablePKCEPlainChallengeMethodProvider interface {
	// GetEnablePKCEPlainChallengeMethod returns the enable PKCE plain challenge method.
	GetEnablePKCEPlainChallengeMethod(ctx context.Context) bool
}
EnablePKCEPlainChallengeMethodProvider returns the provider for configuring the enable PKCE plain challenge method.

type EnforcePKCEForPublicClientsProvider ¶
added in v0.43.0
type EnforcePKCEForPublicClientsProvider interface {
	// GetEnforcePKCEForPublicClients returns the enforcement of PKCE for public clients.
	GetEnforcePKCEForPublicClients(ctx context.Context) bool
}
EnforcePKCEForPublicClientsProvider returns the provider for configuring the enforcement of PKCE for public clients.

type EnforcePKCEProvider ¶
added in v0.43.0
type EnforcePKCEProvider interface {
	// GetEnforcePKCE returns the enforcement of PKCE.
	GetEnforcePKCE(ctx context.Context) bool
}
EnforcePKCEProvider returns the provider for configuring the enforcement of PKCE.

type ExtraClaimsSession ¶
added in v0.40.0
type ExtraClaimsSession interface {
	// GetExtraClaims returns a map to store extra claims.
	// The returned value can be modified in-place.
	GetExtraClaims() map[string]interface{}
}
ExtraClaimsSession provides an interface for session to store any extra claims.

type FormPostHTMLTemplateProvider ¶
added in v0.43.0
type FormPostHTMLTemplateProvider interface {
	// GetFormPostHTMLTemplate returns the form post HTML template.
	GetFormPostHTMLTemplate(ctx context.Context) *template.Template
}
FormPostHTMLTemplateProvider returns the provider for configuring the form post HTML template.

type Fosite ¶
type Fosite struct {
	Store Storage

	Config Configurator
}
Fosite implements OAuth2Provider.

func NewOAuth2Provider ¶
added in v0.43.0
func NewOAuth2Provider(s Storage, c Configurator) *Fosite
func (*Fosite) AuthenticateClient ¶
added in v0.21.0
func (f *Fosite) AuthenticateClient(ctx context.Context, r *http.Request, form url.Values) (Client, error)
AuthenticateClient authenticates client requests using the configured strategy `Fosite.ClientAuthenticationStrategy`, if nil it uses `Fosite.DefaultClientAuthenticationStrategy`

func (*Fosite) DefaultClientAuthenticationStrategy ¶
added in v0.38.0
func (f *Fosite) DefaultClientAuthenticationStrategy(ctx context.Context, r *http.Request, form url.Values) (Client, error)
DefaultClientAuthenticationStrategy provides the fosite's default client authentication strategy, HTTP Basic Authentication and JWT Bearer

func (*Fosite) GetMinParameterEntropy ¶
added in v0.32.4
func (f *Fosite) GetMinParameterEntropy(ctx context.Context) int
GetMinParameterEntropy returns MinParameterEntropy if set. Defaults to fosite.MinParameterEntropy.

func (*Fosite) IntrospectToken ¶
added in v0.4.0
func (f *Fosite) IntrospectToken(ctx context.Context, token string, tokenUse TokenUse, session Session, scopes ...string) (_ TokenUse, _ AccessRequester, err error)
func (*Fosite) NewAccessRequest ¶
func (f *Fosite) NewAccessRequest(ctx context.Context, r *http.Request, session Session) (_ AccessRequester, err error)
Implements

https://tools.ietf.org/html/rfc6749#section-2.3.1 Clients in possession of a client password MAY use the HTTP Basic authentication scheme as defined in [RFC2617] to authenticate with the authorization server. The client identifier is encoded using the "application/x-www-form-urlencoded" encoding algorithm per Appendix B, and the encoded value is used as the username; the client password is encoded using the same algorithm and used as the password. The authorization server MUST support the HTTP Basic authentication scheme for authenticating clients that were issued a client password. Including the client credentials in the request-body using the two parameters is NOT RECOMMENDED and SHOULD be limited to clients unable to directly utilize the HTTP Basic authentication scheme (or other password-based HTTP authentication schemes). The parameters can only be transmitted in the request-body and MUST NOT be included in the request URI.
https://tools.ietf.org/html/rfc6749#section-3.2.1
Confidential clients or other clients issued client credentials MUST authenticate with the authorization server as described in Section 2.3 when making requests to the token endpoint.
If the client type is confidential or the client was issued client credentials (or assigned other authentication requirements), the client MUST authenticate with the authorization server as described in Section 3.2.1.
func (*Fosite) NewAccessResponse ¶
func (f *Fosite) NewAccessResponse(ctx context.Context, requester AccessRequester) (_ AccessResponder, err error)
func (*Fosite) NewAuthorizeRequest ¶
func (f *Fosite) NewAuthorizeRequest(ctx context.Context, r *http.Request) (_ AuthorizeRequester, err error)
func (*Fosite) NewAuthorizeResponse ¶
func (f *Fosite) NewAuthorizeResponse(ctx context.Context, ar AuthorizeRequester, session Session) (_ AuthorizeResponder, err error)
func (*Fosite) NewIntrospectionRequest ¶
added in v0.4.0
func (f *Fosite) NewIntrospectionRequest(ctx context.Context, r *http.Request, session Session) (_ IntrospectionResponder, err error)
NewIntrospectionRequest initiates token introspection as defined in https://tools.ietf.org/search/rfc7662#section-2.1

The protected resource calls the introspection endpoint using an HTTP POST [RFC7231] request with parameters sent as "application/x-www-form-urlencoded" data as defined in [W3C.REC-html5-20141028]. The protected resource sends a parameter representing the token along with optional parameters representing additional context that is known by the protected resource to aid the authorization server in its response.

* token REQUIRED. The string value of the token. For access tokens, this is the "access_token" value returned from the token endpoint defined in OAuth 2.0 [RFC6749], Section 5.1. For refresh tokens, this is the "refresh_token" value returned from the token endpoint as defined in OAuth 2.0 [RFC6749], Section 5.1. Other token types are outside the scope of this specification.

* token_type_hint OPTIONAL. A hint about the type of the token submitted for introspection. The protected resource MAY pass this parameter to help the authorization server optimize the token lookup. If the server is unable to locate the token using the given hint, it MUST extend its search across all of its supported token types. An authorization server MAY ignore this parameter, particularly if it is able to detect the token type automatically. Values for this field are defined in the "OAuth Token Type Hints" registry defined in OAuth Token Revocation [RFC7009].

The introspection endpoint MAY accept other OPTIONAL parameters to provide further context to the query. For instance, an authorization server may desire to know the IP address of the client accessing the protected resource to determine if the correct client is likely to be presenting the token. The definition of this or any other parameters are outside the scope of this specification, to be defined by service documentation or extensions to this specification. If the authorization server is unable to determine the state of the token without additional information, it SHOULD return an introspection response indicating the token is not active as described in Section 2.2.

To prevent token scanning attacks, the endpoint MUST also require some form of authorization to access this endpoint, such as client authentication as described in OAuth 2.0 [RFC6749] or a separate OAuth 2.0 access token such as the bearer token described in OAuth 2.0 Bearer Token Usage [RFC6750]. The methods of managing and validating these authentication credentials are out of scope of this specification.

For example, the following shows a protected resource calling the token introspection endpoint to query about an OAuth 2.0 bearer token. The protected resource is using a separate OAuth 2.0 bearer token to authorize this call.

The following is a non-normative example request:

POST /introspect HTTP/1.1
Host: server.example.com
Accept: application/json
Content-Type: application/x-www-form-urlencoded
Authorization: Bearer 23410913-abewfq.123483

token=2YotnFZFEjr1zCsicMWpAA
In this example, the protected resource uses a client identifier and client secret to authenticate itself to the introspection endpoint. The protected resource also sends a token type hint indicating that it is inquiring about an access token.

The following is a non-normative example request:

POST /introspect HTTP/1.1
Host: server.example.com
Accept: application/json
Content-Type: application/x-www-form-urlencoded
Authorization: Basic czZCaGRSa3F0MzpnWDFmQmF0M2JW

token=mF_9.B5f-4.1JqM&token_type_hint=access_token
func (*Fosite) NewPushedAuthorizeRequest ¶
added in v0.43.0
func (f *Fosite) NewPushedAuthorizeRequest(ctx context.Context, r *http.Request) (_ AuthorizeRequester, err error)
NewPushedAuthorizeRequest validates the request and produces an AuthorizeRequester object that can be stored

func (*Fosite) NewPushedAuthorizeResponse ¶
added in v0.43.0
func (f *Fosite) NewPushedAuthorizeResponse(ctx context.Context, ar AuthorizeRequester, session Session) (_ PushedAuthorizeResponder, err error)
NewPushedAuthorizeResponse executes the handlers and builds the response

func (*Fosite) NewRevocationRequest ¶
added in v0.4.0
func (f *Fosite) NewRevocationRequest(ctx context.Context, r *http.Request) (err error)
NewRevocationRequest handles incoming token revocation requests and validates various parameters as specified in: https://tools.ietf.org/html/rfc7009#section-2.1

The authorization server first validates the client credentials (in case of a confidential client) and then verifies whether the token was issued to the client making the revocation request. If this validation fails, the request is refused and the client is informed of the error by the authorization server as described below.

In the next step, the authorization server invalidates the token. The invalidation takes place immediately, and the token cannot be used again after the revocation.

* https://tools.ietf.org/html/rfc7009#section-2.2 An invalid token type hint value is ignored by the authorization server and does not influence the revocation response.

func (*Fosite) ParseResponseMode ¶
added in v0.36.0
func (f *Fosite) ParseResponseMode(ctx context.Context, r *http.Request, request *AuthorizeRequest) error
func (*Fosite) ResponseModeHandler ¶
added in v0.41.0
func (f *Fosite) ResponseModeHandler(ctx context.Context) ResponseModeHandler
func (*Fosite) WriteAccessError ¶
func (f *Fosite) WriteAccessError(ctx context.Context, rw http.ResponseWriter, req AccessRequester, err error)
func (*Fosite) WriteAccessResponse ¶
func (f *Fosite) WriteAccessResponse(ctx context.Context, rw http.ResponseWriter, requester AccessRequester, responder AccessResponder)
func (*Fosite) WriteAuthorizeError ¶
func (f *Fosite) WriteAuthorizeError(ctx context.Context, rw http.ResponseWriter, ar AuthorizeRequester, err error)
func (*Fosite) WriteAuthorizeResponse ¶
func (f *Fosite) WriteAuthorizeResponse(ctx context.Context, rw http.ResponseWriter, ar AuthorizeRequester, resp AuthorizeResponder)
func (*Fosite) WriteIntrospectionError ¶
added in v0.4.0
func (f *Fosite) WriteIntrospectionError(ctx context.Context, rw http.ResponseWriter, err error)
WriteIntrospectionError responds with token metadata discovered by token introspection as defined in https://tools.ietf.org/search/rfc7662#section-2.2

If the protected resource uses OAuth 2.0 client credentials to authenticate to the introspection endpoint and its credentials are invalid, the authorization server responds with an HTTP 401 (Unauthorized) as described in Section 5.2 of OAuth 2.0 [RFC6749].

If the protected resource uses an OAuth 2.0 bearer token to authorize its call to the introspection endpoint and the token used for authorization does not contain sufficient privileges or is otherwise invalid for this request, the authorization server responds with an HTTP 401 code as described in Section 3 of OAuth 2.0 Bearer Token Usage [RFC6750].

Note that a properly formed and authorized query for an inactive or otherwise invalid token (or a token the protected resource is not allowed to know about) is not considered an error response by this specification. In these cases, the authorization server MUST instead respond with an introspection response with the "active" field set to "false" as described in Section 2.2.

func (*Fosite) WriteIntrospectionResponse ¶
added in v0.4.0
func (f *Fosite) WriteIntrospectionResponse(ctx context.Context, rw http.ResponseWriter, r IntrospectionResponder)
WriteIntrospectionResponse responds with an error if token introspection failed as defined in https://tools.ietf.org/search/rfc7662#section-2.3

The server responds with a JSON object [RFC7159] in "application/ json" format with the following top-level members.

* active REQUIRED. Boolean indicator of whether or not the presented token is currently active. The specifics of a token's "active" state will vary depending on the implementation of the authorization server and the information it keeps about its tokens, but a "true" value return for the "active" property will generally indicate that a given token has been issued by this authorization server, has not been revoked by the resource owner, and is within its given time window of validity (e.g., after its issuance time and before its expiration time). See Section 4 for information on implementation of such checks.

* scope OPTIONAL. A JSON string containing a space-separated list of scopes associated with this token, in the format described in Section 3.3 of OAuth 2.0 [RFC6749].

* client_id OPTIONAL. Client identifier for the OAuth 2.0 client that requested this token.

* username OPTIONAL. Human-readable identifier for the resource owner who authorized this token.

* token_type OPTIONAL. Type of the token as defined in Section 5.1 of OAuth 2.0 [RFC6749].

* exp OPTIONAL. Integer timestamp, measured in the number of seconds since January 1 1970 UTC, indicating when this token will expire, as defined in JWT [RFC7519].

* iat OPTIONAL. Integer timestamp, measured in the number of seconds since January 1 1970 UTC, indicating when this token was originally issued, as defined in JWT [RFC7519].

* nbf OPTIONAL. Integer timestamp, measured in the number of seconds since January 1 1970 UTC, indicating when this token is not to be used before, as defined in JWT [RFC7519].

* sub OPTIONAL. Subject of the token, as defined in JWT [RFC7519]. Usually a machine-readable identifier of the resource owner who authorized this token.

* aud OPTIONAL. Service-specific string identifier or list of string identifiers representing the intended audience for this token, as defined in JWT [RFC7519].

* iss OPTIONAL. String representing the issuer of this token, as defined in JWT [RFC7519].

* jti OPTIONAL. String identifier for the token, as defined in JWT [RFC7519].

Specific implementations MAY extend this structure with their own service-specific response names as top-level members of this JSON object. Response names intended to be used across domains MUST be registered in the "OAuth Token Introspection Response" registry defined in Section 3.1.

The authorization server MAY respond differently to different protected resources making the same request. For instance, an authorization server MAY limit which scopes from a given token are returned for each protected resource to prevent a protected resource from learning more about the larger network than is necessary for its operation.

The response MAY be cached by the protected resource to improve performance and reduce load on the introspection endpoint, but at the cost of liveness of the information used by the protected resource to make authorization decisions. See Section 4 for more information regarding the trade off when the response is cached.

For example, the following response contains a set of information about an active token:

The following is a non-normative example response:

HTTP/1.1 200 OK
Content-Type: application/json

{
  "active": true,
  "client_id": "l238j323ds-23ij4",
  "username": "jdoe",
  "scope": "read write dolphin",
  "sub": "Z5O3upPC88QrAjx00dis",
  "aud": "https://protected.example.net/resource",
  "iss": "https://server.example.com/",
  "exp": 1419356238,
  "iat": 1419350238,
  "extension_field": "twenty-seven"
}
If the introspection call is properly authorized but the token is not active, does not exist on this server, or the protected resource is not allowed to introspect this particular token, then the authorization server MUST return an introspection response with the "active" field set to "false". Note that to avoid disclosing too much of the authorization server's state to a third party, the authorization server SHOULD NOT include any additional information about an inactive token, including why the token is inactive.

The following is a non-normative example response for a token that has been revoked or is otherwise invalid:

HTTP/1.1 200 OK
Content-Type: application/json

{
  "active": false
}
func (*Fosite) WritePushedAuthorizeError ¶
added in v0.43.0
func (f *Fosite) WritePushedAuthorizeError(ctx context.Context, rw http.ResponseWriter, ar AuthorizeRequester, err error)
WritePushedAuthorizeError writes the PAR error

func (*Fosite) WritePushedAuthorizeResponse ¶
added in v0.43.0
func (f *Fosite) WritePushedAuthorizeResponse(ctx context.Context, rw http.ResponseWriter, ar AuthorizeRequester, resp PushedAuthorizeResponder)
WritePushedAuthorizeResponse writes the PAR response

func (*Fosite) WriteRevocationResponse ¶
added in v0.4.0
func (f *Fosite) WriteRevocationResponse(ctx context.Context, rw http.ResponseWriter, err error)
WriteRevocationResponse writes a token revocation response as specified in: https://tools.ietf.org/html/rfc7009#section-2.2

The authorization server responds with HTTP status code 200 if the token has been revoked successfully or if the client submitted an invalid token.

Note: invalid tokens do not cause an error response since the client cannot handle such an error in a reasonable way. Moreover, the purpose of the revocation request, invalidating the particular token, is already achieved.

type G11NContext ¶
added in v0.41.0
type G11NContext interface {
	// GetLang returns the current language in the context
	GetLang() language.Tag
}
G11NContext is the globalization context

type GetJWTMaxDurationProvider ¶
added in v0.43.0
type GetJWTMaxDurationProvider interface {
	// GetJWTMaxDuration returns the JWT max duration.
	GetJWTMaxDuration(ctx context.Context) time.Duration
}
GetJWTMaxDurationProvider returns the provider for configuring the JWT max duration.

type GetSecretsHashingProvider ¶
added in v0.43.0
type GetSecretsHashingProvider interface {
	// GetSecretsHasher returns the client secrets hashing function.
	GetSecretsHasher(ctx context.Context) Hasher
}
GetSecretsHashingProvider provides the client secrets hashing function.

type GlobalSecretProvider ¶
added in v0.43.0
type GlobalSecretProvider interface {
	// GetGlobalSecret returns the global secret.
	GetGlobalSecret(ctx context.Context) ([]byte, error)
}
GlobalSecretProvider returns the provider for configuring the global secret.

type GrantType ¶
added in v0.43.0
type GrantType string
type GrantTypeJWTBearerCanSkipClientAuthProvider ¶
added in v0.43.0
type GrantTypeJWTBearerCanSkipClientAuthProvider interface {
	// GetGrantTypeJWTBearerCanSkipClientAuth returns the grant type JWT bearer can skip client auth.
	GetGrantTypeJWTBearerCanSkipClientAuth(ctx context.Context) bool
}
GrantTypeJWTBearerCanSkipClientAuthProvider returns the provider for configuring the grant type JWT bearer can skip client auth.

type GrantTypeJWTBearerIDOptionalProvider ¶
added in v0.43.0
type GrantTypeJWTBearerIDOptionalProvider interface {
	// GetGrantTypeJWTBearerIDOptional returns the grant type JWT bearer ID optional.
	GetGrantTypeJWTBearerIDOptional(ctx context.Context) bool
}
GrantTypeJWTBearerIDOptionalProvider returns the provider for configuring the grant type JWT bearer ID optional.

type GrantTypeJWTBearerIssuedDateOptionalProvider ¶
added in v0.43.0
type GrantTypeJWTBearerIssuedDateOptionalProvider interface {
	// GetGrantTypeJWTBearerIssuedDateOptional returns the grant type JWT bearer issued date optional.
	GetGrantTypeJWTBearerIssuedDateOptional(ctx context.Context) bool
}
GrantTypeJWTBearerIssuedDateOptionalProvider returns the provider for configuring the grant type JWT bearer issued date optional.

type HMACHashingProvider ¶
added in v0.43.0
type HMACHashingProvider interface {
	// GetHMACHasher returns the hash function.
	GetHMACHasher(ctx context.Context) func() hash.Hash
}
HMACHashingProvider returns the provider for configuring the hash function.

type HTTPClientProvider ¶
added in v0.43.0
type HTTPClientProvider interface {
	// GetHTTPClient returns the HTTP client provider.
	GetHTTPClient(ctx context.Context) *retryablehttp.Client
}
HTTPClientProvider returns the provider for configuring the HTTP client.

type Hasher ¶
added in v0.4.0
type Hasher interface {
	// Compare compares data with a hash and returns an error
	// if the two do not match.
	Compare(ctx context.Context, hash, data []byte) error

	// Hash creates a hash from data or returns an error.
	Hash(ctx context.Context, data []byte) ([]byte, error)
}
Hasher defines how a oauth2-compatible hasher should look like.

type IDTokenIssuerProvider ¶
added in v0.43.0
type IDTokenIssuerProvider interface {
	// GetIDTokenIssuer returns the ID token issuer.
	GetIDTokenIssuer(ctx context.Context) string
}
IDTokenIssuerProvider returns the provider for configuring the ID token issuer.

type IDTokenLifespanProvider ¶
added in v0.43.0
type IDTokenLifespanProvider interface {
	// GetIDTokenLifespan returns the ID token lifespan.
	GetIDTokenLifespan(ctx context.Context) time.Duration
}
IDTokenLifespanProvider returns the provider for configuring the ID token lifespan.

type IntrospectionResponder ¶
added in v0.4.0
type IntrospectionResponder interface {
	// IsActive returns true if the introspected token is active and false otherwise.
	IsActive() bool

	// AccessRequester returns nil when IsActive() is false and the original access request object otherwise.
	GetAccessRequester() AccessRequester

	// GetTokenUse optionally returns the type of the token that was introspected. This could be "access_token", "refresh_token",
	// or if the type can not be determined an empty string.
	GetTokenUse() TokenUse

	//GetAccessTokenType optionally returns the type of the access token that was introspected. This could be "bearer", "mac",
	// or empty string if the type of the token is refresh token.
	GetAccessTokenType() string
}
IntrospectionResponder is the response object that will be returned when token introspection was successful, for example when the client is allowed to perform token introspection. Refer to https://tools.ietf.org/search/rfc7662#section-2.2 for more details.

type IntrospectionResponse ¶
added in v0.4.0
type IntrospectionResponse struct {
	Active          bool            `json:"active"`
	AccessRequester AccessRequester `json:"extra"`
	TokenUse        TokenUse        `json:"token_use,omitempty"`
	AccessTokenType string          `json:"token_type,omitempty"`
	Lang            language.Tag    `json:"-"`
}
func (*IntrospectionResponse) GetAccessRequester ¶
added in v0.4.0
func (r *IntrospectionResponse) GetAccessRequester() AccessRequester
func (*IntrospectionResponse) GetAccessTokenType ¶
added in v0.35.0
func (r *IntrospectionResponse) GetAccessTokenType() string
func (*IntrospectionResponse) GetTokenUse ¶
added in v0.35.0
func (r *IntrospectionResponse) GetTokenUse() TokenUse
func (*IntrospectionResponse) IsActive ¶
added in v0.4.0
func (r *IntrospectionResponse) IsActive() bool
type JWKSFetcherStrategy ¶
added in v0.21.0
type JWKSFetcherStrategy interface {
	// Resolve returns the JSON Web Key Set, or an error if something went wrong. The forceRefresh, if true, forces
	// the strategy to fetch the key from the remote. If forceRefresh is false, the strategy may use a caching strategy
	// to fetch the key.
	Resolve(ctx context.Context, location string, ignoreCache bool) (*jose.JSONWebKeySet, error)
}
JWKSFetcherStrategy is a strategy which pulls (optionally caches) JSON Web Key Sets from a location, typically a client's jwks_uri.

func NewDefaultJWKSFetcherStrategy ¶
added in v0.21.0
func NewDefaultJWKSFetcherStrategy(opts ...func(*DefaultJWKSFetcherStrategy)) JWKSFetcherStrategy
NewDefaultJWKSFetcherStrategy returns a new instance of the DefaultJWKSFetcherStrategy.

type JWKSFetcherStrategyProvider ¶
added in v0.43.0
type JWKSFetcherStrategyProvider interface {
	// GetJWKSFetcherStrategy returns the JWKS fetcher strategy.
	GetJWKSFetcherStrategy(ctx context.Context) JWKSFetcherStrategy
}
JWKSFetcherStrategyProvider returns the provider for configuring the JWKS fetcher strategy.

type JWTScopeFieldProvider ¶
added in v0.43.0
type JWTScopeFieldProvider interface {
	// GetJWTScopeField returns the JWT scope field.
	GetJWTScopeField(ctx context.Context) jwt.JWTScopeFieldEnum
}
JWTScopeFieldProvider returns the provider for configuring the JWT scope field.

type MessageCatalogProvider ¶
added in v0.43.0
type MessageCatalogProvider interface {
	// GetMessageCatalog returns the message catalog.
	GetMessageCatalog(ctx context.Context) i18n.MessageCatalog
}
MessageCatalogProvider returns the provider for configuring the message catalog.

type MinParameterEntropyProvider ¶
added in v0.43.0
type MinParameterEntropyProvider interface {
	// GetMinParameterEntropy returns the minimum parameter entropy.
	GetMinParameterEntropy(_ context.Context) int
}
MinParameterEntropyProvider returns the provider for configuring the minimum parameter entropy.

type OAuth2Provider ¶
type OAuth2Provider interface {
	// NewAuthorizeRequest returns an AuthorizeRequest.
	//
	// The following specs must be considered in any implementation of this method:
	// * https://tools.ietf.org/html/rfc6749#section-3.1
	//	 Extension response types MAY contain a space-delimited (%x20) list of
	//	 values, where the order of values does not matter (e.g., response
	//	 type "a b" is the same as "b a").  The meaning of such composite
	//	 response types is defined by their respective specifications.
	// * https://tools.ietf.org/html/rfc6749#section-3.1.2
	//   The redirection endpoint URI MUST be an absolute URI as defined by
	//   [RFC3986] Section 4.3.  The endpoint URI MAY include an
	//   "application/x-www-form-urlencoded" formatted (per Appendix B) query
	//   component ([RFC3986] Section 3.4), which MUST be retained when adding
	//   additional query parameters.  The endpoint URI MUST NOT include a
	//   fragment component.
	// * https://tools.ietf.org/html/rfc6749#section-3.1.2.2 (everything MUST be implemented)
	NewAuthorizeRequest(ctx context.Context, req *http.Request) (AuthorizeRequester, error)

	// NewAuthorizeResponse iterates through all response type handlers and returns their result or
	// ErrUnsupportedResponseType if none of the handler's were able to handle it.
	//
	// The following specs must be considered in any implementation of this method:
	// * https://tools.ietf.org/html/rfc6749#section-3.1.1
	//	 Extension response types MAY contain a space-delimited (%x20) list of
	//	 values, where the order of values does not matter (e.g., response
	//	 type "a b" is the same as "b a").  The meaning of such composite
	//	 response types is defined by their respective specifications.
	//	 If an authorization request is missing the "response_type" parameter,
	//	 or if the response type is not understood, the authorization server
	//	 MUST return an error response as described in Section 4.1.2.1.
	NewAuthorizeResponse(ctx context.Context, requester AuthorizeRequester, session Session) (AuthorizeResponder, error)

	// WriteAuthorizeError returns the error codes to the redirection endpoint or shows the error to the user, if no valid
	// redirect uri was given. Implements rfc6749#section-4.1.2.1
	//
	// The following specs must be considered in any implementation of this method:
	// * https://tools.ietf.org/html/rfc6749#section-3.1.2
	//   The redirection endpoint URI MUST be an absolute URI as defined by
	//   [RFC3986] Section 4.3.  The endpoint URI MAY include an
	//   "application/x-www-form-urlencoded" formatted (per Appendix B) query
	//   component ([RFC3986] Section 3.4), which MUST be retained when adding
	//   additional query parameters.  The endpoint URI MUST NOT include a
	//   fragment component.
	// * https://tools.ietf.org/html/rfc6749#section-4.1.2.1 (everything)
	// * https://tools.ietf.org/html/rfc6749#section-3.1.2.2 (everything MUST be implemented)
	WriteAuthorizeError(ctx context.Context, rw http.ResponseWriter, requester AuthorizeRequester, err error)

	// WriteAuthorizeResponse persists the AuthorizeSession in the store and redirects the user agent to the provided
	// redirect url or returns an error if storage failed.
	//
	// The following specs must be considered in any implementation of this method:
	// * https://tools.ietf.org/html/rfc6749#rfc6749#section-4.1.2.1
	//   After completing its interaction with the resource owner, the
	//   authorization server directs the resource owner's user-agent back to
	//   the client.  The authorization server redirects the user-agent to the
	//   client's redirection endpoint previously established with the
	//   authorization server during the client registration process or when
	//   making the authorization request.
	// * https://tools.ietf.org/html/rfc6749#section-3.1.2.2 (everything MUST be implemented)
	WriteAuthorizeResponse(ctx context.Context, rw http.ResponseWriter, requester AuthorizeRequester, responder AuthorizeResponder)

	// NewAccessRequest creates a new access request object and validates
	// various parameters.
	//
	// The following specs must be considered in any implementation of this method:
	// * https://tools.ietf.org/html/rfc6749#section-3.2 (everything)
	// * https://tools.ietf.org/html/rfc6749#section-3.2.1 (everything)
	//
	// Furthermore the registered handlers should implement their specs accordingly.
	NewAccessRequest(ctx context.Context, req *http.Request, session Session) (AccessRequester, error)

	// NewAccessResponse creates a new access response and validates that access_token and token_type are set.
	//
	// The following specs must be considered in any implementation of this method:
	// https://tools.ietf.org/html/rfc6749#section-5.1
	NewAccessResponse(ctx context.Context, requester AccessRequester) (AccessResponder, error)

	// WriteAccessError writes an access request error response.
	//
	// The following specs must be considered in any implementation of this method:
	// * https://tools.ietf.org/html/rfc6749#section-5.2 (everything)
	WriteAccessError(ctx context.Context, rw http.ResponseWriter, requester AccessRequester, err error)

	// WriteAccessResponse writes the access response.
	//
	// The following specs must be considered in any implementation of this method:
	// https://tools.ietf.org/html/rfc6749#section-5.1
	WriteAccessResponse(ctx context.Context, rw http.ResponseWriter, requester AccessRequester, responder AccessResponder)

	// NewRevocationRequest handles incoming token revocation requests and validates various parameters.
	//
	// The following specs must be considered in any implementation of this method:
	// https://tools.ietf.org/html/rfc7009#section-2.1
	NewRevocationRequest(ctx context.Context, r *http.Request) error

	// WriteRevocationResponse writes the revoke response.
	//
	// The following specs must be considered in any implementation of this method:
	// https://tools.ietf.org/html/rfc7009#section-2.2
	WriteRevocationResponse(ctx context.Context, rw http.ResponseWriter, err error)

	// IntrospectToken returns token metadata, if the token is valid. Tokens generated by the authorization endpoint,
	// such as the authorization code, can not be introspected.
	IntrospectToken(ctx context.Context, token string, tokenUse TokenUse, session Session, scope ...string) (TokenUse, AccessRequester, error)

	// NewIntrospectionRequest initiates token introspection as defined in
	// https://tools.ietf.org/search/rfc7662#section-2.1
	NewIntrospectionRequest(ctx context.Context, r *http.Request, session Session) (IntrospectionResponder, error)

	// WriteIntrospectionError responds with an error if token introspection failed as defined in
	// https://tools.ietf.org/search/rfc7662#section-2.3
	WriteIntrospectionError(ctx context.Context, rw http.ResponseWriter, err error)

	// WriteIntrospectionResponse responds with token metadata discovered by token introspection as defined in
	// https://tools.ietf.org/search/rfc7662#section-2.2
	WriteIntrospectionResponse(ctx context.Context, rw http.ResponseWriter, r IntrospectionResponder)

	// NewPushedAuthorizeRequest validates the request and produces an AuthorizeRequester object that can be stored
	NewPushedAuthorizeRequest(ctx context.Context, r *http.Request) (AuthorizeRequester, error)

	// NewPushedAuthorizeResponse executes the handlers and builds the response
	NewPushedAuthorizeResponse(ctx context.Context, ar AuthorizeRequester, session Session) (PushedAuthorizeResponder, error)

	// WritePushedAuthorizeResponse writes the PAR response
	WritePushedAuthorizeResponse(ctx context.Context, rw http.ResponseWriter, ar AuthorizeRequester, resp PushedAuthorizeResponder)

	// WritePushedAuthorizeError writes the PAR error
	WritePushedAuthorizeError(ctx context.Context, rw http.ResponseWriter, ar AuthorizeRequester, err error)
}
OAuth2Provider is an interface that enables you to write OAuth2 handlers with only a few lines of code. Check Fosite for an implementation of this interface.

type OmitRedirectScopeParamProvider ¶
added in v0.43.0
type OmitRedirectScopeParamProvider interface {
	// GetOmitRedirectScopeParam must be set to true if the scope query param is to be omitted
	// in the authorization's redirect URI
	GetOmitRedirectScopeParam(ctx context.Context) bool
}
OmitRedirectScopeParamProvider returns the provider for configuring the omit redirect scope param.

type OpenIDConnectClient ¶
added in v0.21.0
type OpenIDConnectClient interface {
	// GetRequestURIs is an array of request_uri values that are pre-registered by the RP for use at the OP. Servers MAY
	// cache the contents of the files referenced by these URIs and not retrieve them at the time they are used in a request.
	// OPs can require that request_uri values used be pre-registered with the require_request_uri_registration
	// discovery parameter.
	GetRequestURIs() []string

	// GetJSONWebKeys returns the JSON Web Key Set containing the public key used by the client to authenticate.
	GetJSONWebKeys() *jose.JSONWebKeySet

	// GetJSONWebKeys returns the URL for lookup of JSON Web Key Set containing the
	// public key used by the client to authenticate.
	GetJSONWebKeysURI() string

	// JWS [JWS] alg algorithm [JWA] that MUST be used for signing Request Objects sent to the OP.
	// All Request Objects from this Client MUST be rejected, if not signed with this algorithm.
	GetRequestObjectSigningAlgorithm() string

	// Requested Client Authentication method for the Token Endpoint. The options are client_secret_post,
	// client_secret_basic, private_key_jwt, and none.
	GetTokenEndpointAuthMethod() string

	// JWS [JWS] alg algorithm [JWA] that MUST be used for signing the JWT [JWT] used to authenticate the
	// Client at the Token Endpoint for the private_key_jwt authentication method.
	GetTokenEndpointAuthSigningAlgorithm() string
}
OpenIDConnectClient represents a client capable of performing OpenID Connect requests.

type PARStorage ¶
added in v0.43.0
type PARStorage interface {
	// CreatePARSession stores the pushed authorization request context. The requestURI is used to derive the key.
	CreatePARSession(ctx context.Context, requestURI string, request AuthorizeRequester) error
	// GetPARSession gets the push authorization request context. The caller is expected to merge the AuthorizeRequest.
	GetPARSession(ctx context.Context, requestURI string) (AuthorizeRequester, error)
	// DeletePARSession deletes the context.
	DeletePARSession(ctx context.Context, requestURI string) (err error)
}
PARStorage holds information needed to store and retrieve PAR context.

type PushedAuthorizeEndpointHandler ¶
added in v0.43.0
type PushedAuthorizeEndpointHandler interface {
	// HandlePushedAuthorizeRequest handles a pushed authorize endpoint request. To extend the handler's capabilities, the http request
	// is passed along, if further information retrieval is required. If the handler feels that he is not responsible for
	// the pushed authorize request, he must return nil and NOT modify session nor responder neither requester.
	HandlePushedAuthorizeEndpointRequest(ctx context.Context, requester AuthorizeRequester, responder PushedAuthorizeResponder) error
}
PushedAuthorizeEndpointHandler is the interface that handles PAR (https://datatracker.ietf.org/doc/html/rfc9126)

type PushedAuthorizeEndpointHandlers ¶
added in v0.43.0
type PushedAuthorizeEndpointHandlers []PushedAuthorizeEndpointHandler
PushedAuthorizeEndpointHandlers is a list of PushedAuthorizeEndpointHandler

func (*PushedAuthorizeEndpointHandlers) Append ¶
added in v0.43.0
func (a *PushedAuthorizeEndpointHandlers) Append(h PushedAuthorizeEndpointHandler)
Append adds an AuthorizeEndpointHandler to this list. Ignores duplicates based on reflect.TypeOf.

type PushedAuthorizeRequestConfigProvider ¶
added in v0.43.0
type PushedAuthorizeRequestConfigProvider interface {
	// GetPushedAuthorizeRequestURIPrefix is the request URI prefix. This is
	// usually 'urn:ietf:params:oauth:request_uri:'.
	GetPushedAuthorizeRequestURIPrefix(ctx context.Context) string

	// GetPushedAuthorizeContextLifespan is the lifespan of the short-lived PAR context.
	GetPushedAuthorizeContextLifespan(ctx context.Context) time.Duration

	// EnforcePushedAuthorize indicates if PAR is enforced. In this mode, a client
	// cannot pass authorize parameters at the 'authorize' endpoint. The 'authorize' endpoint
	// must contain the PAR request_uri.
	EnforcePushedAuthorize(ctx context.Context) bool
}
PushedAuthorizeRequestConfigProvider is the configuration provider for pushed authorization request.

type PushedAuthorizeRequestHandlersProvider ¶
added in v0.43.0
type PushedAuthorizeRequestHandlersProvider interface {
	// GetPushedAuthorizeEndpointHandlers returns the handlers.
	GetPushedAuthorizeEndpointHandlers(ctx context.Context) PushedAuthorizeEndpointHandlers
}
PushedAuthorizeEndpointHandlersProvider returns the provider for configuring the PAR handlers.

type PushedAuthorizeResponder ¶
added in v0.43.0
type PushedAuthorizeResponder interface {
	// GetRequestURI returns the request_uri
	GetRequestURI() string
	// SetRequestURI sets the request_uri
	SetRequestURI(requestURI string)
	// GetExpiresIn gets the expires_in
	GetExpiresIn() int
	// SetExpiresIn sets the expires_in
	SetExpiresIn(seconds int)

	// GetHeader returns the response's header
	GetHeader() (header http.Header)

	// AddHeader adds an header key value pair to the response
	AddHeader(key, value string)

	// SetExtra sets a key value pair for the response.
	SetExtra(key string, value interface{})

	// GetExtra returns a key's value.
	GetExtra(key string) interface{}

	// ToMap converts the response to a map.
	ToMap() map[string]interface{}
}
PushedAuthorizeResponder is the response object for PAR

type PushedAuthorizeResponse ¶
added in v0.43.0
type PushedAuthorizeResponse struct {
	RequestURI string `json:"request_uri"`
	ExpiresIn  int    `json:"expires_in"`
	Header     http.Header
	Extra      map[string]interface{}
}
PushedAuthorizeResponse is the response object for PAR

func (*PushedAuthorizeResponse) AddHeader ¶
added in v0.43.0
func (a *PushedAuthorizeResponse) AddHeader(key, value string)
AddHeader adds

func (*PushedAuthorizeResponse) GetExpiresIn ¶
added in v0.43.0
func (a *PushedAuthorizeResponse) GetExpiresIn() int
GetExpiresIn gets

func (*PushedAuthorizeResponse) GetExtra ¶
added in v0.43.0
func (a *PushedAuthorizeResponse) GetExtra(key string) interface{}
GetExtra gets

func (*PushedAuthorizeResponse) GetHeader ¶
added in v0.43.0
func (a *PushedAuthorizeResponse) GetHeader() http.Header
GetHeader gets

func (*PushedAuthorizeResponse) GetRequestURI ¶
added in v0.43.0
func (a *PushedAuthorizeResponse) GetRequestURI() string
GetRequestURI gets

func (*PushedAuthorizeResponse) SetExpiresIn ¶
added in v0.43.0
func (a *PushedAuthorizeResponse) SetExpiresIn(seconds int)
SetExpiresIn sets

func (*PushedAuthorizeResponse) SetExtra ¶
added in v0.43.0
func (a *PushedAuthorizeResponse) SetExtra(key string, value interface{})
SetExtra sets

func (*PushedAuthorizeResponse) SetRequestURI ¶
added in v0.43.0
func (a *PushedAuthorizeResponse) SetRequestURI(requestURI string)
SetRequestURI sets

func (*PushedAuthorizeResponse) ToMap ¶
added in v0.43.0
func (a *PushedAuthorizeResponse) ToMap() map[string]interface{}
ToMap converts to a map

type RFC6749Error ¶
type RFC6749Error struct {
	ErrorField       string
	DescriptionField string
	HintField        string
	CodeField        int
	DebugField       string
	// contains filtered or unexported fields
}
func ErrorToRFC6749Error ¶
func ErrorToRFC6749Error(err error) *RFC6749Error
func (*RFC6749Error) Cause ¶
added in v0.34.0
func (e *RFC6749Error) Cause() error
func (*RFC6749Error) Debug ¶
func (e *RFC6749Error) Debug() string
func (RFC6749Error) Error ¶
added in v0.7.0
func (e RFC6749Error) Error() string
func (*RFC6749Error) GetDescription ¶
added in v0.33.0
func (e *RFC6749Error) GetDescription() string
GetDescription returns a more description description, combined with hint and debug (when available).

func (RFC6749Error) Is ¶
added in v0.33.0
func (e RFC6749Error) Is(err error) bool
func (RFC6749Error) MarshalJSON ¶
added in v0.33.0
func (e RFC6749Error) MarshalJSON() ([]byte, error)
func (*RFC6749Error) Reason ¶
added in v0.7.0
func (e *RFC6749Error) Reason() string
func (*RFC6749Error) RequestID ¶
added in v0.7.0
func (e *RFC6749Error) RequestID() string
func (*RFC6749Error)
Sanitize
DEPRECATED
added in v0.33.0
func (*RFC6749Error) StackTrace ¶
added in v0.36.0
func (e *RFC6749Error) StackTrace() (trace errors.StackTrace)
StackTrace returns the error's stack trace.

func (*RFC6749Error) Status ¶
added in v0.7.0
func (e *RFC6749Error) Status() string
func (*RFC6749Error) StatusCode ¶
func (e *RFC6749Error) StatusCode() int
func (*RFC6749Error) ToValues ¶
added in v0.33.0
func (e *RFC6749Error) ToValues() url.Values
func (*RFC6749Error) UnmarshalJSON ¶
added in v0.33.0
func (e *RFC6749Error) UnmarshalJSON(b []byte) error
func (RFC6749Error) Unwrap ¶
added in v0.34.0
func (e RFC6749Error) Unwrap() error
func (*RFC6749Error) WithDebug ¶
added in v0.15.0
func (e *RFC6749Error) WithDebug(debug string) *RFC6749Error
func (*RFC6749Error) WithDebugf ¶
added in v0.21.0
func (e *RFC6749Error) WithDebugf(debug string, args ...interface{}) *RFC6749Error
func (*RFC6749Error) WithDescription ¶
added in v0.16.4
func (e *RFC6749Error) WithDescription(description string) *RFC6749Error
func (*RFC6749Error) WithExposeDebug ¶
added in v0.36.0
func (e *RFC6749Error) WithExposeDebug(exposeDebug bool) *RFC6749Error
WithExposeDebug if set to true exposes debug messages

func (*RFC6749Error) WithHint ¶
added in v0.21.0
func (e *RFC6749Error) WithHint(hint string) *RFC6749Error
func (*RFC6749Error) WithHintIDOrDefaultf ¶
added in v0.41.0
func (e *RFC6749Error) WithHintIDOrDefaultf(ID string, def string, args ...interface{}) *RFC6749Error
WithHintIDOrDefaultf accepts the ID of the hint message

func (*RFC6749Error) WithHintTranslationID ¶
added in v0.41.0
func (e *RFC6749Error) WithHintTranslationID(ID string) *RFC6749Error
WithHintTranslationID accepts the ID of the hint message and should be paired with WithHint and WithHintf to add a default message and vaargs.

func (*RFC6749Error) WithHintf ¶
added in v0.21.0
func (e *RFC6749Error) WithHintf(hint string, args ...interface{}) *RFC6749Error
func (RFC6749Error) WithLegacyFormat ¶
added in v0.36.0
func (e RFC6749Error) WithLegacyFormat(useLegacyFormat bool) *RFC6749Error
func (*RFC6749Error) WithLocalizer ¶
added in v0.41.0
func (e *RFC6749Error) WithLocalizer(catalog i18n.MessageCatalog, lang language.Tag) *RFC6749Error
func (*RFC6749Error) WithTrace ¶
added in v0.36.0
func (e *RFC6749Error) WithTrace(err error) *RFC6749Error
func (RFC6749Error) WithWrap ¶
added in v0.36.0
func (e RFC6749Error) WithWrap(cause error) *RFC6749Error
func (*RFC6749Error) Wrap ¶
added in v0.36.0
func (e *RFC6749Error) Wrap(err error)
type RFC6749ErrorJson ¶
added in v0.33.0
type RFC6749ErrorJson struct {
	Name        string `json:"error"`
	Description string `json:"error_description"`
	Hint        string `json:"error_hint,omitempty"`
	Code        int    `json:"status_code,omitempty"`
	Debug       string `json:"error_debug,omitempty"`
}
RFC6749ErrorJson is a helper struct for JSON encoding/decoding of RFC6749Error.

type RedirectSecureCheckerProvider ¶
added in v0.43.0
type RedirectSecureCheckerProvider interface {
	// GetRedirectSecureChecker returns the redirect URL security validator.
	GetRedirectSecureChecker(ctx context.Context) func(context.Context, *url.URL) bool
}
RedirectSecureCheckerProvider returns the provider for configuring the redirect URL security validator.

type RefreshTokenLifespanProvider ¶
added in v0.43.0
type RefreshTokenLifespanProvider interface {
	// GetRefreshTokenLifespan returns the refresh token lifespan.
	GetRefreshTokenLifespan(ctx context.Context) time.Duration
}
RefreshTokenLifespanProvider returns the provider for configuring the refresh token lifespan.

type RefreshTokenScopesProvider ¶
added in v0.43.0
type RefreshTokenScopesProvider interface {
	// GetRefreshTokenScopes returns the refresh token scopes.
	GetRefreshTokenScopes(ctx context.Context) []string
}
RefreshTokenScopesProvider returns the provider for configuring the refresh token scopes.

type Request ¶
type Request struct {
	ID                string       `json:"id" gorethink:"id"`
	RequestedAt       time.Time    `json:"requestedAt" gorethink:"requestedAt"`
	Client            Client       `json:"client" gorethink:"client"`
	RequestedScope    Arguments    `json:"scopes" gorethink:"scopes"`
	GrantedScope      Arguments    `json:"grantedScopes" gorethink:"grantedScopes"`
	Form              url.Values   `json:"form" gorethink:"form"`
	Session           Session      `json:"session" gorethink:"session"`
	RequestedAudience Arguments    `json:"requestedAudience"`
	GrantedAudience   Arguments    `json:"grantedAudience"`
	Lang              language.Tag `json:"-"`
}
Request is an implementation of Requester

func NewRequest ¶
func NewRequest() *Request
func (*Request) AppendRequestedAudience ¶
added in v0.27.0
func (a *Request) AppendRequestedAudience(audience string)
func (*Request) AppendRequestedScope ¶
added in v0.2.0
func (a *Request) AppendRequestedScope(scope string)
func (*Request) GetClient ¶
func (a *Request) GetClient() Client
func (*Request) GetGrantedAudience ¶
added in v0.27.0
func (a *Request) GetGrantedAudience() Arguments
func (*Request) GetGrantedScopes ¶
func (a *Request) GetGrantedScopes() Arguments
func (*Request) GetID ¶
added in v0.4.0
func (a *Request) GetID() string
func (*Request) GetLang ¶
added in v0.41.0
func (a *Request) GetLang() language.Tag
func (*Request) GetRequestForm ¶
func (a *Request) GetRequestForm() url.Values
func (*Request) GetRequestedAt ¶
func (a *Request) GetRequestedAt() time.Time
func (*Request) GetRequestedAudience ¶
added in v0.27.0
func (a *Request) GetRequestedAudience() (audience Arguments)
func (*Request) GetRequestedScopes ¶
added in v0.2.0
func (a *Request) GetRequestedScopes() Arguments
func (*Request) GetSession ¶
func (a *Request) GetSession() Session
func (*Request) GrantAudience ¶
added in v0.27.0
func (a *Request) GrantAudience(audience string)
func (*Request) GrantScope ¶
func (a *Request) GrantScope(scope string)
func (*Request) Merge ¶
func (a *Request) Merge(request Requester)
func (*Request) Sanitize ¶
added in v0.17.0
func (a *Request) Sanitize(allowedParameters []string) Requester
func (*Request) SetID ¶
added in v0.15.0
func (a *Request) SetID(id string)
func (*Request) SetRequestedAudience ¶
added in v0.27.0
func (a *Request) SetRequestedAudience(s Arguments)
func (*Request) SetRequestedScopes ¶
added in v0.2.0
func (a *Request) SetRequestedScopes(s Arguments)
func (*Request) SetSession ¶
func (a *Request) SetSession(session Session)
type Requester ¶
type Requester interface {
	// SetID sets the unique identifier.
	SetID(id string)

	// GetID returns a unique identifier.
	GetID() string

	// GetRequestedAt returns the time the request was created.
	GetRequestedAt() (requestedAt time.Time)

	// GetClient returns the request's client.
	GetClient() (client Client)

	// GetRequestedScopes returns the request's scopes.
	GetRequestedScopes() (scopes Arguments)

	// GetRequestedAudience returns the requested audiences for this request.
	GetRequestedAudience() (audience Arguments)

	// SetRequestedScopes sets the request's scopes.
	SetRequestedScopes(scopes Arguments)

	// SetRequestedAudience sets the requested audience.
	SetRequestedAudience(audience Arguments)

	// AppendRequestedScope appends a scope to the request.
	AppendRequestedScope(scope string)

	// GetGrantScopes returns all granted scopes.
	GetGrantedScopes() (grantedScopes Arguments)

	// GetGrantedAudience returns all granted audiences.
	GetGrantedAudience() (grantedAudience Arguments)

	// GrantScope marks a request's scope as granted.
	GrantScope(scope string)

	// GrantAudience marks a request's audience as granted.
	GrantAudience(audience string)

	// GetSession returns a pointer to the request's session or nil if none is set.
	GetSession() (session Session)

	// SetSession sets the request's session pointer.
	SetSession(session Session)

	// GetRequestForm returns the request's form input.
	GetRequestForm() url.Values

	// Merge merges the argument into the method receiver.
	Merge(requester Requester)

	// Sanitize returns a sanitized clone of the request which can be used for storage.
	Sanitize(allowedParameters []string) Requester
}
Requester is an abstract interface for handling requests in Fosite.

type ResponseModeClient ¶
added in v0.36.0
type ResponseModeClient interface {
	// GetResponseMode returns the response modes that client is allowed to send
	GetResponseModes() []ResponseModeType
}
ResponseModeClient represents a client capable of handling response_mode

type ResponseModeHandler ¶
added in v0.41.0
type ResponseModeHandler interface {
	// ResponseModes returns a set of supported response modes handled
	// by the interface implementation.
	//
	// In an authorize request with any of the provide response modes
	// methods `WriteAuthorizeResponse` and `WriteAuthorizeError` will be
	// invoked to write the successful or error authorization responses respectively.
	ResponseModes() ResponseModeTypes

	// WriteAuthorizeResponse writes successful responses
	//
	// Following headers are expected to be set by default:
	// header.Set("Cache-Control", "no-store")
	// header.Set("Pragma", "no-cache")
	WriteAuthorizeResponse(ctx context.Context, rw http.ResponseWriter, ar AuthorizeRequester, resp AuthorizeResponder)

	// WriteAuthorizeError writes error responses
	//
	// Following headers are expected to be set by default:
	// header.Set("Cache-Control", "no-store")
	// header.Set("Pragma", "no-cache")
	WriteAuthorizeError(ctx context.Context, rw http.ResponseWriter, ar AuthorizeRequester, err error)
}
ResponseModeHandler provides a contract for handling custom response modes

type ResponseModeHandlerExtensionProvider ¶
added in v0.43.0
type ResponseModeHandlerExtensionProvider interface {
	// GetResponseModeHandlerExtension returns the response mode handler extension.
	GetResponseModeHandlerExtension(ctx context.Context) ResponseModeHandler
}
ResponseModeHandlerExtensionProvider returns the provider for configuring the response mode handler extension.

type ResponseModeType ¶
added in v0.36.0
type ResponseModeType string
type ResponseModeTypes ¶
added in v0.41.0
type ResponseModeTypes []ResponseModeType
func (ResponseModeTypes) Has ¶
added in v0.41.0
func (rs ResponseModeTypes) Has(item ResponseModeType) bool
type RevocationHandler ¶
added in v0.4.0
type RevocationHandler interface {
	// RevokeToken handles access and refresh token revocation.
	RevokeToken(ctx context.Context, token string, tokenType TokenType, client Client) error
}
RevocationHandler is the interface that allows token revocation for an OAuth2.0 provider. https://tools.ietf.org/html/rfc7009

RevokeToken is invoked after a new token revocation request is parsed.

https://tools.ietf.org/html/rfc7009#section-2.1 If the particular token is a refresh token and the authorization server supports the revocation of access tokens, then the authorization server SHOULD also invalidate all access tokens based on the same authorization grant (see Implementation Note). If the token passed to the request is an access token, the server MAY revoke the respective refresh token as well.

type RevocationHandlers ¶
added in v0.4.0
type RevocationHandlers []RevocationHandler
RevocationHandlers is a list of RevocationHandler

func (*RevocationHandlers) Append ¶
added in v0.4.0
func (t *RevocationHandlers) Append(h RevocationHandler)
Append adds an RevocationHandler to this list. Ignores duplicates based on reflect.TypeOf.

type RevocationHandlersProvider ¶
added in v0.43.0
type RevocationHandlersProvider interface {
	// GetRevocationHandlers returns the revocation handlers.
	GetRevocationHandlers(ctx context.Context) RevocationHandlers
}
RevocationHandlersProvider returns the provider for configuring the revocation handlers.

type RotatedGlobalSecretsProvider ¶
added in v0.43.0
type RotatedGlobalSecretsProvider interface {
	// GetRotatedGlobalSecrets returns the rotated global secrets.
	GetRotatedGlobalSecrets(ctx context.Context) ([][]byte, error)
}
RotatedGlobalSecretsProvider returns the provider for configuring the rotated global secrets.

type SanitationAllowedProvider ¶
added in v0.43.0
type SanitationAllowedProvider interface {
	// GetSanitationWhiteList is a whitelist of form values that are required by the token endpoint. These values
	// are safe for storage in a database (cleartext).
	GetSanitationWhiteList(ctx context.Context) []string
}
SanitationAllowedProvider returns the provider for configuring the sanitation white list.

type ScopeStrategy ¶
added in v0.2.0
type ScopeStrategy func(haystack []string, needle string) bool
ScopeStrategy is a strategy for matching scopes.

type ScopeStrategyProvider ¶
added in v0.43.0
type ScopeStrategyProvider interface {
	// GetScopeStrategy returns the scope strategy.
	GetScopeStrategy(ctx context.Context) ScopeStrategy
}
ScopeStrategyProvider returns the provider for configuring the scope strategy.

type SendDebugMessagesToClientsProvider ¶
added in v0.43.0
type SendDebugMessagesToClientsProvider interface {
	// GetSendDebugMessagesToClients returns the send debug messages to clients.
	GetSendDebugMessagesToClients(ctx context.Context) bool
}
SendDebugMessagesToClientsProvider returns the provider for configuring the send debug messages to clients.

type Session ¶
added in v0.5.0
type Session interface {
	// SetExpiresAt sets the expiration time of a token.
	//
	//  session.SetExpiresAt(fosite.AccessToken, time.Now().UTC().Add(time.Hour))
	SetExpiresAt(key TokenType, exp time.Time)

	// GetExpiresAt returns the expiration time of a token if set, or time.IsZero() if not.
	//
	//  session.GetExpiresAt(fosite.AccessToken)
	GetExpiresAt(key TokenType) time.Time

	// GetUsername returns the username, if set. This is optional and only used during token introspection.
	GetUsername() string

	// GetSubject returns the subject, if set. This is optional and only used during token introspection.
	GetSubject() string

	// Clone clones the session.
	Clone() Session
}
Session is an interface that is used to store session data between OAuth2 requests. It can be used to look up when a session expires or what the subject's name was.

type Storage ¶
type Storage interface {
	ClientManager
}
Storage defines fosite's minimal storage interface.

type TokenEndpointHandler ¶
type TokenEndpointHandler interface {
	// PopulateTokenEndpointResponse is responsible for setting return values and should only be executed if
	// the handler's HandleTokenEndpointRequest did not return ErrUnknownRequest.
	PopulateTokenEndpointResponse(ctx context.Context, requester AccessRequester, responder AccessResponder) error

	// HandleTokenEndpointRequest handles an authorize request. If the handler is not responsible for handling
	// the request, this method should return ErrUnknownRequest and otherwise handle the request.
	HandleTokenEndpointRequest(ctx context.Context, requester AccessRequester) error

	// CanSkipClientAuth indicates if client authentication can be skipped. By default it MUST be false, unless you are
	// implementing extension grant type, which allows unauthenticated client. CanSkipClientAuth must be called
	// before HandleTokenEndpointRequest to decide, if AccessRequester will contain authenticated client.
	CanSkipClientAuth(ctx context.Context, requester AccessRequester) bool

	// CanHandleRequest indicates, if TokenEndpointHandler can handle this request or not. If true,
	// HandleTokenEndpointRequest can be called.
	CanHandleTokenEndpointRequest(ctx context.Context, requester AccessRequester) bool
}
type TokenEndpointHandlers ¶
type TokenEndpointHandlers []TokenEndpointHandler
TokenEndpointHandlers is a list of TokenEndpointHandler

func (*TokenEndpointHandlers) Append ¶
func (t *TokenEndpointHandlers) Append(h TokenEndpointHandler)
Append adds an TokenEndpointHandler to this list. Ignores duplicates based on reflect.TypeOf.

type TokenEndpointHandlersProvider ¶
added in v0.43.0
type TokenEndpointHandlersProvider interface {
	// GetTokenEndpointHandlers returns the token endpoint handlers.
	GetTokenEndpointHandlers(ctx context.Context) TokenEndpointHandlers
}
TokenEndpointHandlersProvider returns the provider for configuring the token endpoint handlers.

type TokenEntropyProvider ¶
added in v0.43.0
type TokenEntropyProvider interface {
	// GetTokenEntropy returns the token entropy.
	GetTokenEntropy(ctx context.Context) int
}
TokenEntropyProvider returns the provider for configuring the token entropy.

type TokenIntrospectionHandlers ¶
added in v0.4.0
type TokenIntrospectionHandlers []TokenIntrospector
TokenIntrospectionHandlers is a list of TokenValidator

func (*TokenIntrospectionHandlers) Append ¶
added in v0.4.0
func (t *TokenIntrospectionHandlers) Append(h TokenIntrospector)
Append adds an AccessTokenValidator to this list. Ignores duplicates based on reflect.TypeOf.

type TokenIntrospectionHandlersProvider ¶
added in v0.43.0
type TokenIntrospectionHandlersProvider interface {
	// GetTokenIntrospectionHandlers returns the token introspection handlers.
	GetTokenIntrospectionHandlers(ctx context.Context) TokenIntrospectionHandlers
}
TokenIntrospectionHandlersProvider returns the provider for configuring the token introspection handlers.

type TokenIntrospector ¶
added in v0.4.0
type TokenIntrospector interface {
	IntrospectToken(ctx context.Context, token string, tokenUse TokenUse, accessRequest AccessRequester, scopes []string) (TokenUse, error)
}
type TokenType ¶
added in v0.2.0
type TokenType string
type TokenURLProvider ¶
added in v0.43.0
type TokenURLProvider interface {
	// GetTokenURLs returns the token URL.
	GetTokenURLs(ctx context.Context) []string
}
type TokenUse ¶
added in v0.35.0
type TokenUse = TokenType
type UseLegacyErrorFormatProvider ¶
added in v0.43.0
type UseLegacyErrorFormatProvider interface {
	// GetUseLegacyErrorFormat returns whether to use the legacy error format.
	//
	// DEPRECATED: Do not use this flag anymore.
	GetUseLegacyErrorFormat(ctx context.Context) bool
}
UseLegacyErrorFormatProvider returns the provider for configuring whether to use the legacy error format.

DEPRECATED: Do not use this flag anymore.

type VerifiableCredentialsNonceLifespanProvider ¶
added in v0.45.0
type VerifiableCredentialsNonceLifespanProvider interface {
	// GetNonceLifespan returns the nonce lifespan.
	GetVerifiableCredentialsNonceLifespan(ctx context.Context) time.Duration
}
VerifiableCredentialsNonceLifespanProvider returns the provider for configuring the access token lifespan.
