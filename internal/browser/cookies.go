package browser

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type CookieFile struct {
	Cookies []Cookie `json:"cookies"`
}
type Cookie struct {
	Name     string  `json:"name"`
	Value    string  `json:"value"`
	Domain   string  `json:"domain"`
	Path     string  `json:"path"`
	Expires  float64 `json:"expires"`
	HTTPOnly bool    `json:"httpOnly"`
	Secure   bool    `json:"secure"`
	SameSite string  `json:"sameSite"`
}

func (b *Browser) loadCookies(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("[cookies] read: %v", err)
		return
	}
	var cf CookieFile
	if err := json.Unmarshal(data, &cf); err != nil {
		log.Printf("[cookies] parse: %v", err)
		return
	}
	for _, ck := range cf.Cookies {
		expr := cdp.TimeSinceEpoch(time.Unix(int64(ck.Expires), 0))
		sameSite := network.CookieSameSiteLax
		switch ck.SameSite {
		case "Strict":
			sameSite = network.CookieSameSiteStrict
		case "None":
			sameSite = network.CookieSameSiteNone
		}
		chromedp.Run(b.ctx, chromedp.ActionFunc(func(ctx context.Context) error {
			return network.SetCookie(ck.Name, ck.Value).
				WithDomain(ck.Domain).WithPath(ck.Path).
				WithSecure(ck.Secure).WithHTTPOnly(ck.HTTPOnly).
				WithSameSite(sameSite).WithExpires(&expr).
				Do(ctx)
		}))
	}
	log.Printf("[cookies] loaded %d", len(cf.Cookies))
}
