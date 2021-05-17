package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/kelseyhightower/envconfig"
	"github.com/kwyn/bikealert/twilio"
	"github.com/robfig/cron/v3"
)

type Config struct {
	TwilioSID       string `envconfig:"TWILIO_SID"`
	TwilioAuthToken string `envconfig:"TWILIO_AUTH_TOKEN"`
	TwilioNumber    string `envconfig:"TWILIO_NUMBER"`
	RecipientNumber string `envconfig:"RECIPIENT_NUMBER"`
}

type Clock interface {
	Now() time.Time
}

type SystemClock struct{}

func (s *SystemClock) Now() time.Time {
	return time.Now()
}

type Listing struct {
	Title string    `json:"title"`
	Link  string    `json:"link"`
	Date  time.Time `json:"date"`
}

func (l *Listing) String() string {
	return fmt.Sprintf("%v : %v : %v", l.Date, l.Title, l.Link)
}

func Scrape() []*Listing {
	// Request the HTML page.
	res, err := http.Get("https://www.pinkbike.com/buysell/list/?lat=37.7662&lng=-122.1887&distance=100&category=2&price=..2500&framesize=23,27,34,35,36,30,31,47&material=2")
	if err != nil {
		log.Println(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Printf("status code error: %d %s\n", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Println(err)
	}
	listings := make([]*Listing, 0)
	// Find the review items
	doc.Find(".bsitem").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		links := s.Find("a")
		if links.Length() < 2 {
			fmt.Println("Error, only found less than two links")
		}
		title := links.Eq(1)
		href, exists := title.Attr("href")
		if exists {
			l := &Listing{Title: title.Text(), Link: href}
			ScrapeDate(l)
			listings = append(listings, l)

		}
	})
	return listings
}

func ScrapeDate(listing *Listing) error {
	// Request the HTML page.
	res, err := http.Get(listing.Link)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return err
	}

	columns := doc.Find(".buysell-details-column")
	if columns.Length() < 2 {
		return errors.New("found less than two columns")
	}
	c := columns.Eq(1)
	t := c.Text()
	r, err := regexp.Compile(`(\w{3}-\d\d-\d\d\d\d \d?\d:\d\d:\d\d)`)
	if err != nil {
		log.Println(err)
	}
	matches := r.FindStringSubmatch(t)

	if len(matches) >= 2 {
		t, err := parsePinkBikeDate(matches[1])
		if err != nil {
			return fmt.Errorf("could not parse date: %w", err)
		}
		listing.Date = t
	}
	if len(matches) == 1 {
		t, err := parsePinkBikeDate(matches[0])
		if err != nil {
			return fmt.Errorf("could not parse date: %w", err)
		}
		listing.Date = t
	}
	// Do nothing if no matches found.
	return nil
}

func parsePinkBikeDate(s string) (time.Time, error) {
	layout := "Jan-02-2006 15:04:05"
	return time.Parse(layout, s)
}

func filterListings(c Clock, listings []*Listing) []*Listing {
	var filtered []*Listing
	for _, l := range listings {
		if c.Now().Day() == l.Date.Day() && c.Now().Month() == l.Date.Month() {
			filtered = append(filtered, l)
		}
	}
	return filtered
}

func RunScraper() {
	fmt.Println("Read Config")
	var config Config
	err := envconfig.Process("pbscraper", &config)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Setup Twilio")

	var client = &http.Client{
		Timeout: time.Second * 10,
	}
	t := &twilio.Twilio{
		SID:       config.TwilioSID,
		AuthToken: config.TwilioAuthToken,
		Number:    config.TwilioNumber,
		Client:    client,
	}

	fmt.Println("Scrape for bikes")
	listings := Scrape()
	listings = filterListings(&SystemClock{}, listings)
	if len(listings) < 1 {
		fmt.Println("No new bikes")
		return
	}

	message := "New Bikes:\n"
	for _, l := range listings {
		message += l.String() + "\n"
	}
	fmt.Println("Found some bikes\n", message)
	err = t.Send(config.RecipientNumber, "🚵‍♀️ MTN BIKE ALERT 🚵‍♀️:\n"+message)
	if err != nil {
		log.Println(err)
	}
}

func main() {
	location, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		log.Fatal(err)
	}
	c := cron.New(cron.WithLocation(location))
	c.AddFunc("0 7,12,18 * * *", RunScraper)
	c.Run()
}
