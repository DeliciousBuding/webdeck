package browser

import (
	"github.com/chromedp/chromedp"
)

// Navigate loads a URL in the browser. The caller (SRC) decides what URL
// to navigate to — the Gateway does not know about game logic.
func (b *Browser) Navigate(url string) error {
	return chromedp.Run(b.ctx, chromedp.Navigate(url))
}

// DismissHTML evaluates arbitrary JS to dismiss overlays.
// Used by WebUI debug controls and compat API. The JS is
// caller-supplied, not hardcoded to any specific site.
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
