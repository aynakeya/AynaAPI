package imomoe

import (
	"AynaAPI/api/httpc"
	"AynaAPI/api/model"
	"AynaAPI/api/model/e"
	"AynaAPI/utils/vhttp"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"regexp"
	"strconv"
	"strings"
)

const (
	searchApi = Host + "/search.asp?searchword=%s&page=%d"
)

func GetSearchApi(keyword string, page int) string {
	return fmt.Sprintf(searchApi, vhttp.QueryEscapeWithEncoding(keyword, "gb2312"), page)
}

func Search(keyword string, page int) model.ApiResponse {
	if page == 0 {
		page = 1
	}
	result := vhttp.DecodeString(httpc.Get(GetSearchApi(keyword, page), nil).String(),
		"gb2312")
	if result == "" {
		return model.CreateEmptyApiResponseByStatus(e.INTERNAL_ERROR)
	}
	rexp := regexp.MustCompile("页次:[0-9]+/[0-9]+页")
	rawPageNum := []rune(rexp.FindString(result))
	if len(rawPageNum) == 0 {
		return model.CreateEmptyApiResponseByStatus(e.EXTERNAL_API_ERROR)
	}
	pageNum := strings.Split(string(rawPageNum[3:len(rawPageNum)-1]), "/")
	currPage, _ := strconv.Atoi(pageNum[0])
	totalPage, _ := strconv.Atoi(pageNum[1])

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(result))
	if err != nil {
		return model.CreateEmptyApiResponseByStatus(e.INTERNAL_ERROR)
	}
	videoList := make([]*ImomoeVideo, 0)
	doc.Find("div[class~=fire] > div[class=pics] > ul > li").Each(func(i int, selection *goquery.Selection) {
		info := selection.Find("h2 > a")
		if href, b := info.Attr("href"); b {
			title, _ := info.Attr("title")
			v := InitDefault()
			v.Id = regexp.MustCompile("[0-9]+").FindString(href)
			v.Title = title
			if pic, b := selection.Find("a > img").Attr("src"); b {
				v.PictureUrl = pic
			}
			videoList = append(videoList, v)
		}
	})
	return model.CreateApiResponseByStatus(e.SUCCESS, map[string]interface{}{
		"videoList":   videoList,
		"currentPage": currPage,
		"totalPage":   totalPage,
	})
}