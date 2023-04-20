package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

type Message struct {
	Code       string `json:"code,omitempty"`
	State      string `json:"state,omitempty"`
	City       string `json:"city,omitempty"`
	District   string `json:"district,omitempty"`
	Address    string `json:"address,omitempty"`
	Cep        string `json:"cep,omitempty"`
	Logradouro string `json:"logradouro,omitempty"`
	Localidade string `json:"localidade,omitempty"`
	Bairro     string `json:"bairro,omitempty"`
	Uf         string `json:"uf,omitempty"`
	Source     string `json:"source"`
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	params, _ := url.ParseQuery(r.URL.RawQuery)
	if len(params["cep"]) != 1 {
		fmt.Println("CEP parameter invalid")
		return
	}

	cep := params["cep"][0]
	apicep := make(chan Message)
	viacep := make(chan Message)

	go func() {
		apicep <- doRequest(fmt.Sprintf("https://cdn.apicep.com/file/apicep/%s-%s.json", cep[0:5], cep[5:]))
	}()

	go func() {
		viacep <- doRequest(fmt.Sprintf("http://viacep.com.br/ws/%s/json/", cep))
	}()

	var msg Message
	for {
		select {
		case msg = <-apicep:
			msg.Source = "APICEP"
		case msg = <-viacep:
			msg.Source = "ViaCEP"
		case <-time.After(time.Second * 1):
			println("timeout")
		}

		jsonResp, err := json.Marshal(msg)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(msg)

		w.Write(jsonResp)
		return
	}
}

func doRequest(uri string) Message {
	fmt.Println("URL: ", uri)
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()

	req = req.WithContext(ctx)
	c := &http.Client{}
	res, err := c.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	var message Message
	if err := json.Unmarshal(body, &message); err != nil {
		log.Fatal(err)
	}

	return message
}
