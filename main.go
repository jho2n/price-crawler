package main

import (
	"log"
	"sync"

	"github.com/seoulstore/price-crawler/csv"
	"github.com/seoulstore/price-crawler/search"
	"github.com/seoulstore/price-crawler/sheet"
)

func work(wg *sync.WaitGroup, q string) {
	defer wg.Done()

	// fmt.Printf("==============%s===================\n", q)
	totalItemCount, err := search.TotalCount(q)
	if err != nil {
		log.Panic(err)
	}

	totalPageCount := totalItemCount / search.PagingSize
	if totalItemCount%search.PagingSize > 1 {
		totalPageCount++
	}

	// fmt.Printf("total item count: %d\n", totalItemCount)
	// fmt.Printf("total page count: %d\n", totalPageCount)

	csv.WriteQuery(q, totalPageCount, totalItemCount)

	var pResult search.ProductResult
	var wg2 sync.WaitGroup
	wg2.Add(3)

	for i := 1; i <= 3; i++ {
		go search.Products2(&wg2, q, i, &pResult)
	}
	wg2.Wait()

	csv.WriteEP(&pResult.EPS)

	var wg3 sync.WaitGroup
	wg3.Add(len(pResult.CPS))

	for i := 0; i < len(pResult.CPS); i++ {
		item := &pResult.CPS[i]
		go search.CompareProducts2(&wg3, item)
	}
	wg3.Wait()

	// fmt.Printf("\n\n%s END\n\n", q)
}

// \[[a-z0-9가-힣 ,]*착용\]
func main() {
	brands := sheet.GetBrands()

	var wg sync.WaitGroup
	wg.Add(len(brands))

	csv.PrepareEPHeader()
	csv.PrepareQueryHeader()
	for _, q := range brands {
		go work(&wg, q)
	}

	wg.Wait()
}
