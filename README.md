Nucoll is a command-line tool written in Go. It can be used to retrieve data from Twitter and is based on its predecessor, twecoll. Calls on the command line are based on a keyword which instruct nucoll what to do. Below is a list of examples followed by a brief explanation of each command. The limits of the public Twitter API apply and are indicated below.

## Examples

#### Downloading Tweets
Nucoll can download up to about 3200 tweets for a Twitter handle. A handle is specified by a screen name or user ID. It can also retrieve tweets for a given search query.

```
$ nucoll tweets jdevoo
```

The previous example would generate a `jdevoo.qry` file containing all tweets including timestamp and text in utf-8 encoding. In order search for tweets related to a certain hashtag or query expression, use the -q switch and double-quotes around the query string. Note this can retrieve many more tweets and is potentially a lengthy operations.

```
$ nucoll tweets -q "#dg2g"
```

This will also generate a `.qry` file named with a funny-looking name corresponding to the url-encoded search string.

#### Query File
A query file with extension `.qry` is just a text file that can also be created manually or produced by another tool. It contains handles which can be extracted by the init command. You could save a list of company handles to a file called `companies.qry` as in the example below.

```
@nike
@pfizer
@hugoboss
@ikea
@swatch
@pampers
@redbull
```

A switch to the tweets sub-command allows you to retrieve replies to a specified tweet ID. Note that this is limited in scope for the free API. For tweets which generates many comments, it will likely miss most of it.

#### Steps to Generating a Graph
One of the main uses of nucoll is to generate a GML file of second degree relationships. This is a three-step process that takes time due to API throttling by Twitter. In order to generate the graph, nucoll retrieves the handle's friends (or followers) and all friends-of-friends (2nd degree relationships). It then finds the relations between those, ignoring 2nd degree relationships to which the handle is not connected. In other words, it looks only for friend relationships among the friends/followers of the handle or query tweets initially supplied.

In this section, handles will be retrieved from the given handle but they could also be retrieved from a query file. First, we retrieve friends (by default) for the specified handle.

```
$ nucoll init jdevoo
```

This generates a `jdevoo.dat` file. When passed the -i option, init populates an `img` directory with avatar images. It is also possible to initialize from a query file using the -q option. Next, nucoll retrieves friends (that's always the case) of each entry in the `.dat` file.

```
$ nucoll fetch jdevoo
```

This populates the `fdat` directory with files per processed handle in `jdevoo.dat`.

This sub-command now supports the retrieval of followers who retweet content by the provided handle. It uses a maximum count of tweets per follower to examine. Note that this is a time-consuming operation considering it scans the entire follower set.

After running fetch, you generate the graph file in the third and final step.

```
$ nucoll edgelist jdevoo
```

This generates a `jdevoo.gml` file in Graph Model Language. You can use a package such as [Gephi](https://gephi.org/) to visualize your GML file. The GML file will include friends, followers, memberships and statuses counts as properties of each handle. You could then derive additional metrics e.g. the friends-to-followers or listed-to-followers ratios.

## Installation
Download the appropriate binary from the [releases](https://github.com/jdevoo/nucoll/releases) page.

On Windows, unzip the archive and place nucoll.exe on the path. On OS X and Linux, use tar e.g. `tar xf linux-amd64-nucoll.bz2` and place nucoll on the path.

If you have Go installed, you can execute `go get github.com/jdevoo/nucoll` to download, compile and install nucoll.

Then create a working directory to store the data from your expirments. Nucoll creates a number of files and folders to store its data.

* `fdat` directory containing friends of friends files
* `img` directory containing avatar images of friends
* `.dat` extension of account details data (friends, followers, avatar URL, etc. for account friends)
* `.qry` extension of tweets file (timestamp, tweet)
* `.gml` extension of edgelist file (nodes and edges)
* `.f` extension for friends data (fdat)

#### Registering Nucoll
The first time you run a nucoll command, it will ask you for the consumer key and consumer secret. Nucoll relies on oauth2 to authenticate with Twitter. You need to register your own copy of nucoll on first usage.

Twitter now makes you apply for a developer account and provide details about your intent which is just good data governance practice. This process takes some time as Twitter reviews your submission which includes an outline of your planned data experiment in 100 words. Once approved, create an application entry. Enabling sign-in with Twitter or URL callback are not required. I set this GitHub repo as website URL. Permissions are limited to read-only. If you try to run nucoll without being approved, you will get 401 errors on data from anyone but yourself. If you had twecoll previously registered and have been approved by Twitter for using its API, you can use the same Consumer API keys.

## Usage
Nucoll has built-in help and version switches invoked with -h and -v respectively. Each command can also be invoked with the help switch for additional information about its sub-options.

```
$ nucoll -h
usage: nucoll [-h] [-v]
               {resolve,init,fetch,tweets,edgelist} ...

New Collection Tool

optional arguments:
  -h    show this help message and exit
  -v    show program's version number and exit

sub-commands:
  {resolve,init,fetch,tweets,edgelist}
    init                retrieve friends data for screen_name
    fetch               retrieve friends of handles in .dat file
    edgelist            generate graph in GML format
    tweets              retrieve tweets
    resolve             retrieve user_id for screen_name or vice versa
```

## Motivation
The predecessor of nucoll is twecoll which was originally created as submission to the final assignment in Lada Adamic's SNA MOOC on Coursera (now on [openmichigan](https://open.umich.edu/find/open-educational-resources/information/si-508-networks-theory-application)). Twecoll requires the Python 2.7 runtime, is tightly coupled to Twitter and includes an optional dependency on igraph, a third-party SNA library. Instead, nucoll is a re-write in Go and ships as executables for popular operating systems. Its structure is meant to support more than one social network and relies on external tools such as Gephi for network visualization and metrics. It's also fun to learn a new programming language :-)

## License
This project is licensed under the MIT License.

## Citation
In case you use this tool to retrieve data for a paper and consider mentioning it. The version tag and commit hash have to be adapted accordingly.

J.P. de Vooght, nucoll, version Y.Y XXXXXX, (2019), GitHub repository, https://github.com/jdevoo/nucoll

```
@misc{DeVooght2019,
  author = {de Vooght, Jean-Paul},
  title = {nucoll},
  year = {2019},
  publisher = {GitHub},
  journal = {GutHub repository},
  howpublished = {\url{https://github.com/jdevoo/nucoll}},
  commit = {XXXXXX}
}
```
