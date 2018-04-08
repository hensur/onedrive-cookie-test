package main

import (
	"flag"
	"fmt"

	"github.com/hensur/onedrive-cookie-test/odrvcookie"
)

func main() {
	// Parse inpute data from command line args
	user := flag.String("user", "", "Username to get cookie for")
	pass := flag.String("pass", "", "Password for the username")
	addr := flag.String("addr", "", "The sharepoint server that holds the user account")
	flag.Parse()

	ca := odrvcookie.New(*user, *pass, *addr)

	// tokenConf is a SuccessResponse and contains an AuthToken
	tokenConf, err := ca.Cookies()

	if err != nil{
		fmt.Println("error: ", err.Error())
	}

	fmt.Println(tokenConf)
}
