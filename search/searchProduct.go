package search

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/Jeffail/gabs/v2"
	"github.com/PuerkitoBio/goquery"
)

var ssMallID = 297159
var PagingSize = 80
var regex = regexp.MustCompile(`\/297159\/`)
var searchBaseURL = "https://search.shopping.naver.com/search/all"
var compareBaseURL = "https://search.shopping.naver.com/detail/lite.nhn"

type ProductResult struct {
	EPS []EP
	CPS []CP
}

type EP struct {
	Query        string `json:"query"`
	URL          string `json:"productURL"`
	ProductID    string `json:"productID"`
	Price        string `json:"price"`
	ProductName  string `json:"productName"`
	ProductTitle string `json:"productTitle"`
	ImageURL     string `json:"imageURL"`
	Category     string `json:"category"`
}

type CP struct {
	ProductID    string       `json:"productID"`
	ProductURL   string       `json:"productURL"`
	LowPrice     string       `json:"lowPrice"`
	ProductName  string       `json:"productName"`
	ProductTitle string       `json:"productTitle"`
	Products     []CPProducts `json:"products"`
}

type CPProducts struct {
	URL   string `json:"url"`
	Price string `json:"price"`
	Name  string `json:"name"`
	Mall  string `json:"mall"`
}

func (cp *CP) AddProduct(item CPProducts) []CPProducts {
	cp.Products = append(cp.Products, item)
	return cp.Products
}

func getDoc(url string) (doc *goquery.Document, err error) {
	doc = nil
	res, err := http.Get(url)

	if err != nil {
		return
	}

	defer res.Body.Close()

	doc, err = goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return
	}

	return
}

func parseSearchDoc(doc *goquery.Document) (pointer *gabs.Container, err error) {
	pointer = nil
	s := doc.Find("#__NEXT_DATA__")
	jsonParsed, err := gabs.ParseJSON([]byte(s.Text()))
	if err != nil {
		return
	}

	pointer, err = jsonParsed.JSONPointer("/props/pageProps/initialState/products")
	if err != nil {
		return
	}

	return
}

func getSearchDoc(params url.Values) (container *gabs.Container, err error) {
	container = nil
	doc, err := getDoc(searchBaseURL + "?" + params.Encode())
	if err != nil {
		return
	}

	container, err = parseSearchDoc(doc)
	if err != nil {
		return
	}

	return
}

func TotalCount(query string) (total int, err error) {
	total = 0
	params := url.Values{}
	params.Add("query", query)
	params.Add("pagingSize", "1")

	container, err := getSearchDoc(params)
	if err != nil {
		return
	}

	strTotal := container.Search("total").String()
	total, err = strconv.Atoi(strTotal)

	return
}

func replaceDoublequote(str string) string {
	return strings.ReplaceAll(str, "\"", "")
}

func Products2(wg *sync.WaitGroup, query string, page int, pResult *ProductResult) {
	defer wg.Done()
	params := url.Values{}
	params.Add("query", query)
	params.Add("pagingSize", strconv.Itoa(PagingSize))
	params.Add("pagingIndex", strconv.Itoa(page))

	container, err := getSearchDoc(params)
	if err != nil {
		return
	}
	var eps []EP
	var cps []CP

	for _, child := range container.S("list").Children() {
		item := child.Search("item")
		adID := item.Search("adId").String()
		if adID != "null" {
			continue
		}
		mallNo := item.Search("mallNo").Data()
		hasLowPriceByMallNo := item.Search("lowPriceByMallNo")

		mallStr := regex.FindString(hasLowPriceByMallNo.String())

		if mallNo == strconv.Itoa(ssMallID) {
			c1name := item.Search("category1Name")
			c2name := item.Search("category2Name")
			c3name := item.Search("category3Name")
			ep := EP{
				Query:        query,
				URL:          replaceDoublequote(item.Search("mallProductUrl").String()),
				ProductID:    replaceDoublequote(item.Search("mallProductId").String()),
				Price:        replaceDoublequote(item.Search("price").String()),
				ProductName:  replaceDoublequote(item.Search("productName").String()),
				ProductTitle: replaceDoublequote(item.Search("productTitle").String()),
				ImageURL:     replaceDoublequote(item.Search("imageUrl").String()),
				Category:     replaceDoublequote(fmt.Sprintf("%s > %s > %s", c1name, c2name, c3name)),
			}
			eps = append(eps, ep)
		}

		if len(mallStr) == 8 {
			nProductID := item.Search("id").String()
			p := url.Values{}
			p.Add("nvMid", nProductID)
			cp := CP{
				ProductID:    replaceDoublequote(nProductID),
				ProductURL:   replaceDoublequote(compareBaseURL + "?" + p.Encode()),
				LowPrice:     replaceDoublequote(item.Search("lowPrice").String()),
				ProductName:  replaceDoublequote(item.Search("productName").String()),
				ProductTitle: replaceDoublequote(item.Search("productTitle").String()),
			}
			cps = append(cps, cp)
		}
	}
	pResult.EPS = append(pResult.EPS, eps...)
	pResult.CPS = append(pResult.CPS, cps...)
}

func CompareProducts2(wg *sync.WaitGroup, item *CP) {
	defer wg.Done()

	doc, err := getDoc(item.ProductURL)
	if err != nil {
		return
	}

	doc.Find("#section_price_list table[data-chnl-seq]").Each(func(i int, s *goquery.Selection) {
		mallNameNode := s.Find("a[data-mall-name]")
		mallName, _ := mallNameNode.Attr("data-mall-name")

		pNameNode := s.Find("td.lft > a")
		pName := strings.TrimSpace(pNameNode.Text())
		pLink, _ := pNameNode.Attr("href")
		price, _ := pNameNode.Attr("data-product-price")
		cmp := CPProducts{
			Mall:  mallName,
			Name:  pName,
			Price: price,
			URL:   pLink,
		}
		item.AddProduct(cmp)
	})
}
