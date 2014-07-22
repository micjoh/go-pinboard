package pinboard

import (
	"strings"
	"time"
)

type FilterFunc func(*Post) bool

type GetAllPostsOptions struct {
	p              *Pinboard
	tag            string
	start, results int
	fromdt, todt   *time.Time
	meta           int
	errs           []error
}

func (gp *GetAllPostsOptions) Tag(s string) *GetAllPostsOptions {
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

func (gp *GetAllPostsOptions) Start(i int) *GetAllPostsOptions {
	gp.start = i
	return gp
}

func (gp *GetAllPostsOptions) Results(i int) *GetAllPostsOptions {
	gp.results = i
	return gp
}

func (gp *GetAllPostsOptions) From(t *time.Time) *GetAllPostsOptions {
	gp.fromdt = t
	return gp
}

func (gp *GetAllPostsOptions) To(t *time.Time) *GetAllPostsOptions {
	gp.todt = t
	return gp
}

func (gp *GetAllPostsOptions) Meta(b bool) *GetAllPostsOptions {
	if b {
		gp.meta = 1
	} else {
		gp.meta = 0
	}
	return gp
}

func (gp *GetAllPostsOptions) Do() ([]*Post, error) {
	if len(gp.errs) > 0 {
		return nil, gp.errs[0]
	}
	args := &arguments{}
	if gp.tag != "" {
		args.Set("tag", gp.tag)
	}
	if gp.fromdt != nil {
		args.Set("fromdt", gp.fromdt)
	}
	if gp.todt != nil {
		args.Set("todt", gp.todt)
	}

	args.Set("start", gp.start)
	if gp.results > 0 {
		args.Set("results", gp.results)
	}

	json, err := gp.p.call("posts/all", args)
	if err != nil {
		return nil, err
	}
	posts := json.MustArray()
	l := len(posts)
	out := make([]*Post, l)
	for i := 0; i < l; i++ {
		out[i] = parsePost(json.GetIndex(i))
	}
	return out, nil
}
