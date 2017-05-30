package main

import (
	"fmt"
	"flag"
	"os"

	"github.com/msiebuhr/seas-nve"
)

var email = flag.String("email", os.Getenv("EMAIL"), "Username/email for mit.seas-nve.dk (also reads $EMAIL)")
var password = flag.String("password", os.Getenv("PASSWORD"), "Username for mit.seas-nve.dk (also reads $PASSWORD)")

func main() {
	flag.Parse()

	c, err := seasnve.NewClient(*email, *password)
	if err != nil {
		panic(err)
	}

	data, err := c.Metering()
	if err != nil {
		panic(err)
	}

	fmt.Println("# HELP seas_nve_meter_total_kwh The total number of kWh used this year")
	fmt.Println("# TYPE seas_nve_meter_total_kwh counter")
	for _, d := range data.MeteringPoints {
		fmt.Printf("seas_nve_meter_total_kwh{customer=\"%s\",meteringpoint=\"%s\"} %f\n", d.CustomerNumber, d.MeteringPoint, d.ConsumptionYearToDate)
	}
}
