// ymcparser project ymcparser.go
package ymcparser

import (
	"fmt"
	"strconv"
	"strings"

	"sync"

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
		Region  string
		Phone   string
		Addres  string
		Persons []Person
	}
	BranchOffList []BranchOffice

	Parser struct {
		collector *colly.Collector
		ctx       *colly.Context
		baseURL   string
		count     int
		wg        sync.WaitGroup
	}
)

func (p *Parser) Start() {

	p.collector = colly.NewCollector()
	p.ctx = colly.NewContext()
	p.baseURL = "https://ymcanyc.org"
	var elems []*colly.HTMLElement
	var OfficeCard BranchOffice
	// Find and visit all links
	p.collector.OnHTML(".location-list-row", func(e *colly.HTMLElement) {
		//e.Request.Visit(e.Attr("href"))
		e.ForEach(".location-list-item", func(i int, e1 *colly.HTMLElement) {
			OfficeCard.Region = e1.ChildText(".field-borough")
			//	fmt.Println("I=", i, "OfficeCard.Region:", OfficeCard.Region)
			OfficeCard.Name = e1.DOM.Find(".location-card-header").Find("h2").Find("span").Text()
			//	fmt.Println("OfficeCard.Name :", OfficeCard.Name)
			OfficeCard.Addres = e1.DOM.Find(".node__content").Find(".field-location-direction").Text()
			//	fmt.Println("OfficeCard.Adress :", OfficeCard.Addres)
			OfficeCard.Phone = e1.DOM.Find(".wrapper-field-location-phone").Find(".field-location-phone").Find("a").Text()
			//	fmt.Println("OfficeCard.Phone :", OfficeCard.Phone)
			k := strconv.Itoa(i)
			p.ctx.Put(k, OfficeCard)
			p.count = i + 1

			elems = append(elems, e1)
		})
		p.wg.Add(p.count)
		for indx, val := range elems {
			go p.ToOfficePage(strconv.Itoa(indx), val)
		}

	})

	p.collector.Visit(p.baseURL + "/locations?type&amenities")
	p.collector.Wait()
	p.wg.Wait()
	for j := 0; j < p.count; j++ {
		o := p.ctx.GetAny(strconv.Itoa(j)).(BranchOffice)
		fmt.Println("********************************************************************************* ")
		fmt.Println("INDEX ", j)
		fmt.Println("OficeData Addres ", o.Addres)
		fmt.Println("OficeData Name ", o.Name)
		fmt.Println("OficeData Phone ", o.Phone)
		fmt.Println("OficeData Region ", o.Region)
		p.ShowPersons(o.Persons)
	}
	fmt.Println("END")

}
func (p *Parser) ShowPersons(ps []Person) {
	for k, v := range ps {
		fmt.Println("       Namber: ", k, " Adress :", v.Name)
		fmt.Println("       Namber: ", k, " Job :", v.Job)
		fmt.Println("       Namber: ", k, " Phone :", v.Phone)
		fmt.Println("       Namber: ", k, " Email :", v.Email)
		fmt.Println("-------------------------------------")

	}

}
func (p *Parser) ToOfficePage(key string, e *colly.HTMLElement) {
	url := p.GetOfficePageUrl(e)
	fmt.Println("URL:", url)
	fmt.Println("KEY:", key)
	c := colly.NewCollector()
	var officedata BranchOffice = p.ctx.GetAny(key).(BranchOffice)
	fmt.Println("OFFICE:", officedata)
	var yes bool = false

	c.OnHTML(".field-sb-body", func(e *colly.HTMLElement) {
		yes = true
		e.ForEach("p", func(i int, e1 *colly.HTMLElement) {
			/*if key == "1" {
				fmt.Println(i)
			}*/

			mail := e1.DOM.Find("a").Text()
			name, job, phone := GetJob(e1, mail)
			officedata.Persons = append(officedata.Persons, Person{
				Name:   name,
				Adress: "",
				Email:  mail,
				Phone:  phone,
				Job:    job,
			})
			p.ctx.Put(key, officedata)

			//			fmt.Println("prsn.Name :", GetJob(e1))

		})

	})
	c.OnHTML(".field-prgf-description", func(e *colly.HTMLElement) {

		if !yes {
			e11 := e.DOM.Find("h2").Text()
			e12 := e.DOM.Find("h3").Text()
			if (e11 == "Leadership Staff") || (e12 == "Leadership") || (e12 == "Staff") || (e12 == "Board of Managers") {

				e.ForEach("p", func(i int, e1 *colly.HTMLElement) {
					/*if key == "1" {
						fmt.Println(i)
					}*/

					mail := e1.DOM.Find("a").Text()
					name, job, phone := GetJob(e1, mail)
					officedata.Persons = append(officedata.Persons, Person{
						Name:   name,
						Adress: "",
						Email:  mail,
						Phone:  phone,
						Job:    job,
					})

					//			fmt.Println("prsn.Name :", GetJob(e1))
					p.ctx.Put(key, officedata)
				})
			}
		}
	})

	c.Visit(url)
	c.Wait()

	fmt.Println("END Person PAGE", officedata)
	fmt.Println("END OFFICE PAGE")
	p.wg.Done()
}

func (p *Parser) GetOfficePageUrl(e *colly.HTMLElement) string {

	return p.baseURL + e.ChildAttr(".btn-primary", "href") + "/about"
}
func GetJob(e *colly.HTMLElement, trim string) (name, job, phone string) {
	s := e.DOM.Text()
	var a []string
	if strings.Contains(s, "\n") {

		a = strings.Split(s, "\n")
		l := len(a)
		if l > 1 {
			a[1] = strings.TrimSuffix(a[1], trim)
			/*fmt.Println("ENDDDD", a)
			fmt.Println("JOBBB", s)*/
		}
		name = a[0]
		job = a[1]
		if l > 2 {
			phone = a[2]
		}
	} else {
		name = s
	}
	return
}
