package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/logger"
	forcessl "github.com/gobuffalo/mw-forcessl"
	paramlogger "github.com/gobuffalo/mw-paramlogger"
	"github.com/gorilla/sessions"
	"github.com/unrolled/secure"

	"github.com/gobuffalo/buffalo-pop/v2/pop/popmw"
	csrf "github.com/gobuffalo/mw-csrf"
	i18n "github.com/gobuffalo/mw-i18n"
	"github.com/gobuffalo/packr/v2"
	"github.com/navionguy/quotewall/models"
)

// ENV is used to help switch settings based on where the
// application is being run. Default is "development".
var ENV = envy.Get("GO_ENV", "development")
var app *buffalo.App
var T *i18n.Translator

// App is where all routes and middleware for buffalo
// should be defined. This is the nerve center of your
// application.
//
// Routing, middleware, groups, etc... are declared TOP -> DOWN.
// This means if you add a middleware to `app` *after* declaring a
// group, that group will NOT have that new middleware. The same
// is true of resource declarations as well.
//
// It also means that routes are checked in the order they are declared.
// `ServeFiles` is a CATCH-ALL route, so it should always be
// placed last in the route declarations, as it will prevent routes
// declared after it to never be called.
func App() *buffalo.App {
	if app == nil {
		buffaloOptions := buffalo.Options{
			Host:        envy.Get("FORUM_HOST", envy.Get("HOST", "")),
			Env:         ENV,
			SessionName: "_quotewall_session",
			LogLvl:      logger.InfoLevel,
		}
		cookieStore := defaultCookieStore(buffaloOptions)
		buffaloOptions.SessionStore = cookieStore
		app = buffalo.New(buffaloOptions)

		// Automatically redirect to SSL
		app.Use(forceSSL())

		// Log request parameters (filters apply).
		app.Use(paramlogger.ParameterLogger)

		// Protect against CSRF attacks. https://www.owasp.org/index.php/Cross-Site_Request_Forgery_(CSRF)
		// Remove to disable this.
		app.Use(csrf.New)

		// Wraps each request in a transaction.
		//  c.Value("tx").(*pop.Connection)
		// Remove to disable this.
		app.Use(popmw.Transaction(models.DB))

		// Setup and use translations:
		app.Use(translations())

		app.GET("/", HomeHandler)

		cv := &ConversationsResource{}
		app.Resource("/authors", &AuthorsResource{})
		app.GET("/conversations/quickie", cv.QuickieQuote)
		app.GET("/conversations/export/", cv.Export) // this is becoming useless and should probably go away
		app.Resource("/conversations", cv)

		app.ServeFiles("/", assetsBox) // serve files from the public directory
	}

	return app
}

// translations will load locale files, set up the translator `actions.T`,
// and will return a middleware to use to load the correct locale for each
// request.
// for more information: https://gobuffalo.io/en/docs/localization
func translations() buffalo.MiddlewareFunc {
	var err error
	if T, err = i18n.New(packr.New("app:locales", "../locales"), "en-US"); err != nil {
		app.Stop(err)
	}
	return T.Middleware()
}

// forceSSL will return a middleware that will redirect an incoming request
// if it is not HTTPS. "http://example.com" => "https://example.com".
// This middleware does **not** enable SSL. for your application. To do that
// we recommend using a proxy: https://gobuffalo.io/en/docs/proxy
// for more information: https://github.com/unrolled/secure/
func forceSSL() buffalo.MiddlewareFunc {
	return forcessl.Middleware(secure.Options{
		SSLRedirect:     ENV == "production",
		SSLProxyHeaders: map[string]string{"X-Forwarded-Proto": "https"},
	})
}

func defaultCookieStore(opts buffalo.Options) sessions.Store {
	secret := envy.Get("SESSION_SECRET", "")
	if secret == "" && (ENV == "development" || ENV == "test") {
		secret = "buffalo-secret"
	}
	// In production a SESSION_SECRET must be set!
	if secret == "" {
		opts.Logger.Warn("Unless you set SESSION_SECRET env variable, your session storage is not protected!")
	}
	cookieStore := sessions.NewCookieStore([]byte(secret))
	// SameSite field values: strict=3, Lax=2, None=4, Default=1.
	cookieStore.Options.SameSite = 3
	//Cookie secure attributes, see: https://www.owasp.org/index.php/Testing_for_cookies_attributes_(OTG-SESS-002)
	cookieStore.Options.HttpOnly = true
	if ENV == "production" {
		cookieStore.Options.Secure = true
	}
	return cookieStore
}
