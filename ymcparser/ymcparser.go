// ymcparser project ymcparser.go
package ymcparser

import (
	"fmt"

	"github.com/gocolly/colly"
)

const (
	ParseURL = "https://ymcanyc.org/locations?type&amenities"
)

type (
	Person struct {
		Name   string
		Adress string
		Email  string
		Phone  string
		Job    string
	}
	BranchOffice struct {
		Name    string
		Persons []Person
	}
	BranchOffList []BranchOffice
)

func Start() {
	c := colly.NewCollector()

	// Find and visit all links
	c.OnHTML(".location-list-row", func(e *colly.HTMLElement) {
		//e.Request.Visit(e.Attr("href"))
		e.ForEach(".location-list-item", func(i int, e1 *colly.HTMLElement) {
			fmt.Println("I=", i, "ANSW", *e1)

		})
		fmt.Println("END")

	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)

	})

	c.Visit("https://ymcanyc.org/locations?type&amenities")
}
