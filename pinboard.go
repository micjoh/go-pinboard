package pinboard

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bitly/go-simplejson"
)

const apiURL = "https://api.pinboard.in/"
const recentCountMax = 100
const recentCountDefault = 15

var b2s = map[bool]string{true: "yes", false: "no"}

var validURLSchemes = [...]string{"http", "https", "javascript", "mailto", "ftp", "file", "feed"}

type Pinboard struct {
	username       string
	token          string
	api            *url.URL // https://api.pinboard.in/v1/
	version        string
	lastStatusCode int
}

func GetAuthToken(username, password string) (string, error) {
	uri, _ := url.Parse("https://api.pinboard.in/")
	p := &Pinboard{
		username: username,
		api:      uri,
		version:  "v1",
	}
	p.api.User = url.UserPassword(username, password)

	json, err := p.call("user/api_token", nil)
	if err != nil {
		return "", err
	}
	return username + ":" + json.Get("result").MustString(), nil
}

func New(token string) *Pinboard {
	uri, _ := url.Parse(apiURL)
	p := &Pinboard{
		token:    token,
		username: strings.SplitN(token, ":", 2)[0],
		api:      uri,
		version:  "v1",
	}
	v := url.Values{}
	v.Add("auth_token", token)
	p.api.RawQuery = v.Encode()
	return p
}

func NewWithPassword(username, password string) *Pinboard {
	uri, _ := url.Parse(apiURL)
	p := &Pinboard{
		username: username,
		api:      uri,
		version:  "v1",
	}
	p.api.User = url.UserPassword(username, password)
	return p
}

func (p *Pinboard) GetUpdatedTime() (*time.Time, error) {
	json, err := p.call("posts/update", nil)
	if err != nil {
		return nil, err
	}
	t, err := time.Parse(time.RFC3339, json.Get("update_time").MustString())
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (p *Pinboard) AddPost(url, title string) error {
	ap := &AddPostOptions{p: p}
	return ap.URL(url).Title(title).Do()
}

func (p *Pinboard) AddPostWith(url, title string) *AddPostOptions {
	ap := &AddPostOptions{p: p}
	return ap.URL(url).Title(title)
}

func (p *Pinboard) DeletePost(url string) error {
	if err := ValidateURL(url); err != nil {
		return err
	}
	args := &arguments{}
	args.Set("url", url)
	json, err := p.call("posts/delete", args)
	if err != nil {
		return err
	}
	code := json.Get("result_code").MustString()
	if code == "done" {
		return nil
	}
	return PinboardError(code)
}

func (p *Pinboard) GetPost(url string) (*Post, error) {
	posts, err := p.GetPostsWith().URL(url).Do()
	if err != nil || len(posts) == 0 {
		return nil, err
	}
	return posts[0], nil
}

func (p *Pinboard) GetPostsWith() *GetPostsOptions {
	return &GetPostsOptions{p: p}
}
func (p *Pinboard) GetPosts() ([]*Post, error) {
	return p.GetPostsWith().Do()
}
func (p *Pinboard) GetRecentPosts() ([]*Post, error) {
	return p.GetRecentPostsWith().Do()
}
func (p *Pinboard) GetRecentPostsWith() *GetRecentPostsOptions {
	rp := &GetRecentPostsOptions{p: p}
	return rp
}

type PostsAtDates struct {
	Date  time.Time
	Count int
}

type postsAtDateByCount []PostsAtDates

func (a postsAtDateByCount) Len() int           { return len(a) }
func (a postsAtDateByCount) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a postsAtDateByCount) Less(i, j int) bool { return a[i].Count > a[j].Count }

func (p *Pinboard) GetPostsAtDates() ([]PostsAtDates, error) {
	var args *arguments
	// args := &arguments{}
	json, err := p.call("posts/dates", args)
	if err != nil {
		return nil, err
	}
	var pd []PostsAtDates
	for d, c := range json.MustMap() {
		date, _ := time.Parse("2006-01-02", d)
		count, _ := strconv.Atoi(c.(string))
		pd = append(pd, PostsAtDates{date, count})
	}

	sort.Sort(postsAtDateByCount(pd))
	return pd, nil
}

func (p *Pinboard) GetAllPostsWith() *GetAllPostsOptions {
	gp := &GetAllPostsOptions{p: p}
	return gp
}

func (p *Pinboard) GetAllPosts() ([]*Post, error) {
	return p.GetAllPostsWith().Do()
}

func (p *Pinboard) GetSuggestedTags(url string) (popular, recommended []string, err error) {
	err = ValidateURL(url)
	if err != nil {
		return
	}
	json, err := p.call("posts/suggest", &arguments{"url": url})
	if err != nil {
		return
	}
	pop := json.GetIndex(0).Get("popular").MustArray()
	rec := json.GetIndex(1).Get("recommended").MustArray()

	popular = make([]string, 0, len(pop))
	recommended = make([]string, 0, len(recommended))
	for _, tag := range pop {
		popular = append(popular, tag.(string))
	}
	for _, tag := range rec {
		recommended = append(recommended, tag.(string))
	}
	return
}

type TagCloud []struct {
	Tag   string
	Count string
	count int
}

type tagCloudByCount TagCloud

func (a tagCloudByCount) Len() int           { return len(a) }
func (a tagCloudByCount) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a tagCloudByCount) Less(i, j int) bool { return a[i].count > a[j].count }

func (p *Pinboard) GetAllTags() (TagCloud, error) {
	json, err := p.call("tags/get", nil)
	if err != nil {
		return nil, err
	}
	var tags TagCloud
	for t, c := range json.MustMap() {
		count, _ := strconv.Atoi(c.(string))
		tags = append(tags, struct {
			Tag   string
			Count string
			count int
		}{t, c.(string), count})
	}
	sort.Sort(tagCloudByCount(tags))
	return tags, nil
}

func (p *Pinboard) DeleteTag(tag string) error {
	args := &arguments{}
	args.Set("tag", tag)
	json, err := p.call("tags/delete", args)
	if err != nil {
		return err
	}
	code := json.Get("result_code").MustString()
	if code == "done" {
		return nil
	}
	return PinboardError(code)
}
func (p *Pinboard) RenameTag(o, n string) error {
	args := &arguments{}
	args.Set("old", o)
	args.Set("new", n)
	json, err := p.call("tags/rename", args)
	if err != nil {
		return err
	}
	code := json.Get("result_code").MustString()
	if code == "done" {
		return nil
	}
	return PinboardError(code)

}
func (p *Pinboard) GetUserSecret() (string, error) {
	json, err := p.call("user/secret", nil)
	if err != nil {
		return "", err
	}
	return json.Get("result").MustString(), nil
}

func ValidateURL(u string) error {
	parsed, err := url.Parse(u)
	if err != nil {
		return err
	}
	for _, scheme := range validURLSchemes {
		if parsed.Scheme == scheme {
			return nil
		}
	}
	return PinboardAPIError(fmt.Sprintf("Invalid URL: %q", u))
}

//TODO: GetNotes
//TODO: GetNote

func parseTags(s string) ([]string, error) {
	if s == "" {
		return nil, nil
	}
	tags := strings.Split(s, " ")
	for _, t := range tags {
		if strings.ContainsAny(t, "\t\n\v\f\r \u0085\u00A0") {
			return nil, PinboardAPIError(fmt.Sprintf("Invalid Tag: %q", t))
		}
	}
	return tags, nil
}

func parsePost(j *simplejson.Json) *Post {
	shared := false
	if j.Get("shared").MustString() == "yes" {
		shared = true
	}
	toread := false
	if j.Get("toread").MustString() == "yes" {
		toread = true
	}
	tag := j.Get("tags").MustString()
	t, _ := time.Parse(time.RFC3339, j.Get("time").MustString())

	return &Post{
		URL:         j.Get("href").MustString(),
		Title:       j.Get("description").MustString(),
		Description: j.Get("extended").MustString(),
		Hash:        j.Get("hash").MustString(),
		Meta:        j.Get("meta").MustString(),
		Shared:      shared,
		Tag:         tag,
		Time:        t,
		Toread:      toread,
	}

}

func (p *Pinboard) call(method string, args *arguments) (*simplejson.Json, error) {
	p.api.Path = p.version + "/" + method
	if args == nil {
		args = &arguments{}
	}

	args.Set("format", "json")
	if p.token != "" {
		args.Set("auth_token", p.token)
	}

	p.api.RawQuery = args.Encode()
	resp, err := http.Get(p.api.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	json, err := simplejson.NewJson(d)

	switch resp.StatusCode {
	case 429:
		return nil, ErrTooManyRequests
	case 403:
		return nil, ErrForbidden
	case 200:
		return json, nil
	default:
		return nil, PinboardError(resp.Status)
	}
}
