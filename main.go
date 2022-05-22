package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	ics "github.com/arran4/golang-ical"
	"github.com/sclevine/agouti"
)

// Configs defines inputs to this app
type Configs struct {
	Season    string   `json:"season"`
	Titles    []string `json:"titles"`
	URLWinter string   `json:"url_winter"`
	URLSpring string   `json:"url_spring"`
	URLSummer string   `json:"url_summer"`
	URLFall   string   `json:"url_fall"`
}

// AnimeInfo defines information of an anime
type AnimeInfo struct {
	Year       int
	StartMonth time.Month
	Hour       int
	Minute     int
	Week       time.Weekday
}

const debug = false

const timezone = "Asia/Tokyo"

func setStartAtLocalTime(event *ics.VEvent, t time.Time, props ...ics.PropertyParameter) {
	event.SetProperty(ics.ComponentPropertyDtStart, t.Format("20060102T150405"), props...)
}
func setEndAtLocalTime(event *ics.VEvent, t time.Time, props ...ics.PropertyParameter) {
	event.SetProperty(ics.ComponentPropertyDtEnd, t.Format("20060102T150405"), props...)
}

func downloadAnimeInfo(URL string) (io.Reader, error) {
	chromeArgs := agouti.ChromeOptions(
		"args", []string{
			"--headless",
			"--disable-gpu",
		})
	chromeExcludeSwitches := agouti.ChromeOptions(
		"excludeSwitches", []string{
			"enable-logging",
		})

	driver := agouti.ChromeDriver(chromeArgs, chromeExcludeSwitches)
	defer driver.Stop()
	if err := driver.Start(); err != nil {
		return nil, err
	}

	page, err := driver.NewPage()
	if err != nil {
		return nil, err
	}
	if err := page.Session().SetImplicitWait(3000); err != nil {
		return nil, err
	}
	if err := page.Session().SetPageLoad(3000); err != nil {
		return nil, err
	}
	fmt.Println("Accessing to the page...")
	if err := page.Navigate(URL); err != nil {
		return nil, err
	}
	html, err := page.HTML()
	if err != nil {
		return nil, err
	}
	// fmt.Println(html)
	return bytes.NewReader([]byte(html)), nil
}

func loadHTMLFromFile(cacheFilePath string) (io.Reader, error) {
	content, err := ioutil.ReadFile(cacheFilePath)
	return bytes.NewReader(content), err
}

// https://stackoverflow.com/a/38537764
func substring(s string, start int, end int) string {
	startStrIdxgo := 0
	i := 0
	for j := range s {
		if i == start {
			startStrIdxgo = j
		}
		if i == end {
			return s[startStrIdxgo:j]
		}
		i++
	}
	return s[startStrIdxgo:]
}

func parseSeason(s string) (time.Month, error) {
	switch s {
	case "冬":
		return time.January, nil
	case "春":
		return time.April, nil
	case "夏":
		return time.July, nil
	case "秋":
		return time.October, nil
	default:
		return 0, fmt.Errorf("invalid input to stringToCour(): %s", s)
	}
}

func parseWeekday(s string) (time.Weekday, error) {
	switch s {
	case "日":
		return time.Sunday, nil
	case "月":
		return time.Monday, nil
	case "火":
		return time.Tuesday, nil
	case "水":
		return time.Wednesday, nil
	case "木":
		return time.Thursday, nil
	case "金":
		return time.Friday, nil
	case "土":
		return time.Saturday, nil
	default:
		return 0, fmt.Errorf("invalid input to weekStringToInt(): %s", s)
	}
}

func scrape(html io.Reader) (map[string]AnimeInfo, error) {
	doc, err := goquery.NewDocumentFromReader(html)
	if err != nil {
		return nil, err
	}
	pageTitle := doc.Find("title").First().Text() // should be like 2021冬アニメ配信ラインナップ | dアニメストア
	year, err := strconv.Atoi(substring(pageTitle, 0, 4))
	if err != nil {
		log.Printf("page title: %s", pageTitle)
		return nil, err
	}
	cour, err := parseSeason(substring(pageTitle, 4, 5))
	if err != nil {
		return nil, err
	}
	animeByWeek := doc.Find("div.weekWrapper")
	animes := make(map[string]AnimeInfo)
	errorOccurred := false
	var errLoop error
	animeByWeek.Each(func(idx int, s *goquery.Selection) {
		if errorOccurred {
			return
		}
		weekdayText := substring(s.Find("div.weekText").First().Text(), 0, 1)
		if weekdayText == "そ" {
			return
		}
		weekday, err := parseWeekday(weekdayText)
		if err != nil {
			errorOccurred = true
			errLoop = err
			return
		}
		// fmt.Println(weekday)
		s.Find("div.itemModule.list").Each(func(idx int, s2 *goquery.Selection) {
			if errorOccurred {
				return
			}

			startTime := s2.Find("div.workMainText").First().Text() // like 25:00
			// 放送前だと「4月13日 24:00～ <br> TITLE」 のようになっているので一番最後のcontentをtitleとみなす
			// アニメタイトルに改行は入っていないと想定
			title := s2.Find("div.textContainerIn span").Contents().Last().Text()
			hour, err := strconv.Atoi(startTime[:2])
			if err != nil {
				errorOccurred = true
				errLoop = err
				return
			}

			weekdayTmp := weekday
			if hour >= 24 {
				weekdayTmp++
				weekdayTmp %= 7
				hour -= 24
			}
			minute, _ := strconv.Atoi(startTime[3:5])
			animes[title] = AnimeInfo{year, cour, hour, minute, weekdayTmp}
		})
	})
	if errorOccurred {
		return nil, errLoop
	}
	return animes, nil
}

func generateWeekdayToStartDateMap(year int, month time.Month) map[time.Weekday]time.Time {
	m := make(map[time.Weekday]time.Time)
	for i := 0; i < 7; i++ {
		t := time.Date(year, month, 1+i, 0, 0, 0, 0, time.Local)
		m[t.Weekday()] = t
	}
	return m
}

func generateICAL(animes map[string]AnimeInfo, titles []string) (string, error) {
	var weekdayToTime map[time.Weekday]time.Time

	//すべてのアニメで年とクールが同じと仮定
	for _, anime := range animes {
		weekdayToTime = generateWeekdayToStartDateMap(anime.Year, anime.StartMonth)
		break
	}

	cal := ics.NewCalendar()
	cal.SetMethod(ics.MethodRequest)

	for _, title := range titles {
		info, ok := animes[title]
		if !ok {
			return "", fmt.Errorf("%s not found in the scraped data", title)
		}
		t := weekdayToTime[info.Week]
		startDate := time.Date(t.Year(), t.Month(), t.Day(), info.Hour, info.Minute, 0, t.Nanosecond(), t.Location())
		event := cal.AddEvent(title)
		tz := &ics.KeyValues{Key: string(ics.PropertyTzid), Value: []string{timezone}}
		setStartAtLocalTime(event, startDate, tz)
		setEndAtLocalTime(event, startDate.Add(30*time.Minute), tz)
		event.SetSummary(title)
		// 1クール13話を仮定
		event.SetProperty(ics.ComponentProperty(ics.PropertyRrule), "FREQ=WEEKLY;COUNT=13")
	}
	return cal.Serialize(), nil
}

func readConfigs(path string) (*Configs, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var configs Configs
	if err := json.Unmarshal(content, &configs); err != nil {
		return nil, err
	}
	return &configs, nil
}

func main() {
	var html io.Reader
	var err error
	// const animeURL = "https://anime.dmkt-sp.jp/animestore/CF/winter"
	// titles := []string{"ウマ娘 プリティーダービー Season 2", "はたらく細胞BLACK", "のんのんびより のんすとっぷ"}

	var configFilePath string
	flag.Parse()
	if len(flag.Args()) == 1 {
		configFilePath = flag.Args()[0]
	} else {
		configFilePath = "configs.json"
	}
	configs, err := readConfigs(configFilePath)
	if err != nil {
		log.Fatalln(err)
	}

	var url string
	if !debug {
		switch strings.ToLower(configs.Season) {
		case "winter":
			url = configs.URLWinter
		case "spring":
			url = configs.URLSpring
		case "summer":
			url = configs.URLSummer
		case "fall":
			url = configs.URLFall
		default:
			log.Fatalf("invalid season specified: %s\n", configs.Season)
		}
		html, err = downloadAnimeInfo(url)
		// data, _ := ioutil.ReadAll(html)
		// ioutil.WriteFile("cache.html", data, os.ModePerm)
	} else {
		html, err = loadHTMLFromFile("cache.html")
	}
	if err != nil {
		log.Fatalln(err)
	}
	animes, err := scrape(html)
	if err != nil {
		log.Println("Failed to scraping the page")
		log.Fatalln(err)
	}
	if debug {
		fmt.Println(animes)
	}
	if len(animes) == 0 {
		log.Fatalln("anime info does not exists. please check url or try again.")
	}
	ical, err := generateICAL(animes, configs.Titles)
	if err != nil {
		log.Println("Failed to generating a ical file")
		log.Fatalln(err)
	}
	fmt.Println(ical)
	ioutil.WriteFile("anime.ics", []byte(ical), os.ModePerm)
}
