package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func main() {
	url := "https://groupietrackers.herokuapp.com/api/artists"

	response, err := http.Get(url)

	if err != nil {
		log.Fatal(err)
	}

	defer response.Body.Close()

	if response.StatusCode == http.StatusOK {
		bodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			log.Fatal(err)
		}

		data := string(bodyBytes)

		fmt.Println(data)
	}
}
