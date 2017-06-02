package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/msiebuhr/seas-nve"
)

var email = flag.String("email", os.Getenv("EMAIL"), "Username/email for mit.seas-nve.dk (also reads $EMAIL)")
var password = flag.String("password", os.Getenv("PASSWORD"), "Username for mit.seas-nve.dk (also reads $PASSWORD)")

func main() {
	flag.Parse()

	c := seasnve.NewClient()

	err := c.Login(*email, *password)
	if err != nil {
		fmt.Println("# error logging in:", err)
		os.Exit(1)
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
