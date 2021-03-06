// ymcparser project ymcparser.go
package ymcparser

import (
	"fmt"
	"strconv"
	"strings"

	"encoding/json"

	"io/ioutil"
	"log"
	"net/http"

	"os"
	"sync"

	"github.com/gocolly/colly"
)

const (
	ParseURL    = "https://ymcanyc.org/locations?type&amenities"
	GeoLocates  = "https://maps.googleapis.com/maps/api/geocode/json?address="
	ApiKeyPrefx = "+CA&key="
)

type (
	Person struct {
		Name   string
		Adress string
		Email  string
		Phone  string
		Job    string
	}
	RemoteOffice struct {
		Name       string
		Region     string
		Phone      string
		Addres     string
		Persons    []Person
		Latidude   string
		Longtitude string
	}
	RemoteOffList []RemoteOffice

	Parser struct {
		collector    *colly.Collector
		ctx          *colly.Context
		baseURL      string
		count        int
		wgGetPersons sync.WaitGroup
		apikey       string
		errlog       *log.Logger
	}

	GeoLocData struct {
		Lat  float64 `json:"lat"`
		Long float64 `json:"lng"`
	}
	Geometry struct {
		GeoPosition GeoLocData `json:"location"`
	}

	Geodata struct {
		Geo Geometry `json:"geometry"`
	}
	RespResult struct {
		Result []Geodata `json:"results"`
	}
)

func (p *Parser) Run(apikey string) {
	p.apikey = apikey
	p.collector = colly.NewCollector()
	p.ctx = colly.NewContext()

	p.baseURL = "https://ymcanyc.org"
	var elems []*colly.HTMLElement

	p.errlog = log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	p.collector.OnHTML(".location-list-row", func(e *colly.HTMLElement) {

		e.ForEach(".location-list-item", func(i int, e1 *colly.HTMLElement) {
			elems = append(elems, e1)
		})

		p.count = len(elems)

		p.wgGetPersons.Add(p.count)

		for indx := 0; indx < p.count; indx++ {
			go p.ToRemoteOfficesHTMLPages(strconv.Itoa(indx), elems[indx])
		}

	})

	p.collector.Visit(p.baseURL + "/locations?type&amenities")
	p.collector.Wait()
	p.wgGetPersons.Wait()

	for j := 0; j < p.count; j++ {
		o := p.ctx.GetAny(strconv.Itoa(j)).(RemoteOffice)
		fmt.Println("********************************************************************************* ")
		fmt.Println("INDEX ", j)
		fmt.Println("OficeData Addres ", o.Addres)
		fmt.Println("OficeData Name ", o.Name)
		fmt.Println("OficeData Phone ", o.Phone)
		fmt.Println("OficeData Region ", o.Region)
		fmt.Println("OficeData Longtitude ", o.Longtitude)
		fmt.Println("OficeData Latitude ", o.Latidude)
		p.ShowPersons(o.Persons)
	}
	fmt.Println("END")

}
func (p *Parser) GetOfficeData(e *colly.HTMLElement) (ofdtata RemoteOffice) {

	ofdtata.Region = e.ChildText(".field-borough")
	ofdtata.Name = e.DOM.Find(".location-card-header").Find("h2").Find("span").Text()
	ofdtata.Addres = e.DOM.Find(".node__content").Find(".field-location-direction").Text()
	ofdtata.Phone = e.DOM.Find(".wrapper-field-location-phone").Find(".field-location-phone").Find("a").Text()

	return
}
func (p *Parser) ToRemoteOfficesHTMLPages(key string, e *colly.HTMLElement) {

	var err error

	officedata := p.GetOfficeData(e)
	officedata.Latidude, officedata.Longtitude, err = p.GetGeoLocByAdress(officedata.Addres)
	if err != nil {
		p.errlog.Println(err)
	}
	url := p.GetOfficePageUrl(e)

	c := colly.NewCollector()

	var yes bool = false

	c.OnHTML(".field-sb-body", func(e *colly.HTMLElement) {
		yes = true
		e.ForEach("p", func(i int, e1 *colly.HTMLElement) {

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

		})

	})
	c.OnHTML(".field-prgf-description", func(e *colly.HTMLElement) {

		if !yes {
			e11 := e.DOM.Find("h2").Text()
			e12 := e.DOM.Find("h3").Text()
			if (e11 == "Leadership Staff") || (e12 == "Leadership") || (e12 == "Staff") || (e12 == "Board of Managers") {

				e.ForEach("p", func(i int, e1 *colly.HTMLElement) {

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
				})
			}
		}
	})

	c.Visit(url)
	c.Wait()

	/*fmt.Println("END Person PAGE", officedata)
	fmt.Println("END OFFICE PAGE")*/
	p.wgGetPersons.Done()

}
func (p *Parser) GetGeoLocByAdress(addres string) (lat, long string, err error) {

	client := &http.Client{}
	addr := strings.ReplaceAll(addres, " ", "%20")
	rq := GeoLocates + addr + ApiKeyPrefx + p.apikey

	resp, err := client.Get(rq)
	if err != nil {

		return "", "", err
	}

	defer resp.Body.Close()
	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return "", "", err
	}

	var r RespResult
	if err = json.Unmarshal(body, &r); err != nil {
		return "", "", err
	}

	lat = fmt.Sprintf("%.7f", r.Result[0].Geo.GeoPosition.Lat)
	long = fmt.Sprintf("%.7f", r.Result[0].Geo.GeoPosition.Long)

	return
}

//---- out ----}

func (p *Parser) ShowPersons(ps []Person) {
	for k, v := range ps {
		fmt.Println("       Namber: ", k, " Name :", v.Name)
		fmt.Println("       Namber: ", k, " Job :", v.Job)
		fmt.Println("       Namber: ", k, " Phone :", v.Phone)
		fmt.Println("       Namber: ", k, " Email :", v.Email)

		fmt.Println("-------------------------------------")

	}

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
