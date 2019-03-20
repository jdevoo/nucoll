package twitter

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/jdevoo/nucoll/util"
)

// Twitter client with custom RoundTripper to handle throttling
type Twitter struct {
	Client *http.Client
}

// IdsResult for ids REST call
type IdsResult struct {
	IDs        []string `json:"ids"`
	NextCursor uint64   `json:"next_cursor"`
}

// MembersResult for members REST call
type MembersResult struct {
	Users      []UserObject
	NextCursor uint64 `json:"next_cursor"`
}

// SearchResult for search REST call
type SearchResult struct {
	Statuses []TweetObject
}

// UserObject defines attributes retrieved by client for a give user
// Relation can be one of "friends", "followers", "retweeter", or list name
// Subject is the handle for which the record relation holds e.g. membership of list
type UserObject struct {
	ID              uint64 `json:"id"`
	ScreenName      string `json:"screen_name"`
	Protected       bool   `json:"protected"`
	Verified        bool   `json:"verified"`
	FriendsCount    int    `json:"friends_count"`
	FollowersCount  int    `json:"followers_count"`
	ListedCount     int    `json:"listed_count"`
	StatusesCount   int    `json:"statuses_count"`
	CreatedAt       string `json:"created_at"`
	URL             string `json:"url"`
	ProfileImageURL string `json:"profile_image_url"`
	Location        string `json:"location"`
	Relation        string
	Subject         string
}

// TweetObject defines attributes retrieved by client for a given post
type TweetObject struct {
	CreatedAt string `json:"created_at"`
	ID        uint64 `json:"id"`
	User      struct {
		ScreenName string `json:"screen_name"`
	} `json:"user"`
	Text                string `json:"text"`
	InReplyToTweet      uint64 `json:"in_reply_to_status_id"`
	InReplyToUser       uint64 `json:"in_reply_to_user_id"`
	InReplyToScreenName string `json:"in_reply_to_screen_name"`
	//QuoteCount          int    `json:"quote_count"`
	//ReplyCount          int    `json:"reply_count"`
	RetweetCount  int `json:"retweet_count"`
	FavoriteCount int `json:"favorite_count"`
}

// ids returns an array of numeric user IDs
func (ns Twitter) ids(relation string, param string) ([]string, error) {
	var result IdsResult
	var ids []string
	var arg string
	const endpoint = "https://api.twitter.com/1.1/%s/ids.json?%s&stringify_ids=true"

	if util.DigitsOnly(param) {
		arg = "user_id=" + param
	} else {
		arg = "screen_name=" + param
	}
	res, err := ns.Client.Get(fmt.Sprintf(endpoint, relation, arg))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	json.NewDecoder(res.Body).Decode(&result)
	ids = result.IDs
	cursor := result.NextCursor
	for cursor != 0 {
		res, err = ns.Client.Get(fmt.Sprintf(endpoint+"&cursor=%d", relation, arg, cursor))
		if err != nil {
			return nil, err
		}
		json.NewDecoder(res.Body).Decode(&result)
		ids = append(ids, result.IDs...)
		cursor = result.NextCursor
	}
	return ids, nil
}

// members returns an array of hydrated nucoll user objects belonging to a list
func (ns Twitter) members(list string, param string) ([]UserObject, error) {
	var result MembersResult
	const endpoint = "https://api.twitter.com/1.1/lists/members.json?slug=%s&owner_screen_name=%s&count=5000&skip_status=true"

	res, err := ns.Client.Get(fmt.Sprintf(endpoint, list, param))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	json.NewDecoder(res.Body).Decode(&result)
	for i := range result.Users {
		result.Users[i].Relation = list
		result.Users[i].Subject = param
	}
	users := result.Users
	cursor := result.NextCursor
	for cursor != 0 {
		res, err = ns.Client.Get(fmt.Sprintf(endpoint+"&cursor=%d", list, param, cursor))
		if err != nil {
			return nil, err
		}
		json.NewDecoder(res.Body).Decode(&result)
		for i := range result.Users {
			result.Users[i].Relation = list
			result.Users[i].Subject = param
		}
		users = append(users, result.Users...)
		cursor = result.NextCursor
	}
	return users, nil
}

// show returns a hydrated user object for a given handle
func (ns Twitter) show(handle string) (UserObject, error) {
	var endpoint string
	var result UserObject

	if util.DigitsOnly(handle) {
		endpoint = "https://api.twitter.com/1.1/users/show.json?user_id=%s"
	} else {
		endpoint = "https://api.twitter.com/1.1/users/show.json?screen_name=%s"
	}
	res, err := ns.Client.Get(fmt.Sprintf(endpoint, handle))
	if err != nil {
		return result, err
	}
	json.NewDecoder(res.Body).Decode(&result)

	return result, nil
}

func (ns Twitter) retweetersOf(handle string, maxCount int) ([]string, error) {
	var followers []string
	var ids []string
	var err error
	var result SearchResult
	var endpoint = "https://api.twitter.com/1.1/statuses/user_timeline.json?user_id=%s&count=200&include_rts=true"

	followers, err = ns.ids("followers", handle)
	if err != nil {
		return nil, err
	}
	for _, id := range followers {
	TWEETS:
		for c, maxID := 0, uint64(0); c < maxCount; {
			if maxID != 0 {
				endpoint += fmt.Sprintf("&max_id=%d", maxID)
			}
			res, err := ns.Client.Get(fmt.Sprintf(endpoint, id))
			if err != nil {
				return nil, err
			}
			json.NewDecoder(res.Body).Decode(&result.Statuses)
			if len(result.Statuses) == 0 {
				break
			}
			log.Printf("processed %d tweets from %s\n", len(result.Statuses), id)
			for _, tweet := range result.Statuses {
				if tweet.InReplyToScreenName == handle {
					ids = append(ids, id)
					break TWEETS
				}
				if tweet.ID < maxID || maxID == 0 {
					maxID = tweet.ID
				}
			}
			if len(result.Statuses) < maxCount {
				break
			}
			c += len(result.Statuses)
			// optimization for 64 bit integers
			maxID--
		}
	}
	return ids, nil
}

// Init supports retrieve handles from: list membership, a query file, followers who retweet or a friend/follow relationship
func (ns Twitter) Init(followersFlag bool, maxPostCount int, queryFlag bool, nomentionFlag bool, membership string, imageFlag bool, args []string) {
	var result []UserObject
	var err error
	var ids []string
	var relation string
	var endpoint string
	var arg string
	var filename string

	ns.Client, err = NewClient()
	if err != nil {
		log.Fatal("failed to create Twitter client: ", err)
	}

	// list members use case: write user objects to disk and return
	if membership != "" {
		result, err = ns.members(membership, args[0])
		if err != nil {
			log.Fatal("failed to retrieve members: ", err)
		}
		filename, err = util.CSVWriter(args[0], util.DatExt, false, result)
		if err != nil {
			log.Fatal("failed to write dat file: ", err)
		}
		log.Printf("%s created\n", filename)
		return
	}

	if followersFlag {
		relation = "followers"
	} else {
		relation = "friends"
	}

	switch {
	case queryFlag:
		// query search or manually created query file
		ids, err = util.QueryReader(args[0], nomentionFlag)
	case maxPostCount > 0:
		// followers who retweet tweets by this handle
		ids, err = ns.retweetersOf(args[0], maxPostCount)
	default:
		// basic relation use case
		ids, err = ns.ids(relation, args[0])
	}
	if err != nil {
		log.Fatal(err)
	}

	// populate a hydrated array of user objects based on array of IDs
	for page, width := 0, 100; page*width < len(ids); page++ {
		if (page+1)*width >= len(ids) {
			arg = fmt.Sprint(ids[page*width:])
		} else {
			arg = fmt.Sprint(ids[page*width : (page+1)*width])
		}
		arg = strings.Trim(strings.Join(strings.Fields(arg), ","), "[]")
		if util.DigitsOnly(ids[page*width]) {
			endpoint = "https://api.twitter.com/1.1/users/lookup.json?user_id=%s"
		} else {
			endpoint = "https://api.twitter.com/1.1/users/lookup.json?screen_name=%s"
		}
		res, err := ns.Client.Get(fmt.Sprintf(endpoint, arg))
		if err != nil {
			log.Fatal("failed to use Twitter client: ", err)
		}
		json.NewDecoder(res.Body).Decode(&result)
		for i := range result {
			if maxPostCount > 0 {
				result[i].Relation = "retweeter"
			} else {
				result[i].Relation = relation
			}
			result[i].Subject = args[0]
			// ignore erros on downloads
			if imageFlag {
				util.DownloadImage(result[i].ID, result[i].ProfileImageURL)
			}
		}
		filename, err = util.CSVWriter(args[0], util.DatExt, page > 0, result)
		if err != nil {
			log.Fatal("failed to write file: ", err)
		}
		log.Printf("processed %d starting from %s\n", strings.Count(arg, ",")+1, ids[page*width])
	}
	log.Printf("%s created\n", filename)
}

// Fetch retrieves second-degree "friends" from handles collected with Init
func (ns Twitter) Fetch(forceFlag bool, fetchCount int, args []string) {
	var err error

	ns.Client, err = NewClient()
	if err != nil {
		log.Fatal("failed to create Twitter client: ", err)
	}

	data := []UserObject{}
	if err = util.CSVReader(args[0], util.DatExt, &data); err != nil {
		log.Fatal(err)
	}
	for _, user := range data {
		uid := fmt.Sprintf("%d", user.ID)
		// skip if file exists and flag to force call not set
		if !forceFlag && util.FdatExists(uid) {
			continue
		}
		if user.FriendsCount > fetchCount {
			log.Printf("skipping %s (%d friends)\n", user.ScreenName, user.FriendsCount)
			continue
		}
		ids, err := ns.ids("friends", uid)
		if err != nil {
			log.Fatal(err)
		}
		if _, err := util.FdatWriter(uid, ids); err != nil {
			log.Fatal("failed to write friends file: ", err)
		}
		log.Printf("processed %s\n", user.ScreenName)
	}
}

// Edgelist constructs the network of who is "friends" with whom among handles returned by Init
func (ns Twitter) Edgelist(egoFlag bool, missingFlag bool, args []string) {
	var cols = []string{
		"ID",
		"ScreenName",
		"Protected",
		"Verified",
		"FriendsCount",
		"FollowersCount",
		"ListedCount",
		"StatusesCount",
		"CreatedAt",
		"ProfileImageURL",
		"Relation",
		"Subject",
	}
	var filename string
	var err error

	data := []UserObject{}
	for _, handle := range args {
		if err = util.CSVReader(handle, util.DatExt, &data); err != nil {
			log.Fatal(err)
		}
		if egoFlag {
			var self UserObject
			ns.Client, err = NewClient()
			if err != nil {
				log.Fatal("failed to create Twitter client: ", err)
			}
			self, err = ns.show(handle)
			if err != nil {
				log.Fatal("failed to retrieve handle details: ", err)
			}
			data = append(data, self)
		}
	}
	// call GMLWriter using ScreenName as label for nodes
	if filename, err = util.GMLWriter(args, data, missingFlag, cols, "ScreenName"); err != nil {
		log.Fatal(err)
	}

	log.Printf("%s created\n", filename)
}

func (tweets *SearchResult) filterByTweetID(tweetID uint64) {
	pos := 0
	for i := 0; i < len((*tweets).Statuses); i++ {
		if (*tweets).Statuses[i].InReplyToTweet == tweetID {
			(*tweets).Statuses[pos] = (*tweets).Statuses[i]
			pos++
		}
	}
	(*tweets).Statuses = (*tweets).Statuses[:pos]
}

// Posts retrieves tweets from a search query, user list, replies to a given tweet ID or from a handle
func (ns Twitter) Posts(queryFlag bool, list string, postID uint64, args []string) {
	var result SearchResult
	var err error
	var endpoint string
	var filename string

	ns.Client, err = NewClient()
	if err != nil {
		log.Fatal("failed to create Twitter client: ", err)
	}

	for maxID := uint64(0); ; {
		switch {
		case queryFlag:
			endpoint = fmt.Sprintf("https://api.twitter.com/1.1/search/tweets.json?q=%s&result_type=recent&count=100", url.QueryEscape(args[0]))
		case list != "":
			endpoint = fmt.Sprintf("https://api.twitter.com/1.1/lists/statuses.json?slug=%s&owner_screen_name=%s&count=100", list, args[0])
		case postID != 0:
			endpoint = fmt.Sprintf("https://api.twitter.com/1.1/search/tweets.json?q=%s&result_type=recent&count=100&since_id=%d", url.QueryEscape("to:"+args[0]), postID)
		default:
			endpoint = fmt.Sprintf("https://api.twitter.com/1.1/statuses/user_timeline.json?screen_name=%s&count=200&include_rts=true", args[0])
		}
		if maxID != 0 {
			endpoint += fmt.Sprintf("&max_id=%d", maxID)
		}
		res, err := ns.Client.Get(endpoint)
		if err != nil {
			log.Fatal("failed to use Twitter client: ", err)
		}
		if queryFlag || postID != 0 {
			json.NewDecoder(res.Body).Decode(&result)
		} else {
			json.NewDecoder(res.Body).Decode(&result.Statuses)
		}
		if postID != 0 {
			(&result).filterByTweetID(postID)
		}
		if len(result.Statuses) == 0 {
			break
		}
		filename, err = util.CSVWriter(args[0], util.QueryExt, maxID != 0, result.Statuses)
		if maxID == 0 {
			log.Printf("%s created\n", filename)
		}
		if err != nil {
			log.Fatal("failed to write posts: ", err)
		}
		log.Printf("processed %d tweets\n", len(result.Statuses))
		for _, tweet := range result.Statuses {
			if tweet.ID < maxID || maxID == 0 {
				maxID = tweet.ID
			}
		}
		// optimization for 64 bit integers
		maxID--
	}
}

// Resolve converts screen names to IDs and vice versa along with basic stats
func (ns Twitter) Resolve(args []string) {
	var uo UserObject
	var err error

	ns.Client, err = NewClient()
	if err != nil {
		log.Fatal("failed to create Twitter client: ", err)
	}

	for _, handle := range args {
		uo, err = ns.show(handle)
		if err != nil {
			log.Fatal("failed to retrieve handle details: ", err)
		}
		if util.DigitsOnly(handle) {
			fmt.Printf("%s, %s, %d friends, %d followers, %d memberships, %d tweets\n", handle, uo.ScreenName, uo.FriendsCount, uo.FollowersCount, uo.ListedCount, uo.StatusesCount)
		} else {
			fmt.Printf("%s, %d, %d friends, %d followers, %d memberships, %d tweets\n", handle, uo.ID, uo.FriendsCount, uo.FollowersCount, uo.ListedCount, uo.StatusesCount)
		}
	}
}
