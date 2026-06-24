package browser

import (
	
	"context"
	"log/slog"
	"time"

	"github.com/chromedp/chromedp"
)

// Navigate loads a URL and waits for page stability.
func (b *Browser) Navigate(url string) error {
	slog.Info("navigating", "url", url)
	return chromedp.Run(b.ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Sleep(3*time.Second),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var title string
			chromedp.Title(&title).Do(ctx)
			slog.Info("page ready", "title", title)
			return nil
		}),
	)
}

// DismissHTML evaluates arbitrary JS to dismiss overlays.
func (b *Browser) DismissHTML() {
	chromedp.Run(b.ctx,
		chromedp.Evaluate(`
			document.querySelector('.van-button--default')?.click();
			document.querySelector('.van-button--danger')?.click();
			document.querySelector('.van-overlay')?.click();
		`, nil),
	)
}

// InspectWhite returns DOM elements with non-transparent backgrounds.
// Debug tool — not used in production path.
func (b *Browser) InspectWhite() string {
	var result string
	chromedp.Run(b.ctx, chromedp.Evaluate(`(function(){
		var white=[];
		document.querySelectorAll('*').forEach(function(el){
			var s=window.getComputedStyle(el);
			var bg=s.backgroundColor;
			if(bg&&bg!=='rgba(0, 0, 0, 0)'&&bg!=='rgb(0, 0, 0)'&&bg!=='transparent'){
				var r=el.getBoundingClientRect();
				if(r.width>100&&r.height>50){
					white.push({tag:el.tagName,cls:el.className.substring(0,80),bg:bg,w:Math.round(r.width),h:Math.round(r.height),top:Math.round(r.top)});
				}
			}
		});
		return JSON.stringify(white);
	})()`, &result))
	return result
}
