package pinboard

import (
	"strings"
	"time"
)

type GetPostsOptions struct {
	p    *Pinboard
	tag  string
	dt   *time.Time
	url  string
	meta string
	errs []error
}

func (gp *GetPostsOptions) URL(u string) *GetPostsOptions {
	if err := ValidateURL(u); err != nil {
		gp.errs = append(gp.errs, err)
	} else {
		gp.URL(u)
	}
	return gp
}

func (gp *GetPostsOptions) Meta(b bool) *GetPostsOptions {
	gp.meta = b2s[b]
	return gp
}

func (gp *GetPostsOptions) Time(t *time.Time) *GetPostsOptions {
	gp.dt = t
	return gp
}

func (gp *GetPostsOptions) Tags(s string) *GetPostsOptions {
	tags, err := parseTags(s)
	if err != nil {
		gp.errs = append(gp.errs, err)
	} else if len(tags) > 3 {
		gp.errs = append(gp.errs, PinboardAPIError("you cannot specify more than 3 tags."))
	} else {
		gp.tag = strings.Join(tags, " ")
	}
	return gp
}

func (gp *GetPostsOptions) Do() ([]*Post, error) {
	if len(gp.errs) > 0 {
		return nil, gp.errs[0]
	}
	args := &arguments{}
	if len(gp.tag) > 0 {
		args.Set("tag", gp.tag)
	}
	if gp.dt != nil {
		args.Set("dt", gp.dt)
	}
	args.Set("url", gp.url)
	args.Set("meta", gp.meta)

	json, err := gp.p.call("posts/get", args)
	if err != nil {
		return nil, err
	}
	posts := json.Get("posts").MustArray()
	l := len(posts)
	out := make([]*Post, l)
	for i := 0; i < l; i++ {
		out[i] = parsePost(json.Get("posts").GetIndex(i))
	}
	return out, nil
}
