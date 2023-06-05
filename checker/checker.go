package checker

import (
	"encoding/json"
	"errors"
	"find_qty/m"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"unicode"

	"github.com/PuerkitoBio/goquery"
)

func StoreAmount(inputID string, client *http.Client) (string, error) {

	_, err := strconv.Atoi(inputID)
	if err != nil {
		return "", errors.New(m.WrongCode)
	}

	name, elementID, err := getNameAndID(inputID, client)
	if err != nil {
		return "", err
	}

	uri := "https://www.virage24.ru/ajax/productStoreAmount.php"

	//values for method post from "uri"
	vals := url.Values{}

	vals.Add("USE_STORE_PHONE", "Y")
	vals.Add("SCHEDULE", "")
	vals.Add("USE_MIN_AMOUNT", "N")
	vals.Add("MIN_AMOUNT", "10")
	vals.Add("ELEMENT_ID", elementID)
	vals.Add("STORE_PATH", "%2Fcontacts%2Fstores%2F%23store_id%23%2F")
	vals.Add("MAIN_TITLE", "%D0%9D%D0%B0%D0%BB%D0%B8%D1%87%D0%B8%D0%B5+%D0%BD%D0%B0+%D1%81%D0%BA%D0%BB%D0%B0%D0%B4%D0%B0%D1%85%2C+%D0%BC%D0%B0%D0%B3%D0%B0%D0%B7%D0%B8%D0%BD%D0%B0%D1%85")
	vals.Add("MAX_AMOUNT", "20")
	vals.Add("USE_ONLY_MAX_AMOUNT", "Y")
	vals.Add("SHOW_EMPTY_STORE", "Y")
	vals.Add("SHOW_GENERAL_STORE_INFORMATION", "N")
	vals.Add("USER_FIELDS%5B%5D", "")
	vals.Add("USER_FIELDS%5B%5D", "UF_CATALOG_ICON")
	vals.Add("USER_FIELDS%5B%5D", "")
	vals.Add("FIELDS%5B%5D", "")
	vals.Add("FIELDS%5B%5D", "")
	vals.Add("STORES_FILTER_ORDER", "SORT_ASC")
	vals.Add("STORES_FILTER", "SORT")
	vals.Add("STORES%5B%5D", "2")
	vals.Add("STORES%5B%5D", "4")
	vals.Add("STORES%5B%5D", "9")
	vals.Add("STORES%5B%5D", "7")
	vals.Add("STORES%5B%5D", "3")
	vals.Add("STORES%5B%5D", "10")
	vals.Add("STORES%5B%5D", "14")
	vals.Add("USE_STORES", "true")
	vals.Add("SITE_ID", "s1")

	resp, err := client.PostForm(uri, vals)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	stores, err := parseHTML(resp, "span")
	if err != nil {
		return "", err
	}

	res, err := formatMsg(stores)
	if err != nil {
		return "", err
	}

	result := name + "\n\n" + res

	return result, nil
}

func getNameAndID(inputID string, client *http.Client) (string, string, error) {

	//struct for response
	type FacetValue struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	type Category struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Direct   bool   `json:"direct"`
		LinkURL  string `json:"link_url"`
		ImageURL string `json:"image_url"`
	}

	type AttributeList map[string][]string

	type Facet struct {
		Name     string       `json:"name"`
		DataType string       `json:"dataType"`
		Values   []FacetValue `json:"values"`
	}

	type Product struct {
		ID         string        `json:"id"`
		Available  bool          `json:"available"`
		Name       string        `json:"name"`
		Price      string        `json:"price"`
		Score      float64       `json:"score"`
		Categories []Category    `json:"categories"`
		Attributes AttributeList `json:"attributes"`
		LinkURL    string        `json:"link_url"`
		ImageURL   string        `json:"image_url"`
	}

	type Response struct {
		Query          string       `json:"query"`
		Correction     string       `json:"correction"`
		TotalHits      int          `json:"totalHits"`
		ZeroQueries    bool         `json:"zeroQueries"`
		Products       []Product    `json:"products"`
		Facets         []Facet      `json:"facets"`
		SelectedFacets []FacetValue `json:"selectedFacets"`
	}

	urlSearch := "https://sort.diginetica.net/search?st=" + inputID + "&apiKey=2C0VV885MK&fullData=true"

	resp, err := client.Get(urlSearch)
	if err != nil {
		return "", "", err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	var response Response

	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", "", err
	}

	if len(response.Products) > 0 {
		product := response.Products[0]
		name := product.Name + "\n"
		price := product.Price + "\n"
		prodInfo := fmt.Sprintf("%s\nЦена - %s", name, price)
		return prodInfo, product.ID, nil
	}

	return "", "", errors.New("not found ID")
}

func parseHTML(r *http.Response, tag string) (string, error) {

	doc, err := goquery.NewDocumentFromReader(r.Body)
	if err != nil {
		return "", err
	}

	b := strings.Builder{}

	spanTags := doc.Find(tag)
	spanTags.Each(func(i int, s *goquery.Selection) {
		b.WriteString(s.Text())
	})

	return b.String(), nil
}

func formatMsg(html string) (string, error) {

	b := strings.Builder{}

	for _, v := range html {
		if unicode.Is(unicode.Cyrillic, v) ||
			unicode.Is(unicode.Digit, v) ||
			v == ':' ||
			v == ' ' ||
			v == ',' ||
			v == '\n' ||
			v == '.' {
			_, err := b.WriteRune(v)
			if err != nil {
				return "", err
			}
		}
	}

	list := strings.Split(b.String(), "\n")
	if len(list) < 12 {
		return "", errors.New(m.SomethingWrongWithStores)
	}

	str := strings.Join(list[9:11], "\n")

	return str, nil
}
