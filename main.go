package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

// Article 表示一篇新闻文章
type Article struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	Site  string `json:"site"`
	Time  string `json:"time"`
}

// Site 表示一个新闻网站的配置
type Site struct {
	Name       string
	URL        string
	TitlePath  string
	LinkPath   string
	TimePath   string
	TimeFormat string
}

func main() {
	// 定义要爬取的网站配置
	sites := []Site{
		{
			Name:      "36氪",
			URL:       "https://tophub.today/n/Q1Vd5Ko85R",
			TitlePath: "td.al a",
			LinkPath:  "td.al a",
			TimePath:  "td:nth-child(3)",
		},
		{
			Name:      "少数派",
			URL:       "https://tophub.today/n/Y2KeDGQdNP",
			TitlePath: "td.al a",
			LinkPath:  "td.al a",
			TimePath:  "td:nth-child(3)",
		},
		{
			Name:      "虎嗅网",
			URL:       "https://tophub.today/n/5VaobgvAj1",
			TitlePath: "td.al a",
			LinkPath:  "td.al a",
			TimePath:  "td:nth-child(3)",
		},
		{
			Name:      "掘金",
			URL:       "https://tophub.today/n/QaqeEaVe9R",
			TitlePath: "td.al a",
			LinkPath:  "td.al a",
			TimePath:  "td:nth-child(3)",
		},
		{
			Name:      "机器之心",
			URL:       "https://tophub.today/n/5VaobgvAj1",
			TitlePath: "td.al a",
			LinkPath:  "td.al a",
			TimePath:  "td:nth-child(3)",
		},
		{
			Name:      "AI新智界",
			URL:       "https://tophub.today/n/EZ7jl0X9kO",
			TitlePath: "td.al a",
			LinkPath:  "td.al a",
			TimePath:  "td:nth-child(3)",
		},
	}

	// 创建收集器
	c := colly.NewCollector(
		colly.AllowedDomains("tophub.today"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
	)

	// 设置重试策略
	setupRetryPolicy(c)

	// 设置限速
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: 2 * time.Second,
	})

	// 存储所有文章
	var articles []Article

	// 为每个网站设置爬取规则
	for _, site := range sites {
		c.OnHTML("table.table tbody tr", func(e *colly.HTMLElement) {
			article := Article{
				Site:  site.Name,
				Title: strings.TrimSpace(e.ChildText(site.TitlePath)),
				URL:   e.ChildAttr(site.LinkPath, "href"),
				Time:  strings.TrimSpace(e.ChildText(site.TimePath)),
			}

			// 只保存有效的文章
			if article.Title != "" && article.URL != "" {
				articles = append(articles, article)
			}
		})

		// 访问网站
		err := c.Visit(site.URL)
		if err != nil {
			log.Printf("Error visiting %s: %v\n", site.Name, err)
			continue
		}
	}

	// 将结果保存为JSON文件
	saveToJSON(articles, "news.json")
}

func saveToJSON(articles []Article, filename string) {
	file, err := json.MarshalIndent(articles, "", "  ")
	if err != nil {
		log.Fatal("Error marshaling to JSON:", err)
	}

	err = os.WriteFile(filename, file, 0644)
	if err != nil {
		log.Fatal("Error writing JSON file:", err)
	}

	fmt.Printf("Successfully saved %d articles to %s\n", len(articles), filename)
}
