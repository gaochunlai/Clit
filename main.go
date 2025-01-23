package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"html/template"
	"runtime"

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

// 添加新的函数来检查数据文件
func getTodayDataFile() (string, bool) {
	today := time.Now().Format("2006-01-02")
	filename := fmt.Sprintf("news_%s.json", today)

	// 检查文件是否存在
	if info, err := os.Stat(filename); err == nil {
		// 检查文件是否是今天创建的
		if time.Since(info.ModTime()).Hours() < 24 {
			return filename, true
		}
	}
	return filename, false
}

// 添加新的函数来加载现有数据
func loadExistingData(filename string) (map[string][]Article, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var articles []Article
	if err := json.Unmarshal(data, &articles); err != nil {
		return nil, err
	}

	// 按网站分组整理文章
	newsBysite := make(map[string][]Article)
	for _, article := range articles {
		newsBysite[article.Site] = append(newsBysite[article.Site], article)
	}
	return newsBysite, nil
}

func main() {
	// 获取今天的数据文件名和状态
	filename, exists := getTodayDataFile()
	var newsBysite map[string][]Article

	if exists {
		// 如果今天的数据已存在，直接加载
		fmt.Println("Loading existing data for today...")
		var err error
		newsBysite, err = loadExistingData(filename)
		if err != nil {
			log.Printf("Error loading existing data: %v\n", err)
			// 如果加载失败，继续执行爬取逻辑
		} else {
			// 成功加载数据，直接启动web服务
			startWebServer(newsBysite)
			return
		}
	}

	// 如果没有今天的数据或加载失败，执行爬取逻辑
	fmt.Println("Fetching new data...")

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
	saveToJSON(articles, filename)

	// 按网站分组整理文章
	newsBysite = make(map[string][]Article)
	for _, article := range articles {
		newsBysite[article.Site] = append(newsBysite[article.Site], article)
	}

	// 启动web服务
	startWebServer(newsBysite)
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

// 打开默认浏览器
func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Printf("Error opening browser: %v\n", err)
	}
}

// 将web服务器启动逻辑抽取为单独的函数
func startWebServer(newsBysite map[string][]Article) {
	// 启动Web服务器
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("templates/news.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, newsBysite)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// 自动打开默认浏览器
	go func() {
		time.Sleep(100 * time.Millisecond) // 等待服务器启动
		openBrowser("http://localhost:8080")
	}()

	fmt.Println("Starting server at http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
