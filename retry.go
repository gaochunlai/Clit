package main

import (
	"time"

	"github.com/gocolly/colly/v2"
)

func setupRetryPolicy(c *colly.Collector) {
	retries := 0
	c.OnError(func(r *colly.Response, err error) {
		if retries < 3 {
			retries++
			time.Sleep(time.Duration(retries) * time.Second)
			r.Request.Retry()
		}
	})
}
