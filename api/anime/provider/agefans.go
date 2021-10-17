package provider

import (
	"AynaAPI/api/anime/core"
	"AynaAPI/api/anime/rule"
	"AynaAPI/api/core/e"
	"AynaAPI/api/httpc"
	"AynaAPI/utils/vhttp"
	"AynaAPI/utils/vstring"
	"fmt"
	"github.com/aynakeya/deepcolor"
	"github.com/spf13/cast"
	"github.com/tidwall/gjson"
	"math"
	"regexp"
	"strings"
	"time"
)

type Agefans struct {
	BaseUrl    string
	SearchAPI  string
	PlayUrlAPI string
	Rules      rule.AgefansRules
}

func (p *Agefans) GetName() string {
	return "agefans"
}

func (p *Agefans) Validate(meta core.ProviderMeta) bool {
	return meta.Name == p.GetName() &&
		regexp.MustCompile("^"+regexp.QuoteMeta(p.BaseUrl)).FindString(meta.Url) != ""
}

func _newAgefans() *Agefans {
	return &Agefans{
		BaseUrl:    "https://www.agefans.cc",
		SearchAPI:  "https://www.agefans.cc/search?query=%s&page=%d",
		PlayUrlAPI: "https://www.agefans.cc/_getplay?aid=%s&playindex=%s&epindex=%s",
		Rules:      rule.InitializeAgefansRules(),
	}
}

var AgefansAPI *Agefans

func init() {
	AgefansAPI = _newAgefans()
}

func (p *Agefans) getSearchApi(keyword string) string {
	return fmt.Sprintf(p.SearchAPI, keyword, 1)
}

func (p *Agefans) getPlayUrlAPI(aid string, playindex string, epindex string) string {
	return fmt.Sprintf(p.PlayUrlAPI, aid, playindex, epindex)
}

func (p *Agefans) GetAnimeMeta(meta core.ProviderMeta) (core.AnimeMeta, error) {
	aMeta := core.AnimeMeta{Provider: meta}
	if !p.Validate(meta) {
		return aMeta, e.NewError(e.PROVIDER_META_NOT_VALIED)
	}
	err := p.UpdateAnimeMeta(&aMeta)
	return aMeta, err
}

func (p *Agefans) UpdateAnimeMeta(meta *core.AnimeMeta) error {
	id := regexp.MustCompile("/detail/[0-9]+").FindString(meta.Provider.Url)
	if id == "" {
		return e.NewError(e.INTERNAL_ERROR)
	}
	meta.Provider.Name = "agefans"
	result, err := deepcolor.Fetch(deepcolor.Tentacle{
		Url:         meta.Provider.Url,
		Charset:     "utf-8",
		ContentType: deepcolor.TentacleContentTypeHTMl,
	}, httpc.GetCORSString, nil, nil)
	if err != nil {
		return e.NewError(e.EXTERNAL_API_ERROR)
	}
	meta.Title = result.GetSingle(p.Rules.InfoTitle)
	meta.Year = result.GetSingle(p.Rules.InfoYear)
	meta.Cover = result.GetSingle(p.Rules.InfoCover)
	meta.Description = result.GetSingle(p.Rules.InfoDesc)
	meta.Tags = strings.Split(result.GetSingle(p.Rules.InfoTag), " ")
	return nil
}

func (p *Agefans) GetAnime(meta core.AnimeMeta) (core.Anime, error) {
	anime := core.Anime{AnimeMeta: meta}
	err := p.UpdateAnime(&anime)
	return anime, err
}
func (p *Agefans) UpdateAnime(anime *core.Anime) error {
	result, err := deepcolor.Fetch(deepcolor.Tentacle{
		Url:         anime.Provider.Url,
		Charset:     "utf-8",
		ContentType: deepcolor.TentacleContentTypeHTMl,
	}, httpc.GetCORSString, nil, nil)
	if err != nil {
		return e.NewError(e.EXTERNAL_API_ERROR)
	}
	ids := result.GetList(p.Rules.InfoVideos)
	urlNames := result.GetList(p.Rules.InfoVideoNames)
	anime.Playlists = make([]core.Playlist, 0)
	current_playlist_id := "-1"
	current_playlist_index := -1
	for index, id := range ids {
		tmp := strings.Split(id, "-")
		if len(tmp) < 3 {
			continue
		}
		animeId, playlistId, epId := tmp[0], tmp[1], tmp[2]
		if playlistId != current_playlist_id {
			current_playlist_id = playlistId
			anime.Playlists = append(anime.Playlists, core.Playlist{
				Name:   playlistId,
				Videos: make([]core.AnimeVideo, 0),
			})
			current_playlist_index = len(anime.Playlists) - 1
		}
		anime.Playlists[current_playlist_index].Videos = append(anime.Playlists[current_playlist_index].Videos,
			core.AnimeVideo{
				Title: urlNames[index],
				Url:   "",
				Provider: core.ProviderMeta{
					Name: "",
					Url:  p.getPlayUrlAPI(animeId, playlistId, epId),
				},
			})
	}

	//for _, playlist := range anime.Playlists {
	//	for _, v := range playlist.Videos {
	//		err = p.UpdateAnimeVideo(&v)
	//		if err != nil{
	//		}
	//	}
	//}
	return nil
}

func (p *Agefans) getCookie(t1 int) string {
	timeNow := time.Now().UnixNano() / (1000000)
	t1Tmp := int64(math.Round(float64(t1)/1000)) >> 0x5
	k2 := (t1Tmp*(t1Tmp%0x1000)*0x3+0x1450f)*(t1Tmp%0x1000) + t1Tmp
	t2 := timeNow
	t2 = t2 - t2%10 + k2%10
	return fmt.Sprintf("t1=%d;k2=%d;t2=%d", t1, k2, t2)
}

func (p *Agefans) UpdateAnimeVideo(video *core.AnimeVideo) error {
	url := video.Provider.Url
	resp := httpc.Head(url, map[string]string{
		"referer": p.BaseUrl,
	})

	initiator := regexp.MustCompile("t1=[^;]*;").FindString(resp.Header.Get("set-cookie"))

	if initiator == "" {
		return e.NewError(e.EXTERNAL_API_ERROR)
	}
	t1, _ := vstring.SliceString(initiator, 3, -1)

	authCookie := p.getCookie(cast.ToInt(t1))
	resp = httpc.Get(url, map[string]string{
		"referer": p.BaseUrl,
		"cookie":  authCookie,
	})
	video.Provider.Name = regexp.MustCompile("</?play>").
		ReplaceAllString(gjson.Parse(resp.String()).Get("playid").String(), "")
	video.Url = vhttp.QueryUnescapeWithEncoding(gjson.Parse(resp.String()).Get("vurl").String(), "utf-8")
	return nil
}

func (p *Agefans) Search(keyword string) (core.AnimeSearchResult, error) {
	result, err := deepcolor.Fetch(deepcolor.Tentacle{
		Url:         p.getSearchApi(keyword),
		Charset:     "utf-8",
		ContentType: deepcolor.TentacleContentTypeHTMl,
	}, httpc.GetCORSString, nil, nil)

	if err != nil {
		return core.AnimeSearchResult{}, e.NewError(e.EXTERNAL_API_ERROR)
	}
	var sResults = make([]core.AnimeMeta, 0)
	urls := result.GetList(p.Rules.SearchURL)
	titles := result.GetList(p.Rules.SearchTitle)
	years := result.GetList(p.Rules.SearchYear)
	tags := result.GetList(p.Rules.SearchTag)
	covers := result.GetList(p.Rules.SearchCover)
	desc := result.GetList(p.Rules.SearchDesc)
	for index, url := range urls {
		meta := core.AnimeMeta{
			Title:       titles[index],
			Year:        years[index],
			Tags:        strings.Split(tags[index], " "),
			Cover:       covers[index],
			Description: desc[index],
			Provider: core.ProviderMeta{
				Name: p.GetName(),
				Url:  vhttp.JoinUrl(p.BaseUrl, url),
			},
		}
		sResults = append(sResults, meta)
	}
	return core.AnimeSearchResult{Result: sResults}, nil
}