package seasnve

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type Client struct {
	c         http.Client
	authToken string
}

func NewClient() Client {
	jar, _ := cookiejar.New(nil)
	c := Client{
		c: http.Client{Jar: jar},
	}

	return c
}

func (c *Client) Login(username, password string) error {
	vals := url.Values{
		"main_0$txtEmail":    {username},
		"main_0$txtPassword": {password},
		//"main_0$chkLoginRememberMe": {"on"},
		"main_0$ctl01": {"LOG IND"},
	}
	// Make login request
	resp, err := c.c.Get("https://mit.seas-nve.dk/login/private")
	if err != nil {
		return err
	}
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
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err = c.c.Do(req)
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)

	// Read out Autentication token
	r, _ = regexp.Compile(`var token = '(Basic [^']+)';`)
	authJs := r.FindSubmatch(body)

	if len(authJs) != 2 { // Match and extraction
		return errors.New("Not authorized. Wrong username or password")
	}

	c.authToken = string(authJs[1])

	return err
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

const AGGREGATION_DAY = "Day"

type Points struct {
	MeteringPoints []struct {
		MeteringPoint string
		Values        []struct {
			Start time.Time
			End   time.Time
			Value float64
		}
	}
}

func (c *Client) MeteringPoints(point string, start, end time.Time, aggr string) (Points, error) {
	p := Points{}

	// Switch times if things are out of order
	if !start.Before(end) {
		end, start = start, end
	}

	url := fmt.Sprintf(
		"https://mit.seas-nve.dk/api/v1.0/profile/consumption/?meteringpoints=%s&start=%s&end=%s&aggr=%s",
		point,
		start.Format("2006-1-2"),
		end.Format("2006-1-2"),
		aggr,
	)

	err := c.do("GET", url, &p)

	return p, err
}
