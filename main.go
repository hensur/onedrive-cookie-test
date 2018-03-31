package main

import (
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"
)

// SuccessResponse hold a response from the sharepoint webdav
type SuccessResponse struct {
	XMLName xml.Name            `xml:"Envelope"`
	Succ    SuccessResponseBody `xml:"Body"`
}

// SuccessResponseBody is the body of a success response, it holds the token
type SuccessResponseBody struct {
	XMLName xml.Name
	Type    string    `xml:"RequestSecurityTokenResponse>TokenType"`
	Created time.Time `xml:"RequestSecurityTokenResponse>Lifetime>Created"`
	Expires time.Time `xml:"RequestSecurityTokenResponse>Lifetime>Expires"`
	Token   string    `xml:"RequestSecurityTokenResponse>RequestedSecurityToken>BinarySecurityToken"`
}

const reqString = `<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"
xmlns:a="http://www.w3.org/2005/08/addressing"
xmlns:u="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-utility-1.0.xsd">
<s:Header>
<a:Action s:mustUnderstand="1">http://schemas.xmlsoap.org/ws/2005/02/trust/RST/Issue</a:Action>
<a:ReplyTo>
<a:Address>http://www.w3.org/2005/08/addressing/anonymous</a:Address>
</a:ReplyTo>
<a:To s:mustUnderstand="1">https://login.microsoftonline.com/extSTS.srf</a:To>
<o:Security s:mustUnderstand="1"
 xmlns:o="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd">
<o:UsernameToken>
  <o:Username>{{ .Username }}</o:Username>
  <o:Password>{{ .Password }}</o:Password>
</o:UsernameToken>
</o:Security>
</s:Header>
<s:Body>
<t:RequestSecurityToken xmlns:t="http://schemas.xmlsoap.org/ws/2005/02/trust">
<wsp:AppliesTo xmlns:wsp="http://schemas.xmlsoap.org/ws/2004/09/policy">
  <a:EndpointReference>
    <a:Address>{{ .Address }}</a:Address>
  </a:EndpointReference>
</wsp:AppliesTo>
<t:KeyType>http://schemas.xmlsoap.org/ws/2005/05/identity/NoProofKey</t:KeyType>
<t:RequestType>http://schemas.xmlsoap.org/ws/2005/02/trust/Issue</t:RequestType>
<t:TokenType>urn:oasis:names:tc:SAML:1.0:assertion</t:TokenType>
</t:RequestSecurityToken>
</s:Body>
</s:Envelope>`

func main() {
	// Parse inpute data from command line args
	user := flag.String("user", "", "Username to get cookie for")
	pass := flag.String("pass", "", "Password for the username")
	addr := flag.String("addr", "", "The sharepoint server that holds the user account")
	flag.Parse()

	// Create map for the request Template
	reqData := map[string]interface{}{
		"Username": user,
		"Password": pass,
		"Address":  addr,
	}

	// tokenConf is a SuccessResponse and contains an AuthToken
	tokenConf := getSPToken(reqData)

	fmt.Println(getSPCookie(*addr, tokenConf))
}

func getSPCookie(hostURL string, conf *SuccessResponse) (cookieStr string) {
	spRoot, err := url.Parse(hostURL)
	if err != nil {
		panic(err)
	}

	// Create the endpoint URL
	u, err := url.Parse("https://" + spRoot.Host + "/_forms/default.aspx?wa=wsignin1.0")
	if err != nil {
		log.Fatal(err)
	}

	// To authenticate with davfs or anything else we need two cookies (rtFa and FedAuth)
	// In order to get them we use the token we got earlier and a cookieJar
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		log.Fatal(err)
	}

	// New Client with cookieJar
	client := &http.Client{
		Jar: jar,
	}

	// Send the Token as a Post request
	if _, err = client.Post(u.String(), "text/xml", strings.NewReader(conf.Succ.Token)); err != nil {
		log.Fatal(err)
	}

	for _, cookie := range jar.Cookies(u) {
		if (cookie.Name == "rtFa") || (cookie.Name == "FedAuth") {
			cookieStr += cookie.Name + "=" + cookie.Value + "; "
		}
	}
	return cookieStr
}

func getSPToken(reqData map[string]interface{}) *SuccessResponse {
	// Parse reqString as template
	t := template.Must(template.New("authXML").Parse(reqString))

	// Fill template and supply buffer to http Request
	buf := &bytes.Buffer{}
	if err := t.Execute(buf, reqData); err != nil {
		panic(err)
	}

	// Execute the first request which gives us an auth token for the sharepoint service
	req, err := http.NewRequest("POST", "https://login.microsoftonline.com/extSTS.srf", buf)
	if err != nil {
		panic(err)
	}

	// New Client to execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err.Error())
	}
	defer resp.Body.Close()

	// Convert the response to string
	respBuf := bytes.Buffer{}
	respBuf.ReadFrom(resp.Body)
	s := respBuf.Bytes()

	// Try to unmarshal response into XML which contains the auth Token
	var conf SuccessResponse
	err = xml.Unmarshal(s, &conf)
	if err != nil {
		panic(err)
	}

	return &conf
}
