package pinboard

import (
	"fmt"
	"strings"
)

type GetRecentPostsOptions struct {
	p     *Pinboard
	tag   string
	count int
	errs  []error
}

func (rp *GetRecentPostsOptions) Count(i int) *GetRecentPostsOptions {
	if i > 100 {
		rp.errs = append(rp.errs, PinboardAPIError(fmt.Sprintf("you cannot get more than %d recent posts", recentCountMax)))
	} else {
		rp.count = i
	}
	return rp
}

func (rp *GetRecentPostsOptions) Tags(s string) *GetRecentPostsOptions {
	tags, err := parseTags(s)
	if err != nil {
		rp.errs = append(rp.errs, err)
	} else if len(tags) > 3 {
		rp.errs = append(rp.errs, PinboardAPIError("you cannot specify more than 3 tags."))
	} else {
		rp.tag = strings.Join(tags, " ")
	}
	return rp
}

func (rp *GetRecentPostsOptions) Do() ([]*Post, error) {
	if len(rp.errs) > 0 {
		return nil, rp.errs[0]
	}
	args := &arguments{}
	if len(rp.tag) > 0 {
		args.Set("tag", rp.tag)
	}
	if rp.count == 0 {
		rp.count = recentCountDefault
	}
	args.Set("count", rp.count)

	json, err := rp.p.call("posts/recent", args)
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
