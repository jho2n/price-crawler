package csv

import (
	"bufio"
	"encoding/csv"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/seoulstore/price-crawler/search"
)

var date string

func init() {
	time.LoadLocation("Asia/Seoul")
	// 15:04:05
	date = time.Now().Format("2006-01-02")
}

func PrepareEPHeader() {
	absPath, _ := filepath.Abs("./csv/EP.csv")
	file, err := os.Create(absPath)
	if err != nil {
		panic(err)
	}

	wr := csv.NewWriter(bufio.NewWriter(file))
	wr.Write([]string{"no", "날짜", "키워드", "상품번호", "이미지", "카테고리", "상품명", "서울스토어 가격", "페이지수", "위치"})
	wr.Flush()
}

func PrepareQueryHeader() {
	absPath, _ := filepath.Abs("./csv/query.csv")
	file, err := os.Create(absPath)
	if err != nil {
		panic(err)
	}

	wr := csv.NewWriter(bufio.NewWriter(file))
	wr.Write([]string{"no", "날짜", "키워드", "페이지수", "노출상품수"})
	wr.Flush()
}

func PrepareCPHeader() {
	absPath, _ := filepath.Abs("./csv/CP.csv")
	file, err := os.Create(absPath)
	if err != nil {
		panic(err)
	}

	wr := csv.NewWriter(bufio.NewWriter(file))
	wr.Write([]string{"no", "날짜", "키워드", "id", "경쟁업체수", "최저가여부", "최저가업체", "최저가", "상품번호", "이미지", "카테고리", "상품명", "서울스토어가격", "페이지수", "위치"})
	wr.Flush()
}

func WriteEP(eps *[]search.EP) {
	absPath, _ := filepath.Abs("./csv/EP.csv")
	file, err := os.OpenFile(absPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}

	var str [][]string
	wr := csv.NewWriter(bufio.NewWriter(file))
	for idx, ep := range *eps {
		str = append(str, []string{strconv.Itoa(idx), date, ep.Query, ep.ProductID, ep.ImageURL, ep.Category, ep.ProductName, ep.Price, strconv.Itoa(ep.Page), strconv.Itoa(ep.Position)})
	}

	wr.WriteAll(str)
	wr.Flush()
}

func WriteQuery(q string, pages int, items int) {
	absPath, _ := filepath.Abs("./csv/query.csv")
	file, err := os.OpenFile(absPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}

	wr := csv.NewWriter(bufio.NewWriter(file))
	wr.Write([]string{"", date, q, strconv.Itoa(pages), strconv.Itoa(items)})
	wr.Flush()
}

func WriteCP(cps *[]search.CP) {
	absPath, _ := filepath.Abs("./csv/CP.csv")
	file, err := os.OpenFile(absPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}

	var str [][]string
	wr := csv.NewWriter(bufio.NewWriter(file))
	for idx, cp := range *cps {
		isCheepest := "N"
		if cp.IsCheepest {
			isCheepest = "Y"
		}
		str = append(str, []string{strconv.Itoa(idx), date, cp.Query, cp.ProductID, strconv.Itoa(cp.MallCount), isCheepest, cp.CheepsetMall, cp.LowPrice, cp.MallPID, cp.ImageURL, cp.Category, cp.ProductName, cp.MallPrice, strconv.Itoa(cp.Page), strconv.Itoa(cp.Position)})
	}

	wr.WriteAll(str)
	wr.Flush()
}
