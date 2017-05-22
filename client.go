package seasnve

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
)

type Client struct {
	c         http.Client
	authToken string
}

func NewClient(username, password string) (Client, error) {
	jar, _ := cookiejar.New(nil)
	c := Client{
		c: http.Client{Jar: jar},
	}

	vals := url.Values{
		"main_0$txtEmail":    {username},
		"main_0$txtPassword": {passoword},
		//"main_0$chkLoginRememberMe": {"on"},
		"main_0$ctl01": {"LOG IND"},
	}
	// Make login request
	resp, err := c.c.Get("https://mit.seas-nve.dk/login/private")
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	// Not successful? Try enhancing with values from hidden input fields
	r, _ := regexp.Compile(`\<input type="hidden" name="[^"]+" id="([^"]+)" value="([^"]+)" />`)
	res := r.FindAllSubmatch(body, -1)
	for _, i := range res {
		// Returns WHOLE_MATCH, ID, VALUE
		vals.Add(string(i[1]), string(i[2]))
	}

	// Make login request
	req, err := http.NewRequest("POST", "https://mit.seas-nve.dk/login/private", strings.NewReader(vals.Encode()))
	if err != nil {
		return c, err
	}
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	//req.Header.Set("Accept-Encoding","gzip, deflate, br")
	req.Header.Set("Accept-Language", "da-DK,en-US;q=0.8,da;q=0.6,en;q=0.4")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("DNT", "1")
	req.Header.Set("Host", "mit.seas-nve.dk")
	req.Header.Set("Origin", "https://mit.seas-nve.dk")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Referer", "https://mit.seas-nve.dk/login/private")
	//Upgrade-Insecure-Requests:1
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36)")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err = c.c.Do(req)
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)

	// Read out Autentication token
	r, _ = regexp.Compile(`var token = '(Basic [^']+)';`)
	authJs := r.FindSubmatch(body)
	c.authToken = string(authJs[1])

	return c, err
}

func (c *Client) do(method, url string, out interface{}) error {
	// Make login request
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-type", "application/json")
	req.Header.Set("Authorization", c.authToken)
	req.Header.Set("User-Agent", "https://github.com/msiebuhr/seas-nve")
	resp, err := c.c.Do(req)
	defer resp.Body.Close()

	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)

	json.Unmarshal(body, out)
	return nil
}

type Metering struct {
	MeteringPoints []struct {
		Address struct {
			City    string
			Door    string
			Floor   string
			GeoPosX string
			GeoPosY string
			Letter  string
			Number  string
			Street  string
			ZipCode string
		}
		ConsumptionYearToDate       float64
		CustomerDwellingInformation struct {
			NumberOfAdults   int
			NumberOfChildren int
		}
		EnergyMark           int64
		IsAmr                bool
		IsInDerogationPeriod bool
		MeterType            string
		MeteringPoint        string
		CustomerNumber       string

		YearlyConsumption []struct {
			Consumption float64
			Year        int64
		}
	}
}

func (c *Client) Metering() (Metering, error) {
	m := Metering{}
	err := c.do("GET", "https://mit.seas-nve.dk/api/v1.0/profile/metering/", &m)
	return m, err
}

type Management struct {
	Address struct {
		City       string
		PostalCode int64
		StreetName string
		Number     int64
		Letter     string
		Floor      string
		Door       string
	}
	FirstName           string
	LastName            string
	PhoneNumber         string
	MobilePhoneNumber   string
	Email               string
	Company             string
	ConcernPermission   bool
	PasswordDefined     bool
	CPRDefined          int // What ENUM? 2=false
	IsPrimaryPerson     bool
	HasActiveAgreements bool
}

func (c *Client) Management() (Management, error) {
	m := Management{}
	err := c.do("GET", "https://mit.seas-nve.dk/api/v1.0/profile/management/", &m)
	return m, err
}

const AGGREGATION_DAY = "day"

type Points struct {
	MeteringPoints []struct {
		MeteringPoint string
		Values []struct {
			Start time.Time
			End time.Time
			Value float64
		}
	}
}

func (c *Client) MeteringPoints(point string, start, end time.Time, aggr string) (Points, error) {
	p := Points{}
	err := c.do("GET", "https://mit.seas-nve.dk/api/v1.0/profile/consumption/?meteringpoints=571313175200099652&start=2017-5-1&end=2017-5-31&aggr=Day", &p)
	return p, err
}
