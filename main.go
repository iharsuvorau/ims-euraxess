package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// offer describes a job offer page of Euraxess: https://euraxess.ec.europa.eu/jobs/421010.
type offer struct {
	title             string
	uri               string
	organization      string
	researchField     string
	researcherProfile string
	deadline          string
	location          string
	typeOfContract    string
	jobStatus         string
	referenceNumber   string
	body              string
	requirements      requirements
}

type requirements struct {
	researchField             string
	yearsOfResearchExperience string
	educationLevel            string
	languages                 string
}

func main() {
	links, err := collectOfferLinks("https://euraxess.ec.europa.eu/jobs/search?keywords=Intelligent%20Materials%20and%20Systems%20Lab")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v", links)

	offers, err := collectOffers(links)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("results: %+v", offers)

	// TODO: render wikitext from offers

	// TODO: create or update pages on the wiki
}

type offerLink struct {
	title string
	uri   string
}

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

	// section#block-system-main > div.view-content > div.views-row > h2 > a
	doc.Find("#block-system-main div.view-content div.views-row").Each(func(i int, s *goquery.Selection) {
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

func collectOffers(links []offerLink) ([]*offer, error) {
	offers := make([]*offer, len(links))
	for i := range links {
		offer, err := collectOffer(links[i]) // TODO: make concurrent
		if err != nil {
			return nil, err
		}
		offers[i] = offer
	}
	return offers, nil
}

func collectOffer(link offerLink) (*offer, error) {
	resp, err := http.Get(link.uri)
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

	body, err := doc.Find(".node-offer-posting .field-body").Html()
	if err != nil {
		return nil, err
	}

	// cleaning up complex markup
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

	ofr := offer{
		title:             link.title,
		uri:               link.uri,
		organization:      strings.TrimSpace(doc.Find(".node-offer-posting ul.list-items .field-company-institute").First().Text()),
		researchField:     researchField,
		researcherProfile: strings.TrimSpace(doc.Find(".node-offer-posting ul.list-items .field-research-profile").First().Text()),
		deadline:          strings.TrimSpace(doc.Find(".node-offer-posting ul.list-items .field-application-deadline").First().Text()),
		location:          strings.TrimSpace(doc.Find(".node-offer-posting ul.list-items .field-country").First().Text()),
		typeOfContract:    strings.TrimSpace(doc.Find(".node-offer-posting ul.list-items .field-type-of-contract").First().Text()),
		jobStatus:         strings.TrimSpace(doc.Find(".node-offer-posting ul.list-items .field-job-status").First().Text()),
		referenceNumber:   strings.TrimSpace(doc.Find(".node-offer-posting ul.list-items .field-reference-number").First().Text()),
		body:              body,
		requirements: requirements{
			researchField:             strings.TrimSpace(doc.Find(".field-required-research-xp .field-research-field").First().Text()),
			yearsOfResearchExperience: strings.TrimSpace(doc.Find(".field-required-research-xp .field-years-of-research").First().Text()),
			educationLevel:            strings.TrimSpace(doc.Find(".field-offer-requirements .field-education-level").First().Text()),
			languages:                 strings.TrimSpace(doc.Find(".field-offer-requirements .field-language-level").First().Text()),
		},
	}

	return &ofr, nil
}
