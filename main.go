package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"sync"
)

// Agente estándar
const (
	userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		"Chrome/53.0.2785.143 " +
		"Safari/537.36"
)

func downloadFile(URL, fileName string, wg *sync.WaitGroup) error {
	defer wg.Done()
	// Obtiene las respuestas al buscar las imágenes
	response, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return errors.New("Received non 200 response code")
	}
	// Crea un archivo vacío
	file, err := os.Create("./data/" + fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	// Escribe el conjunto de bytes al archivo.
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}
	return nil
}

func fetchUrl(url string, chFailedUrls chan string, chIsFinished chan bool) {

	// Abre la url
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", userAgent)
	resp, err := client.Do(req)

	defer func() {
		chIsFinished <- true
	}()

	// Si no se logra abrir, crear un objeto con los erróneos
	if err != nil || resp.StatusCode != 200 {
		chFailedUrls <- url
		return
	} else {
		// Lee el cuerpo de la respuesta
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
		}
		// convierte el cuerpo de la respuesta a string
		sb := string(body)
		var re1 = regexp.MustCompile(`img data-src="(.*?)"`)
		ms := re1.FindAllStringSubmatch(sb, -1)
		ss := make([]string, len(ms))
		var wg sync.WaitGroup
		// registra la instancia de la rutina
		wg.Add(len(ms))
		for i, m := range ms {
			fileName := "image" + strconv.Itoa(i) + ".jpeg"
			URL := m[1]
			ss[i] = fileName
			go downloadFile(URL, fileName, &wg)
		}
		wg.Wait()
		fmt.Println("done.")
		fmt.Println(ss)
	}

}

func main() {

	// Crea una búsqueda de imagenes
	query := "perros"
	urlsList := [10]string{
		"https://www.google.cl/search?q=" + query + "&hl=en&tbm=isch",
	}

	chFailedUrls := make(chan string)
	chIsFinished := make(chan bool)

	for _, url := range urlsList {
		// Creamos un thread desde donde anidaremos más threads.
		go fetchUrl(url, chFailedUrls, chIsFinished)
	}

	failedUrls := make([]string, 0)
	for i := 0; i < len(urlsList); {
		select {
		case url := <-chFailedUrls:
			failedUrls = append(failedUrls, url)
		case <-chIsFinished:
			i++
		}
	}

}
