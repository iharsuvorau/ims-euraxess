package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"bitbucket.org/iharsuvorau/mediawiki"
	"github.com/PuerkitoBio/goquery"
)

// Offer describes a job offer page of Euraxess: https://euraxess.ec.europa.eu/jobs/421010.
type Offer struct {
	Title             string
	URI               string
	Organization      string
	ResearchField     string
	ResearcherProfile string
	Deadline          string
	Location          string
	TypeOfContract    string
	HoursPerWeek      string
	JobStatus         string
	ReferenceNumber   string
	Body              template.HTML
	Requirements      Requirements
}

// Requirements describes offer requirements.
type Requirements struct {
	ResearchField             string
	YearsOfResearchExperience string
	EducationLevel            string
	Languages                 string
}

// TODO: a job has an expiration date after which it must be removed from the wiki
// TODO: fix templating issue

func main() {
	exuri := flag.String("uri", "", "euraxess URI to parse")
	mwuri := flag.String("mwuri", "localhost/mediawiki", "mediawiki URI")
	page := flag.String("page", "Job Offers", "page title to update with new offers")
	section := flag.String("section", "Euraxess Offers", "section title on the page to create or update with new offers")
	keyword := flag.String("keyword", "IMS lab", "keyword to look for in Organization field at the Euraxess website")
	name := flag.String("name", "", "login name of the bot for updating pages")
	pass := flag.String("pass", "", "login password of the bot for updating pages")
	offersTmpl := flag.String("tmpl", "offers.tmpl", "template for the offers list")
	logPath := flag.String("log", "euraxess.log", "specify the filepath for a log file, if it's empty all messages are logged into stdout")
	flag.Parse()
	if len(*exuri) == 0 || len(*mwuri) == 0 || len(*name) == 0 || len(*pass) == 0 || len(*section) == 0 || len(*offersTmpl) == 0 {
		log.Fatal("all flags are compulsory, use -h to see the documentation")
	}

	var logger *log.Logger
	if len(*logPath) > 0 {
		f, err := os.Create(*logPath)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		logger = log.New(f, "", log.LstdFlags)
	} else {
		logger = log.New(os.Stdout, "", log.LstdFlags)
	}

	links, err := collectOfferLinks(*exuri)
	if err != nil {
		log.Fatal(err)
	}
	if len(links) == 0 {
		return
	}
	logger.Printf("%v links have been collected", len(links))

	offers := collectOffers(links, logger)
	logger.Printf("%v offers have been collected", len(offers))

	// filter only IMS offers
	imsOffers := []*Offer{}
	for _, v := range offers {
		if strings.Contains(strings.ToLower(v.Organization), strings.ToLower(*keyword)) {
			imsOffers = append(imsOffers, v)
		}
	}

	markup, err := renderOffers(imsOffers, *offersTmpl)
	if err != nil {
		log.Fatal(err)
	}

	// According to https://www.mediawiki.org/wiki/API:Edit there are following contentmodels
	// available: MassMessageListContent, flow-board, Scribunto, JsonSchema, NewsletterContent,
	// wikitext, javascript, json, css, tex
	var contentModel = "wikitext"

	_, err = mediawiki.UpdatePage(*mwuri, *page, markup, contentModel, *name, *pass, *section)
	if err != nil {
		log.Fatal(err)
	}
	logger.Println("wiki page has been updated")
}

type offerLink struct {
	title string
	uri   string
}

// collectOfferLinks collects information from a search page provided via a flag.
func collectOfferLinks(path string) ([]offerLink, error) {
	resourceURI, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code is not 200")
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	links := []offerLink{}

	contentRows := doc.Find("#block-system-main div.view-content div.views-row")

	// no offers found
	if contentRows.Length() == 0 {
		return links, nil
	}

	// collecting found offers
	contentRows.Each(func(i int, s *goquery.Selection) {
		item := s.Find("h2 a")
		title := item.Text()
		href, ok := item.Attr("href")

		if len(title) > 0 && ok {
			uri := resourceURI.ResolveReference(&url.URL{Path: href})
			links = append(links, offerLink{
				title: title,
				uri:   uri.String(),
			})
		}
	})

	return links, nil
}

func collectOffersSequential(links []offerLink, logger *log.Logger) ([]*Offer, error) {
	offers := make([]*Offer, len(links))
	for i := range links {
		offer, err := collectOffer(links[i], logger)
		if err != nil {
			return nil, err
		}
		offers[i] = offer
	}

	return offers, nil
}

// collectOffers downloads and parses offers concurrently and logs any errors.
func collectOffers(links []offerLink, logger *log.Logger) []*Offer {
	offers := []*Offer{}
	var wg sync.WaitGroup
	for _, link := range links {
		wg.Add(1)
		go func(link offerLink) {
			defer func() {
				wg.Done()
				logger.Println("wg done called")
			}()
			logger.Printf("getting %v", link.title)
			offer, err := collectOffer(link, logger)
			if err != nil {
				logger.Printf("failed to collect an offer: %v", err)
			} else {
				offers = append(offers, offer)
			}
		}(link)
	}
	wg.Wait()

	return offers
}

func collectOffer(link offerLink, logger *log.Logger) (*Offer, error) {
	logger.Printf("downloading %v", link.uri)

	resp, err := http.Get(link.uri)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("request isn't successful for %v: %v", link, err)
	}
	defer resp.Body.Close()

	logger.Printf("parsing %v", link.uri)
	defer logger.Printf("done with %v", link.uri)

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	// replacing body links with wikitext markup
	body, err := doc.Find(".node-offer-posting .field-body").Html()
	if err != nil {
		return nil, err
	}
	bodyLinks := doc.Find(".node-offer-posting .field-body a")
	bodyLinksContent := make([][]string, bodyLinks.Length())
	bodyLinks.Each(func(i int, link *goquery.Selection) {
		html, err := goquery.OuterHtml(link)
		if err != nil {
			return
		}
		href, _ := link.Attr("href")
		text := link.Text()
		bodyLinksContent[i] = []string{html, href, text}
	})
	for _, link := range bodyLinksContent {
		newLink := fmt.Sprintf("[%s %s]", link[1], link[2])
		body = strings.ReplaceAll(body, link[0], newLink)
	}

	// cleaning up complex markup of one field
	researchField := strings.TrimSpace(doc.Find(".node-offer-posting ul.list-items .field-research-field").First().Text())
	{
		a := strings.Split(researchField, "\n")
		ar := []string{}
		for _, item := range a {
			if s := strings.TrimSpace(item); len(s) > 0 {
				ar = append(ar, s)
			}
		}
		researchField = strings.Join(ar, ", ")
	}

	ofr := Offer{
		Title:             link.title,
		URI:               link.uri,
		Organization:      strings.TrimSpace(doc.Find(".node-offer-posting ul.list-items .field-company-institute").First().Text()),
		ResearchField:     researchField,
		ResearcherProfile: strings.TrimSpace(doc.Find(".node-offer-posting ul.list-items .field-research-profile").First().Text()),
		Deadline:          strings.TrimSpace(doc.Find(".node-offer-posting ul.list-items .field-application-deadline").First().Text()),
		Location:          strings.TrimSpace(doc.Find(".node-offer-posting ul.list-items .field-country").First().Text()),
		TypeOfContract:    strings.TrimSpace(doc.Find(".node-offer-posting ul.list-items .field-type-of-contract").First().Text()),
		HoursPerWeek:      strings.TrimSpace(doc.Find(".node-offer-posting ul.list-items .field-hours-per-week").First().Text()),
		JobStatus:         strings.TrimSpace(doc.Find(".node-offer-posting ul.list-items .field-job-status").First().Text()),
		ReferenceNumber:   strings.TrimSpace(doc.Find(".node-offer-posting ul.list-items .field-reference-number").First().Text()),
		Body:              template.HTML(body),
		Requirements: Requirements{
			ResearchField:             strings.TrimSpace(doc.Find(".field-required-research-xp .field-research-field").First().Text()),
			YearsOfResearchExperience: strings.TrimSpace(doc.Find(".field-required-research-xp .field-years-of-research").First().Text()),
			EducationLevel:            strings.TrimSpace(doc.Find(".field-offer-requirements .field-education-level").First().Text()),
			Languages:                 strings.TrimSpace(doc.Find(".field-offer-requirements .field-language-level").First().Text()),
		},
	}

	return &ofr, nil
}

// renderOffers produces HTML markup provided templates and data.
func renderOffers(offers []*Offer, paths ...string) (string, error) {
	var tmpl = template.Must(template.ParseFiles(paths...))
	var out bytes.Buffer
	err := tmpl.Execute(&out, offers)
	return out.String(), err
}
