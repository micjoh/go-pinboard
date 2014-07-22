package pinboard

import (
	"fmt"
	"strings"
	"time"
)

type AddPostOptions struct {
	p                          *Pinboard
	url, description, extended string
	tag                        string
	replace, shared, toread    string
	dt                         *time.Time
	errs                       []error
}

func (ap *AddPostOptions) URL(u string) *AddPostOptions {
	if err := ValidateURL(u); err != nil {
		ap.errs = append(ap.errs, err)
	} else {
		ap.url = u
	}
	return ap
}

func (ap *AddPostOptions) Title(t string) *AddPostOptions {
	if t == "" {
		ap.errs = append(ap.errs, PinboardAPIError(fmt.Sprintf("Invalid Title: %q", t)))
	} else {
		ap.description = t
	}
	return ap
}

func (ap *AddPostOptions) Description(s string) *AddPostOptions {
	ap.extended = s
	return ap
}
func (ap *AddPostOptions) Replace(b bool) *AddPostOptions {
	ap.replace = b2s[b]
	return ap
}
func (ap *AddPostOptions) Shared(b bool) *AddPostOptions {
	ap.shared = b2s[b]
	return ap
}
func (ap *AddPostOptions) Toread(b bool) *AddPostOptions {
	ap.toread = b2s[b]
	return ap
}
func (ap *AddPostOptions) Time(t *time.Time) *AddPostOptions {
	ap.dt = t
	return ap
}
func (ap *AddPostOptions) Tag(s string) *AddPostOptions {
	tags, err := parseTags(s)
	if err != nil {
		ap.errs = append(ap.errs, err)
	} else if len(tags) > 100 {
		ap.errs = append(ap.errs, PinboardAPIError("you cannot specify more than 100 tags."))
	} else {
		ap.tag = strings.Join(tags, " ")
	}
	return ap
}

func (ap *AddPostOptions) Do() error {
	if len(ap.errs) > 0 {
		return ap.errs[0]
	}
	args := &arguments{}
	args.Set("url", ap.url)
	args.Set("description", ap.description)

	if ap.extended != "" {
		args.Set("extended", ap.extended)
	}

	if len(ap.tag) > 0 {
		args.Set("tags", ap.tag)
	}

	args.Set("replace", ap.replace)
	args.Set("shared", ap.shared)
	args.Set("toread", ap.toread)
	if ap.dt != nil {
		args.Set("dt", ap.dt)
	}

	json, err := ap.p.call("posts/add", args)
	if err != nil {
		return err
	}
	code := json.Get("result_code").MustString()
	if code == "done" {
		return nil
	}
	return PinboardError(code)

}
