package imomoe

import (
	"AynaAPI/api/model"
	"AynaAPI/api/model/e"
	"encoding/json"
	"fmt"
	"github.com/spf13/cast"
	"regexp"
	"strings"
)

const Host string = "http://www.imomoe.la"

type ImomoeVideo struct {
	Id         string              `json:"id"`
	SourceId   string              `json:"source_id"`
	EpId       string              `json:"ep_id"`
	Title      string              `json:"title"`
	PictureUrl string              `json:"pic"`
	Urls       []string            `json:"-"`
	Episodes   map[string][]string `json:"episodes"`
}

func InitDefault() *ImomoeVideo {
	return &ImomoeVideo{
		Id:       "0",
		SourceId: "0",
		EpId:     "0",
	}
}

func InitWithUid(uid string) *ImomoeVideo {
	idString := regexp.MustCompile("imomoe-[0-9]+-[0-9]+-[0-9]+").FindString(uid)
	if idString == "" {
		return nil
	}
	idstringL := strings.Split(idString, "-")
	return &ImomoeVideo{
		Id:       idstringL[1],
		SourceId: idstringL[2],
		EpId:     idstringL[3],
	}
}

func InitWithUrl(url string) *ImomoeVideo {
	return nil
}

func GenerateUniqueId(id string, sourcdId string, epId string) string {
	return fmt.Sprintf("imomoe-%s-%s-%s", id, sourcdId, epId)
}

// MarshalJSON method from http://choly.ca/post/go-json-marshalling/
func (self *ImomoeVideo) MarshalJSON() ([]byte, error) {
	type FakeV ImomoeVideo
	return json.Marshal(&struct {
		Uid string `json:"uid"`
		*FakeV
	}{
		Uid:   self.GetUniqueId(),
		FakeV: (*FakeV)(self),
	})
}

func (self *ImomoeVideo) GetUniqueId() string {
	return fmt.Sprintf("imomoe-%s-%s-%s", self.Id, self.SourceId, self.EpId)
}

func (self *ImomoeVideo) Initialize() bool {
	resp := GetInfo(self.Id, self.SourceId, self.EpId)
	if resp.Status != e.SUCCESS {
		return false
	}
	self.PictureUrl = cast.ToString(resp.Data["pic"])
	self.Title = cast.ToString(resp.Data["title"])
	self.Episodes = cast.ToStringMapStringSlice(resp.Data["episodes"])
	resp = GetPlayData(cast.ToStringMapString(resp.Data)["playdataUrl"])
	if resp.Status != e.SUCCESS {
		return false
	}
	self.Urls = cast.ToStringSlice(resp.Data["urls"])
	return true
}

func (self *ImomoeVideo) GetPlayUrls() []model.ApiResource {
	pUrls := make([]model.ApiResource, 0)
	for _, url := range self.Urls {
		resp := ResolveVideoUrl(url)
		if resp.Status == 0 {
			pUrls = append(pUrls, model.ApiResource{
				Url:    cast.ToString(resp.Data["realUrl"]),
				Header: nil,
			})
		}
	}
	return pUrls
}