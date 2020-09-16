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
var regex = regexp.MustCompile(`\d+\/297159\/\d+\/\d+\/\|`)
var searchBaseURL = "https://search.shopping.naver.com/search/all"
var compareBaseURL = "https://search.shopping.naver.com/detail/lite.nhn"

type ProductResult struct {
	CPS []CP
	EPS []EP
}

type EP struct {
	Category     string `json:"category"`
	ImageURL     string `json:"imageURL"`
	Page         int    `json:"page"`
	Position     int    `json:"position"`
	Price        string `json:"price"`
	ProductID    string `json:"productID"`
	ProductName  string `json:"productName"`
	ProductTitle string `json:"productTitle"`
	Query        string `json:"query"`
	URL          string `json:"productURL"`
}

type CP struct {
	Category     string       `json:"category"`
	CheepsetMall string       `json:"CheepestMall"`
	ImageURL     string       `json:"imageURL"`
	IsCheepest   bool         `json:"isCheepest"`
	LowPrice     string       `json:"lowPrice"`
	MallCount    int          `json:"mallCount"`
	MallPID      string       `json:"mallPID"`
	MallPrice    string       `json:"mallPrice"`
	Page         int          `json:"page"`
	Position     int          `json:"position"`
	ProductID    string       `json:"productID"`
	ProductName  string       `json:"productName"`
	Products     []CPProducts `json:"products"`
	ProductTitle string       `json:"productTitle"`
	ProductURL   string       `json:"productURL"`
	Query        string       `json:"query"`
}

type CPProducts struct {
	Mall  string `json:"mall"`
	Name  string `json:"name"`
	Price string `json:"price"`
	URL   string `json:"url"`
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

func createEP(category string, query string, page int, idx int, item *gabs.Container) EP {
	return EP{
		Category:     category,
		ImageURL:     replaceDoublequote(item.Search("imageUrl").String()),
		Page:         page,
		Position:     idx,
		Price:        replaceDoublequote(item.Search("price").String()),
		ProductID:    replaceDoublequote(item.Search("mallProductId").String()),
		ProductName:  replaceDoublequote(item.Search("productName").String()),
		ProductTitle: replaceDoublequote(item.Search("productTitle").String()),
		Query:        query,
		URL:          replaceDoublequote(item.Search("mallProductUrl").String()),
	}
}

func createCP(category string, query string, page int, idx int, item *gabs.Container, hasLowPriceByMallNo string, hasSS []int) CP {
	ssInfo := strings.Split(strings.Replace(hasLowPriceByMallNo[hasSS[0]:hasSS[1]], "|", "", -1), "/")
	nProductID := item.Search("id").String()

	p := url.Values{}
	p.Add("nvMid", nProductID)
	return CP{Category: category,
		CheepsetMall: replaceDoublequote(item.Path("lowMallList.0.name").String()),
		ImageURL:     replaceDoublequote(item.Search("imageUrl").String()),
		IsCheepest:   hasSS[0] == 0,
		LowPrice:     replaceDoublequote(item.Search("lowPrice").String()),
		MallCount:    len(strings.Split(hasLowPriceByMallNo, "|")),
		MallPID:      ssInfo[len(ssInfo)-3],
		MallPrice:    ssInfo[len(ssInfo)-2],
		Page:         page,
		Position:     idx,
		ProductID:    replaceDoublequote(nProductID),
		ProductName:  replaceDoublequote(item.Search("productName").String()),
		ProductTitle: replaceDoublequote(item.Search("productTitle").String()),
		ProductURL:   replaceDoublequote(compareBaseURL + "?" + p.Encode()),
		Query:        query,
	}
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

	for idx, child := range container.S("list").Children() {
		item := child.Search("item")
		adID := item.Search("adId").String()
		if adID != "null" {
			continue
		}
		mallNo := item.Search("mallNo").Data()
		hasLowPriceByMallNo := replaceDoublequote(item.Search("lowPriceByMallNo").String())
		hasSS := regex.FindIndex([]byte(hasLowPriceByMallNo))

		var category string
		if mallNo == strconv.Itoa(ssMallID) || len(hasSS) > 0 {
			c1name := item.Search("category1Name")
			c2name := item.Search("category2Name")
			c3name := item.Search("category3Name")
			category = replaceDoublequote(fmt.Sprintf("%s > %s > %s", c1name, c2name, c3name))
		}

		if mallNo == strconv.Itoa(ssMallID) {
			ep := createEP(category, query, page, idx, item)
			eps = append(eps, ep)
		}

		if len(hasSS) > 0 {
			cp := createCP(category, query, page, idx, item, hasLowPriceByMallNo, hasSS)
			cps = append(cps, cp)
		}
	}
	pResult.EPS = append(pResult.EPS, eps...)
	pResult.CPS = append(pResult.CPS, cps...)
}

// func CompareProducts2(wg *sync.WaitGroup, item *CP) {
// 	defer wg.Done()

// 	doc, err := getDoc(item.ProductURL)
// 	if err != nil {
// 		return
// 	}

// 	doc.Find("#section_price_list table[data-chnl-seq]").Each(func(i int, s *goquery.Selection) {
// 		mallNameNode := s.Find("a[data-mall-name]")
// 		mallName, _ := mallNameNode.Attr("data-mall-name")

// 		pNameNode := s.Find("td.lft > a")
// 		pName := strings.TrimSpace(pNameNode.Text())
// 		pLink, _ := pNameNode.Attr("href")
// 		price, _ := pNameNode.Attr("data-product-price")
// 		cmp := CPProducts{
// 			Mall:  mallName,
// 			Name:  pName,
// 			Price: price,
// 			URL:   pLink,
// 		}
// 		item.AddProduct(cmp)
// 	})
// }
