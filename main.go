// Command simple is a chromedp example demonstrating how to do a simple google
// search.
package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp/runner"
	"github.com/pkg/errors"

	"github.com/chromedp/chromedp"
)

func main() {
	var err error
	url := "https://www.google.com/"
	// create context
	ctxt, cancel := context.WithCancel(context.Background())
	defer cancel()
	// create chrome instance
	c, err := chromedp.New(ctxt,
		chromedp.WithLog(log.Printf),
		chromedp.WithRunnerOptions(runner.WindowSize(412, 800),
			runner.UserAgent(""),
			runner.Flag("headless", true)),
	)
	if err != nil {
		log.Fatal(err)
	}
	rInfo := ChromeRunnerInfo{
		Url:    url,
		SaveTo: "111.jpeg",
	}
	err = c.Run(ctxt, chromedp.Tasks{
		navigate(&rInfo),
		chromedp.Sleep(3 * time.Second),
		getBaseUrl(&rInfo),
		setLayoutMetrics(),
		captureScreenshot(&rInfo),
	})

	fmt.Printf("%+v", rInfo)
	// shutdown chrome
	err = c.Shutdown(ctxt)
	if err != nil {
		log.Fatal(err)
	}

	// wait for chrome to finish
	err = c.Wait()
	if err != nil {
		log.Fatal(err)
	}

	// err = ioutil.WriteFile("contact-form.png", buf, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

type ChromeRunnerInfo struct {
	Url      string
	SaveTo   string
	Redirect string
	ErrTxt   string
}

func navigate(rInfo *ChromeRunnerInfo) chromedp.ActionFunc {
	return func(ctxt context.Context, h cdp.Executor) error {
		th, ok := h.(*chromedp.TargetHandler)
		if !ok {
			return chromedp.ErrInvalidHandler
		}
		frameID, _, errTxt, err := page.Navigate(rInfo.Url).Do(ctxt, th)
		if errTxt != "" { // ERR_CONNECTION_TIMED_OUT   ERR_NAME_NOT_RESOLVED
			rInfo.ErrTxt = errTxt
			return errors.New(errTxt)
		}
		if err != nil {
			return err
		}

		return th.SetActive(ctxt, frameID)
	}
}

func getBaseUrl(rInfo *ChromeRunnerInfo) chromedp.ActionFunc {
	return func(ctxt context.Context, h cdp.Executor) error {
		_, ok := h.(*chromedp.TargetHandler)
		if !ok {
			return chromedp.ErrInvalidHandler
		}
		root, err := dom.GetDocument().Do(ctxt, h)
		if err != nil {
			return err
		}
		rInfo.Redirect = root.BaseURL
		return nil
	}
}

func setLayoutMetrics() chromedp.ActionFunc {
	return func(i context.Context, executor cdp.Executor) error {
		_, _, contentSize, err := page.GetLayoutMetrics().Do(i, executor)
		if err != nil {
			return err
		}
		w := int64(contentSize.Width)
		h := int64(contentSize.Height)
		scale := 1.0
		err = emulation.SetDeviceMetricsOverride(w, h, scale, false).WithScale(scale).Do(i, executor)
		if err != nil {
			return err
		}

		return nil
	}
}

func captureScreenshot(rInfo *ChromeRunnerInfo) chromedp.ActionFunc {
	return func(i context.Context, executor cdp.Executor) error {
		data, err := page.CaptureScreenshot().WithFormat(page.CaptureScreenshotFormatJpeg).WithQuality(90).Do(i, executor)
		if err != nil {
			return err
		}
		ioutil.WriteFile(rInfo.SaveTo, data, 0644)
		return nil
	}
}
