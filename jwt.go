package middleware

import (
	"context"
	"net/http"

	"github.com/dgrijalva/jwt-go"
)

// JWTFunc defines a user supplied authorisation function.
// The func is given the current context and a valid MapClaims
// This is the point at which the user can do further validation / authorisation on the claims.JWTFunc
// The context returned will be used at the context for further chained http handlers.
// JWT authorisation fails if this returns an error, and further chained http handlers are not called.
type JWTFunc func(context.Context, jwt.MapClaims) (context.Context, error)

// JWTOptions defines the user supplied JWT configuration options.
type JWTOptions struct {
	secret   []byte
	authFunc JWTFunc
}

// JWT is middleware which handles authentication for JsonWebTokens
func JWT(options JWTOptions) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		authenticater := jwtAuth{secret: options.secret, userSuppliedFunc: options.authFunc}

		return Auth(authenticater.authenticate)(next)
	}
}

// jwtAuth is the private version of JWTOptions which contains the authentication function passed to Auth middleware
type jwtAuth struct {
	secret           []byte
	userSuppliedFunc JWTFunc
}

func (auth jwtAuth) authenticate(ctx context.Context, tokenString string) (context.Context, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return auth.secret, nil
	})
	if err != nil {
		return ctx, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// fmt.Printf("%+v\n", token)
		// fmt.Printf("%+v\n", claims)
		if auth.userSuppliedFunc != nil {
			return auth.userSuppliedFunc(ctx, claims)
		}
		return ctx, nil
	}

	// fmt.Println(err)
	return ctx, err
}
