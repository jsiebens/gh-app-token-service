package main

import (
	"context"
	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/go-github/v57/github"
	"github.com/labstack/echo/v4"
	"github.com/mitchellh/go-homedir"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	Issuer       = "https://token.actions.githubusercontent.com"
	BearerSchema = "Bearer "
)

func main() {
	if err := start(); err != nil {
		log.Fatal(err)
	}
}

func start() error {
	ctx := context.Background()

	appId := os.Getenv("APP_ID")
	privateKey, err := pathOrContents(os.Getenv("APP_PRIVATE_KEY"))
	if err != nil {
		return err
	}
	appIdInt, err := strconv.ParseInt(appId, 10, 64)
	if err != nil {
		return err
	}

	tr, err := ghinstallation.NewAppsTransport(http.DefaultTransport, appIdInt, []byte(privateKey))
	if err != nil {
		return err
	}

	client := github.NewClient(&http.Client{Transport: tr})

	provider, err := oidc.NewProvider(ctx, Issuer)
	if err != nil {
		return err
	}

	verifier := provider.Verifier(&oidc.Config{SkipClientIDCheck: true})

	e := echo.New()
	e.HideBanner = true
	e.GET("/key", func(c echo.Context) error {
		ctx := c.Request().Context()

		authHeader := c.Request().Header.Get("Authorization")

		if len(authHeader) == 0 || !strings.HasPrefix(authHeader, BearerSchema) {
			return echo.ErrUnauthorized
		}

		idToken, err := verifier.Verify(ctx, authHeader[len(BearerSchema):])
		if err != nil {
			return echo.ErrBadRequest
		}

		claims := &githubClaims{}
		if err := idToken.Claims(&claims); err != nil {
			return echo.ErrBadRequest
		}

		installation, _, err := client.Apps.FindOrganizationInstallation(ctx, claims.RepositoryOwner)
		if err != nil {
			return echo.ErrForbidden
		}

		token, _, err := client.Apps.CreateInstallationToken(ctx, *installation.ID, &github.InstallationTokenOptions{})
		if err != nil {
			return echo.ErrForbidden
		}

		return c.String(http.StatusOK, *token.Token)
	})

	return e.Start(":8080")
}

type githubClaims struct {
	RepositoryOwner string `json:"repository_owner"`
}

func pathOrContents(poc string) (string, error) {
	if len(poc) == 0 {
		return poc, nil
	}

	path := poc
	if path[0] == '~' {
		var err error
		path, err = homedir.Expand(path)
		if err != nil {
			return path, err
		}
	}

	if _, err := os.Stat(path); err == nil {
		contents, err := os.ReadFile(path)
		if err != nil {
			return string(contents), err
		}
		return string(contents), nil
	}

	return poc, nil
}
