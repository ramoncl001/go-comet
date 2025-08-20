package comet

import (
	"context"
	"crypto/hmac"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ramoncl001/comet/ioc"
)

var errInvalidToken = errors.New("invalid token")

const (
	ClaimAudience  = "aud"
	ClaimExpiresAt = "exp"
	ClaimID        = "jti"
	ClaimIssuedAt  = "iat"
	ClaimIssuer    = "iss"
	ClaimNotBefore = "nbf"
	ClaimSubject   = "sub"
	ClaimSessionID = "sid"
)

type JwtProvider interface {
	GenerateToken(claims Claims, secret string) string
	ValidateToken(token string, secret string) (Claims, error)
}

type JwtConfigurations struct {
	Issuer     string
	Audience   string
	Expiration int64
	SecretKey  string
}

type DefaultJwtSessionManager struct {
	SessionManager
	config      JwtConfigurations
	provider    JwtProvider
	userManager UserManager
}

func NewDefaultJwtSessionManager(config JwtConfigurations, provider JwtProvider, userManager UserManager) SessionManager {
	return &DefaultJwtSessionManager{
		config:      config,
		provider:    provider,
		userManager: userManager,
	}
}

func (sm *DefaultJwtSessionManager) Validate(req *Request) (Claims, error) {
	authHeader := req.Headers["Authorization"][0]
	if authHeader == "" {
		return nil, errInvalidToken
	}
	return sm.provider.ValidateToken(strings.ReplaceAll(authHeader, "Bearer ", ""), sm.config.SecretKey)
}

func (sm *DefaultJwtSessionManager) GetUser(ctx context.Context) (ApplicationUser, error) {
	id := ctx.Value("user_id")
	if id == nil {
		return nil, errors.New("session not started")
	}

	user := sm.userManager.GetByID(fmt.Sprintf("%v", id))
	if user == nil {
		return nil, errors.New("user does not exists")
	}

	return user, nil
}

func (sm *DefaultJwtSessionManager) GetToken(claims Claims) string {
	return sm.provider.GenerateToken(claims, sm.config.SecretKey)
}

var DefaultJwtAuthenticationMiddleware = func(next RequestHandler) RequestHandler {
	return func(req *Request) Response {
		manager, err := ioc.ResolveTransient[SessionManager](req.Context())
		if err != nil {
			return Error("error resolving dependency")
		}

		claims, err := manager.Validate(req)
		if err != nil {
			return Unauthorized()
		}

		req = req.WithContext(context.WithValue(req.Context(), "user_id", claims["sub"]))

		return next(req)
	}
}

type defaultJwtProvider struct {
	JwtProvider
}

func NewDefaultJwtProvider() JwtProvider {
	return &defaultJwtProvider{}
}

func (pv *defaultJwtProvider) GenerateToken(claims Claims, secret string) string {
	header := struct {
		Alg string `json:"alg"`
		Typ string `json:"typ"`
	}{
		Alg: "HS256",
		Typ: "JWT",
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		panic(err)
	}
	headerEncoded := base64.RawURLEncoding.EncodeToString(headerJSON)

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		panic(err)
	}
	claimsEncoded := base64.RawURLEncoding.EncodeToString(claimsJSON)

	signatureData := fmt.Sprintf("%s.%s", headerEncoded, claimsEncoded)
	signature := hmac_sha256(signatureData, secret)

	return fmt.Sprintf("%s.%s.%s", headerEncoded, claimsEncoded, signature)
}

func (pv *defaultJwtProvider) ValidateToken(token string, secret string) (Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errInvalidToken
	}

	signatureData := fmt.Sprintf("%s.%s", parts[0], parts[1])
	calculatedSig := hmac_sha256(signatureData, secret)

	if !hmac.Equal([]byte(calculatedSig), []byte(parts[2])) {
		return nil, errInvalidToken
	}

	claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errInvalidToken
	}

	var claims Claims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, errInvalidToken
	}

	currentTime := time.Now().Unix()

	if exp, ok := claims[ClaimExpiresAt].(float64); ok {
		if currentTime > int64(exp) {
			return nil, errInvalidToken
		}
	}

	if nbf, ok := claims[ClaimNotBefore].(float64); ok {
		if currentTime < int64(nbf) {
			return nil, errInvalidToken
		}
	}

	return claims, nil
}
