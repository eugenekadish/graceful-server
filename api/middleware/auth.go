package middleware

import (
	"context"
	"fmt"
	"runtime"

	"net/http"

	"github.com/dgrijalva/jwt-go"

	"gitlab.ido-services.com/luxtrust/auth"
	"gitlab.ido-services.com/luxtrust/base-component/util"
	"gitlab.ido-services.com/luxtrust/logging"
)

// User is an internal representation of a user per HTTP request.
type User struct {
	token  *jwt.Token
	claims *auth.CustomClaims
}

// SetToken sets an authentication token for the user session.
func (u *User) SetToken(token *jwt.Token) error {

	var ok bool

	if u.claims, ok = token.Claims.(*auth.CustomClaims); !ok {
		return fmt.Errorf("interface %v could not implement *auth.CustomClaims", token.Claims)
	}

	u.token = token

	return nil
}

// GetToken extracts the validation token from User.
func (u *User) GetToken() *jwt.Token {
	return u.token
}

// GetClaims extracts the claim from User.
func (u *User) GetClaims() *auth.CustomClaims {
	return u.claims
}

// AuthValidator provides functionality necessary for authentication with a bearer token.
type AuthValidator interface {
	GenerateToken(string) (string, error)
	GenerateOnBehalfToken(string, string) (string, error)
	ValidateToken(string) (*jwt.Token, error)
	Authenticate(string) (bool, *jwt.Token, error)
}

// AuthMiddleware provides a middleware that verifies required authentication for endpoints.
type AuthMiddleware struct {
	logger    logging.Logger
	validator AuthValidator
}

// AuthOption provides the client a callback that is used to dynamically specify attributes for an
// AuthMiddleware.
type AuthOption func(*AuthMiddleware)

// WithAuthLogger is used for specifying the Logger for an AuthMiddleware.
func WithAuthLogger(logger logging.Logger) AuthOption {
	return func(am *AuthMiddleware) { am.logger = logger }
}

// WithAuthValidator is used for specifying a token validator for an AuthMiddleware.
func WithAuthValidator(validator AuthValidator) AuthOption {
	return func(am *AuthMiddleware) { am.validator = validator }
}

// NewAuthMiddleware is a variadic constructor for an AuthMiddleware.
func NewAuthMiddleware(cfg Configurer, opts ...AuthOption) *AuthMiddleware {

	var err error

	var am *AuthMiddleware
	var validator AuthValidator

	// TODO: Provide better defaults or supply the `component name`, `version` and `build` information
	// in through the constructor.
	var defaultLoggerInfo = logging.DefaultLoggerInfo{
		Build:           util.Build,
		Component:       cfg.GetString("name"),
		APIVersion:      util.APIVersion,
		SoftwareVersion: util.SoftwareVersion,
	}
	var log = logging.New(defaultLoggerInfo, "json")
	var publicKey = []byte(cfg.GetString("api.public_key"))

	// NOTE: This default is probably a bad idea if we do NOT want to hold an unused key
	if validator, err = auth.NewValidator(publicKey); err != nil {
		// TODO: Clean this up in some kind of wrapper function
		// Could use: https://golang.org/pkg/log/#pkg-constants
		_, file, line, _ := runtime.Caller(1)
		log.
			WithError(err).
			Fatalf("%s:%d %v", file, line, err)
	}

	am = &AuthMiddleware{
		logger:    log,
		validator: validator,
	}

	for _, opt := range opts {
		opt(am)
	}

	return am
}

// Wrapper is a pass through function for handlers that implicitly performs additional business
// logic per request.
func (am *AuthMiddleware) Wrapper(next http.Handler) http.Handler {

	// NOTE: There is no error handeling in this function
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var (
			valid bool

			err   error
			token string

			authUser    User
			parsedToken *jwt.Token
		)

		if token, err = auth.GetTokenFromHeader(r.Header); err != nil {
			_, file, line, _ := runtime.Caller(0)

			// QUESTION: Why is this error logged asynchronously?
			go am.logger.
				WithError(err).
				WithField("endpoint", r.RequestURI).
				WithField("remote_addr", r.RemoteAddr).
				Errorf("%s:%d %v", file, line, err)

			w.WriteHeader(http.StatusUnauthorized)

			return
		}

		if valid, parsedToken, _ = am.validator.Authenticate(token); !valid {

			w.WriteHeader(http.StatusForbidden)

			return
		}

		if err = authUser.SetToken(parsedToken); err != nil {
			// TODO: Clean this up in some kind of wrapper function
			// Could use: https://golang.org/pkg/log/#pkg-constants
			_, file, line, _ := runtime.Caller(0)

			am.logger.
				WithError(err).
				Fatalf("%s:%d %v", file, line, err)
		}

		// QUESTION: Use a `string` instead of `User{}` as the key here?
		ctx := context.WithValue(r.Context(), User{}, authUser)
		req := r.WithContext(ctx)

		next.ServeHTTP(w, req) // next(w, req)
	})
}
