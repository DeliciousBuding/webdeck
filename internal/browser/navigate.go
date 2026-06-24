package browser

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/chromedp/chromedp"
)

// Navigate enters the cloud game: navigate → dismiss dialogs → wait for video.
func (b *Browser) Navigate() error {
	return chromedp.Run(b.ctx,
		chromedp.Navigate("https://sr.mihoyo.com/cloud/"),
		chromedp.Sleep(3*time.Second),
		chromedp.ActionFunc(func(ctx context.Context) error {
			for i := 0; i < 90; i++ {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}
				var hasVideo bool
				chromedp.Evaluate(`!!document.querySelector("video.game-player__video")`, &hasVideo).Do(ctx)
				if hasVideo {
					log.Printf("[browser] video ready (%ds)", i*2)
					return nil
				}
				chromedp.Evaluate(`document.querySelector('.van-button--danger')?.click()`, nil).Do(ctx)
				chromedp.Evaluate(`document.querySelector('.van-button--default')?.click()`, nil).Do(ctx)
				chromedp.Evaluate(`document.querySelector('.van-overlay')?.click()`, nil).Do(ctx)
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(2 * time.Second):
				}
			}
			return fmt.Errorf("timeout: no game video after 180s")
		}),
	)
}

// DismissHTML clicks away overlay HTML dialogs.
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
