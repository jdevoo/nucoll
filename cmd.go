package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jdevoo/nucoll/twitter"
	"github.com/jdevoo/nucoll/util"
)

// SocialNetworkService defines the interface for services such as Twitter
type SocialNetworkService interface {
	Init(followersFlag bool, maxPostCount int, queryFlag bool, nomentionFlag bool, list string, imageFlag bool, args []string)
	Fetch(forceFlag bool, fetchCount int, args []string)
	Edgelist(egoFlag bool, missingFlag bool, args []string)
	Posts(queryFlag bool, list string, postID uint64, args []string)
	Resolve(args []string)
}

var (
	version      string // set by go tool
	golang       string // set by go tool
	githash      string // set by go tool
	initMembers  string
	maxPostCount int
	fetchCount   int
	postsList    string
	postsPostID  uint64

	helpFlag    = flag.Bool("h", false, "show this help message and exit")
	versionFlag = flag.Bool("v", false, "print version and exit")

	initCommand       = flag.NewFlagSet("init", flag.ExitOnError)
	initFollowersFlag = initCommand.Bool("o", false, "retrieve followers (default friends)")
	initQueryFlag     = initCommand.Bool("q", false, fmt.Sprintf("extract handles from %s file (default screen_name)", util.QueryExt))
	initNomentionFlag = initCommand.Bool("n", false, fmt.Sprintf("ignore mentions from %s file (default false)", util.QueryExt))
	initImageFlag     = initCommand.Bool("i", false, "download images (default false)")

	edgelistCommand     = flag.NewFlagSet("edgelist", flag.ExitOnError)
	edgelistEgoFlag     = edgelistCommand.Bool("e", false, "include screen_name (default false)")
	edgelistMissingFlag = edgelistCommand.Bool("m", false, "include missing handles (default false)")

	fetchCommand   = flag.NewFlagSet("fetch", flag.ExitOnError)
	fetchForceFlag = fetchCommand.Bool("f", false, fmt.Sprintf("ignore existing %s files (default false)", util.FdatExt))

	resolveCommand = flag.NewFlagSet("resolve", flag.ExitOnError)

	postsCommand   = flag.NewFlagSet("tweets", flag.ExitOnError)
	postsQueryFlag = postsCommand.Bool("q", false, "argument is a quoted query string (default screen_name)")

	// Usage overrides PrintDefaults
	Usage = func() {
		fmt.Println("Usage: " + filepath.Base(os.Args[0]) + " [-h] [-v]")
		fmt.Println("              {init,fetch,edgelist,tweets,resolve} ...")
		fmt.Println()
		fmt.Println("New Collection Tool")
		fmt.Println()
		fmt.Println("Sub-commands:")
		fmt.Println("  init         retrieve friends data for screen_name")
		fmt.Printf("  fetch        retrieve friends of handles in %s file\n", util.DatExt)
		fmt.Println("  edgelist     generate graph in GML format")
		fmt.Println("  tweets       retrieve tweets")
		fmt.Println("  resolve      retrieve user_id for screen_name or vice versa")
		fmt.Println()
		fmt.Println("Optional arguments:")
		flag.PrintDefaults()
	}
)

func init() {
	initCommand.StringVar(&initMembers, "m", "", "extract member handles from list owned by screen_name")
	initCommand.IntVar(&maxPostCount, "r", 0, "tweet count limit when looking for retweets by followers")
	initCommand.Usage = func() {
		fmt.Println("Usage: " + filepath.Base(os.Args[0]) + " init [-h] [i] [-m list] [-n] [-o] [-q] [-r N] screen_name")
		initCommand.PrintDefaults()
	}
	fetchCommand.IntVar(&fetchCount, "c", 5000, "skip if friends count above limit")
	fetchCommand.Usage = func() {
		fmt.Println("Usage: " + filepath.Base(os.Args[0]) + " fetch [-h] [-c N] [-f] screen_name")
		fetchCommand.PrintDefaults()
	}
	edgelistCommand.Usage = func() {
		fmt.Println("Usage: " + filepath.Base(os.Args[0]) + " edgelist [-h] [-e] [-m] screen_name [screen_name...]")
		edgelistCommand.PrintDefaults()
	}
	resolveCommand.Usage = func() {
		fmt.Println("Usage: " + filepath.Base(os.Args[0]) + " resolve [-h] screen_name [screen_name...]")
	}
	postsCommand.StringVar(&postsList, "m", "", "extract tweets from list")
	postsCommand.Uint64Var(&postsPostID, "p", 0, "replies to tweet id by screen_name")
	postsCommand.Usage = func() {
		fmt.Println("Usage: " + filepath.Base(os.Args[0]) + " tweets [-h] [-p id] [-m list] [-q] <screen_name | \"query\">")
		postsCommand.PrintDefaults()
	}
}

func main() {
	var sns SocialNetworkService

	flag.Parse()
	if *versionFlag {
		fmt.Printf("New Collection Tool %s (%s %s)\n", version, golang, githash)
		os.Exit(0)
	}
	if len(flag.Args()) == 0 || *helpFlag {
		Usage()
		os.Exit(1)
	}

	sns = twitter.Twitter{}

	switch os.Args[1+flag.NFlag()] {
	case "init":
		if err := initCommand.Parse(os.Args[2:]); err == nil {
			if initCommand.NArg() == 1 {
				sns.Init(*initFollowersFlag, maxPostCount, *initQueryFlag, *initNomentionFlag, initMembers, *initImageFlag, initCommand.Args())
			} else {
				initCommand.Usage()
				os.Exit(1)
			}
		}
	case "edgelist":
		if err := edgelistCommand.Parse(os.Args[2:]); err == nil {
			if edgelistCommand.NArg() > 0 {
				sns.Edgelist(*edgelistEgoFlag, *edgelistMissingFlag, edgelistCommand.Args())
			} else {
				edgelistCommand.Usage()
				os.Exit(1)
			}
		}
	case "fetch":
		if err := fetchCommand.Parse(os.Args[2:]); err == nil {
			if fetchCommand.NArg() == 1 {
				sns.Fetch(*fetchForceFlag, fetchCount, fetchCommand.Args())
			} else {
				fetchCommand.Usage()
				os.Exit(1)
			}
		}
	case "resolve":
		if err := resolveCommand.Parse(os.Args[2:]); err == nil {
			if resolveCommand.NArg() > 0 {
				sns.Resolve(resolveCommand.Args())
			} else {
				resolveCommand.Usage()
				os.Exit(1)
			}
		}
	case "tweets":
		if err := postsCommand.Parse(os.Args[2:]); err == nil {
			if postsCommand.NArg() > 0 {
				sns.Posts(*postsQueryFlag, postsList, postsPostID, postsCommand.Args())
			} else {
				postsCommand.Usage()
				os.Exit(1)
			}
		}
	default:
		fmt.Printf("%q is not a valid command\n", os.Args[1])
		os.Exit(1)
	}
	os.Exit(0)
}
