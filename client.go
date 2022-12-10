package mox

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/publicsuffix"
)

var (
	moxBaseURL = "https://mox.moe"
)

type Comics struct {
	Title   string
	Authors []string
	Books   []*Book
}

type Book struct {
	ID           string
	Name         string
	MobiVIPPath  string
	MobiVIP2Path string
	EpubVIPPath  string
	EpubVIP2Path string
}

func NewClient() (client *http.Client, err error) {
	var jar *cookiejar.Jar
	jar, err = cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})

	if err != nil {
		return
	}

	client = &http.Client{
		Jar: jar,
	}

	return
}

type MoxClient struct {
	*http.Client
	Ctx context.Context
}

func NewMoxClient(ctx context.Context) (c *MoxClient, err error) {
	var client *http.Client

	client, err = NewClient()
	if err != nil {
		return
	}

	c = &MoxClient{Ctx: ctx, Client: client}

	return
}

func (c *MoxClient) Login() (err error) {
	var u string
	var form url.Values
	var req *http.Request
	var resp *http.Response
	config := c.Ctx.Value("config").(*Config)

	u = fmt.Sprintf("%s/%s", moxBaseURL, "login_do.php")

	form = url.Values{}
	form.Add("email", config.Mox.Email)
	form.Add("passwd", config.Mox.Password)

	req, err = http.NewRequest("POST", u, strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err = c.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	return
}

func (c *MoxClient) Logout() (err error) {
	var u string
	var resp *http.Response

	u = fmt.Sprintf("%s/%s", moxBaseURL, "logout.php")

	resp, err = c.Get(u)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	return
}

func (c *MoxClient) GetComics(id int) (comics *Comics, err error) {
	var u string
	var resp *http.Response
	var doc *goquery.Document
	var sel *goquery.Selection

	u = fmt.Sprintf("%s/c/%d.htm", moxBaseURL, id)

	resp, err = c.Get(u)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound {
		var patBookData *regexp.Regexp
		var patVolInfo *regexp.Regexp
		var html, bookDataPath string
		var mVolInfo [][]string
		var bodyBytes []byte

		doc, err = goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			return
		}

		patBookData = regexp.MustCompile(`book_data.php\?h=\w+`)
		patVolInfo = regexp.MustCompile(`"volinfo=(.+?)"`)

		comics = new(Comics)
		comics.Title = doc.Find(".font_big").Text()

		sel = doc.Find("font.status").Eq(0)
		sel = sel.Find("a")
		for i := 0; i < len(sel.Nodes); i++ {
			node := sel.Eq(i)
			if node.Text() != "" {
				comics.Authors = append(comics.Authors, node.Text())
			}
		}

		html, err = doc.Html()
		if err != nil {
			return
		}

		bookDataPath = patBookData.FindString(html)
		if bookDataPath == "" {
			err = fmt.Errorf("book data is not found")
			return
		}

		u = fmt.Sprintf("%s/%s", moxBaseURL, bookDataPath)

		resp, err = c.Get(u)
		if err != nil {
			return
		}
		defer resp.Body.Close()

		bodyBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			return
		}

		mVolInfo = patVolInfo.FindAllStringSubmatch(string(bodyBytes), -1)
		if mVolInfo == nil {
			err = fmt.Errorf("volinfo is not found")
			return
		}

		for _, m := range mVolInfo {
			volInfo := strings.Split(m[1], ",")
			bookID := volInfo[0]
			bookName := volInfo[5]
			book := &Book{
				ID:           bookID,
				Name:         bookName,
				MobiVIPPath:  fmt.Sprintf("down/%d/%s/0/1/1-0/", id, bookID),
				MobiVIP2Path: fmt.Sprintf("down/%d/%s/1/1/1-0/", id, bookID),
				EpubVIPPath:  fmt.Sprintf("down/%d/%s/0/2/1-0/", id, bookID),
				EpubVIP2Path: fmt.Sprintf("down/%d/%s/1/2/1-0/", id, bookID),
			}
			comics.Books = append(comics.Books, book)
		}
	} else {
		err = fmt.Errorf("%d", resp.StatusCode)
	}

	return
}

func (c *MoxClient) DownloadComics(id int) (err error) {
	var comics *Comics
	config := c.Ctx.Value("config").(*Config)

	comics, err = c.GetComics(id)
	if err != nil {
		return
	}

	for _, book := range comics.Books {
		var f *os.File
		var u string
		var resp *http.Response

		fileName := fmt.Sprintf(
			"%s/[Mox][%s]%s.kepub.epub",
			config.Mox.DownloadPath,
			comics.Title,
			book.Name,
		)

		u = fmt.Sprintf("%s/%s", moxBaseURL, book.EpubVIP2Path)
		fmt.Println(fmt.Sprintf("download book from %s to %s...", u, fileName))

		f, err = os.Create(fileName)
		if err != nil {
			return
		}

		resp, err = c.Get(u)
		if err != nil {
			return
		}
		defer resp.Body.Close()

		_, err = io.Copy(f, resp.Body)

		defer f.Close()
	}

	return
}
