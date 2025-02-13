package reddit

import (
	"encoding/json"
	"os"
	"strconv"
)

type RedditPostDef struct {
	Subreddit      string
	Author         string
	MaybePost      string
	MaybeCommentId string
}

type Thing[T any] struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Kind string `json:"kind"`
	Data T      `json:"data"`
}

type Listing[T any] struct {
	Before   string `json:"before"`
	After    string `json:"after"`
	Modhash  string `json:"modhash"`
	Children []T    `json:"children"`
}

type Replies Thing[Listing[Thing[Comment]]]

type RepliesWrapper struct {
	Replies *Replies
	Empty   bool
}

func (cr *RepliesWrapper) UnmarshalJSON(data []byte) error {
	if len(data) == 2 {
		cr.Empty = true
		return nil
	}
	cr.Replies = &Replies{}
	return json.Unmarshal(data, cr.Replies)
}

type FalseOrFloat float64

func (cr *FalseOrFloat) UnmarshalJSON(data []byte) error {
	result := FalseOrFloat(100500.0)
	if data[0] == 'f' || data[0] == 'F' {
		*cr = result
		return nil
	}
	asFloat, err := strconv.ParseFloat(string(data), 64)
	if err != nil {
		return err
	}
	*cr = FalseOrFloat(asFloat)
	return nil
}

type Comment struct {
	// The number of upvotes
	Ups int32 `json:"Ups"`

	// The number of downvotes
	Downs int32 `json:"Downs"`

	// the time of creation in UTC epoch-second format. Note that neither of these ever have a non-zero fraction
	CreatedUTC float32 `json:"created_utc"`

	//who approved this comment. null if nobody or you are not a mod
	Approved_by string `json:"approved_by"`

	//the account name of the poster
	Author string `json:"author"`

	//the CSS class of the author's flair. subreddit specific
	AuthorFlairCssClass string `json:"author_flair_css_class"`

	//the text of the author's flair. subreddit specific
	AuthorFlairText string `json:"author_flair_text"`

	//who removed this comment. null if nobody or you are not a mod
	BannedBy string `json:"banned_by"`

	//the raw text. this is the unformatted text which includes the raw markup characters such as ** for bold. <, >, and & are escaped.
	Body string `json:"body"`

	//the formatted HTML text as displayed on reddit. For example, text that is emphasised by * will now have <em> tags wrapping it. Additionally, bullets and numbered lists will now be in HTML list format. NOTE: The HTML string will be escaped. You must unescape to get the raw HTML.
	BodyHtml string `json:"body_html"`

	//false if not edited, edit date in UTC epoch-seconds otherwise. NOTE: for some old edited comments on reddit.com, this will be set to true instead of edit date.
	Edited *FalseOrFloat `json:"edited"`

	//the number of times this comment received reddit gold
	Gilded int `json:"gilded"`

	//how the logged-in user has voted on the comment - True = upvoted, False = downvoted, null = no vote
	Likes bool `json:"likes"`

	//present if the comment is being displayed outside its thread (user pages, /r/subreddit/comments/.json, etc.). Contains the author of the parent link
	LinkAuthor string `json:"link_author"`

	//ID of the link this comment is in
	LinkId string `json:"link_id"`

	//present if the comment is being displayed outside its thread (user pages, /r/subreddit/comments/.json, etc.). Contains the title of the parent link
	LinkTitle string `json:"link_title"`

	//present if the comment is being displayed outside its thread (user pages, /r/subreddit/comments/.json, etc.). Contains the URL of the parent link
	LinkUrl string `json:"link_url"`

	//how many times this comment has been reported, null if not a mod
	NumReports int `json:"num_reports"`

	//ID of the thing this comment is a reply to, either the link or a comment in it
	ParentId string `json:"parent_id"`

	//A list of replies to this comment
	Replies *RepliesWrapper `json:"replies"`

	//true if this post is saved by the logged in user
	Saved bool `json:"saved"`

	//the net-score of the comment
	Score int `json:"score"`

	//Whether the comment's score is currently hidden.
	ScoreHidden bool `json:"score_hidden"`

	//subreddit of thing excluding the /r/ prefix. "pics"
	Subreddit string `json:"subreddit"`

	//the id of the subreddit in which the thing is located
	SubredditId string `json:"subreddit_id"`

	//to allow determining whether they have been distinguished by moderators/admins. null = not distinguished. moderator = the green [M]. admin = the red [A]. special = various other special distinguishes http://redd.it/19ak1b
	Distinguished string `json:"distinguished"`
}

func ParseLink(link string) *RedditPostDef {
	subexp := redditRe.SubexpNames()
	submatch := redditRe.FindStringSubmatch(link)
	result := &RedditPostDef{}
	for i, key := range submatch {
		switch subexp[i] {
		case "subreddit":
			result.Subreddit = key
		case "author":
			result.Author = key
		case "maybePostName":
			result.MaybePost = key
		case "maybeCommentId":
			result.MaybeCommentId = key
		}
	}
	return result
}

type Client interface {
	GetPost(def *RedditPostDef) error
}

type redditClient struct {
}

func (rc *redditClient) GetPost(def *RedditPostDef) error {
	result := []Thing[Listing[Thing[Comment]]]{}
	resdata, err := os.ReadFile("/home/arusakov/devel/lanseg/chronicler/reddit.json")
	if err != nil {
		return err
	}
	if err = json.Unmarshal(resdata, &result); err != nil {
		return err
	}

	resBAfter, err := json.Marshal(result)
	if err != nil {
		return err
	}
	return os.WriteFile("/home/arusakov/devel/lanseg/chronicler/reddit-out.json", resBAfter, 0777)
}
