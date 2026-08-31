package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nhassl3/IpBuild-backend/pkg/logger"
)

const (
	originHeader        = "Origin"
	secFetchSiteHeader  = "Sec-Fetch-Site"
	requestedWithHeader = "X-Requested-With"

	// RequestedWithValue is the value the SPA must send on cookie-driven
	// endpoints. Any non-safelisted header forces a CORS preflight, which the
	// browser refuses for an origin outside AllowOrigins — so a forged request
	// never reaches the handler. A cross-site <form> cannot set it at all.
	RequestedWithValue = "fetch"

	crossSiteCode  = "CROSS_SITE_BLOCKED"
	crossSiteValue = "cross-site request blocked"
)

// CrossSiteGuard rejects state-changing requests that a browser initiated from
// an origin outside allowOrigins.
//
// The refresh cookie is SameSite=None in production, so the browser attaches it
// to requests from any site. CORS keeps the attacker from reading the response,
// but the request itself still executes: because RefreshToken rotates and
// blacklists the old JTI, a drive-by POST from evil.com can revoke a victim's
// refresh token server-side while the replacement Set-Cookie is dropped by the
// browser's third-party cookie rules — locking the victim out of their session.
//
// Requests carrying neither header (curl, server-to-server) pass through: they
// cannot carry a victim's cookie, so there is nothing to forge.
func CrossSiteGuard(allowOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowOrigins))
	for _, o := range allowOrigins {
		if o = strings.TrimSpace(o); o != "" {
			allowed[strings.ToLower(strings.TrimSuffix(o, "/"))] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		if origin := c.GetHeader(originHeader); origin != "" {
			if _, ok := allowed[strings.ToLower(strings.TrimSuffix(origin, "/"))]; !ok {
				logger.From(c.Request.Context()).Warn("csrf: origin not allowed",
					logger.String("origin", origin))
				newErrorResponseWithCode(c, http.StatusForbidden, crossSiteCode, crossSiteValue)
				return
			}
			c.Next()
			return
		}

		// No Origin: fall back to Fetch Metadata. Absent header means an old
		// browser or a non-browser client — neither carries the victim's cookie
		// in a way the attacker controls.
		if site := c.GetHeader(secFetchSiteHeader); site == "cross-site" {
			logger.From(c.Request.Context()).Warn("csrf: cross-site fetch metadata",
				logger.String("sec_fetch_site", site))
			newErrorResponseWithCode(c, http.StatusForbidden, crossSiteCode, crossSiteValue)
			return
		}

		c.Next()
	}
}

// RequireRequestedWith enforces the X-Requested-With header on endpoints that
// authenticate purely by cookie. See RequestedWithValue for why this blocks
// forged requests before they reach the handler.
func RequireRequestedWith(c *gin.Context) {
	if c.GetHeader(requestedWithHeader) != RequestedWithValue {
		logger.From(c.Request.Context()).Warn("csrf: missing X-Requested-With header")
		newErrorResponseWithCode(c, http.StatusForbidden, crossSiteCode, crossSiteValue)
		return
	}
	c.Next()
}
