// Package dodp implements DAISY Online Delivery Protocol v1.
package dodp

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"time"
)

// Supported input types
const (
	TEXT_NUMERIC      = "TEXT_NUMERIC"
	TEXT_ALPHANUMERIC = "TEXT_ALPHANUMERIC"
	AUDIO             = "AUDIO"
)

// The identifiers of the content list for getContentList operation
const (
	Issued  = "issued"
	New     = "new"
	Expired = "expired"
)

// The identifiers of the question for getQuestions operation
const (
	Default = "default"
	Search  = "search"
	Back    = "back"
)

// SOAP message envelope
type envelope struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	Body    body
}

// SOAP message body
type body struct {
	XMLName xml.Name `xml:"Body"`
	Content any
}

func (b *body) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	for {
		token, err := d.Token()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		switch v := token.(type) {
		// We unmarshal only the first element inside the body as content. All other elements, if present, are ignored
		case xml.StartElement:
			if err := d.DecodeElement(b.Content, &v); err != nil {
				return err
			}
			return d.Skip()
		}
	}
}

// SOAP fault
type Fault struct {
	XMLName     xml.Name `xml:"Fault"`
	Faultstring string   `xml:"faultstring"`
}

func (f *Fault) Error() string {
	return f.Faultstring
}

// DAISY Online client
type Client struct {
	url        string
	httpClient *http.Client
	ctx        context.Context
}

func NewClient(url string, timeout time.Duration) *Client {
	return NewClientWithContext(context.TODO(), url, timeout)
}

// Creates an instance of a new DAISY Online client with context and the specified service URL.
// Timeout limits the execution time of each HTTP request for this client.
// Zero timeout means no timeout.
func NewClientWithContext(ctx context.Context, url string, timeout time.Duration) *Client {
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic("Invalid cookie jar")
	}

	return &Client{
		url: url,
		httpClient: &http.Client{
			Jar:     jar,
			Timeout: timeout,
		},
		ctx: ctx,
	}
}

func (c *Client) call(action string, args any, rs any) error {
	var reqEnv envelope
	reqEnv.Body.Content = args

	buf := bytes.NewBufferString(xml.Header)
	enc := xml.NewEncoder(buf)
	if err := enc.Encode(reqEnv); err != nil {
		return err
	}
	if err := enc.Close(); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(c.ctx, http.MethodPost, c.url, buf)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "text/xml; charset=utf-8")
	req.Header.Add("Accept", "text/xml")
	req.Header.Add("Accept-Encoding", "gzip")
	req.Header.Add("SOAPAction", "/"+action)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return err
		}
		reader = gzipReader
		defer gzipReader.Close()
	}

	var respEnv envelope
	dec := xml.NewDecoder(reader)

	if resp.StatusCode != http.StatusOK {
		fault := &Fault{}
		respEnv.Body.Content = fault
		if err := dec.Decode(&respEnv); err != nil {
			return err
		}
		return fmt.Errorf("fault: %w", fault)
	}

	respEnv.Body.Content = rs
	return dec.Decode(&respEnv)
}

type logOn struct {
	XMLName  xml.Name `xml:"http://www.daisy.org/ns/daisy-online/ logOn"`
	Username string   `xml:"username"`
	Password string   `xml:"password"`
}

type logOnResponse struct {
	XMLName     xml.Name `xml:"http://www.daisy.org/ns/daisy-online/ logOnResponse"`
	LogOnResult bool     `xml:"logOnResult"`
}

// Logs a Reading System on to a Service.
func (c *Client) LogOn(username, password string) (bool, error) {
	action := "logOn"
	req := logOn{
		Username: username,
		Password: password,
	}
	resp := logOnResponse{}
	if err := c.call(action, req, &resp); err != nil {
		return false, fmt.Errorf("%v operation: %w", action, err)
	}
	return resp.LogOnResult, nil
}

type logOff struct {
	XMLName xml.Name `xml:"http://www.daisy.org/ns/daisy-online/ logOff"`
}

type logOffResponse struct {
	XMLName      xml.Name `xml:"http://www.daisy.org/ns/daisy-online/ logOffResponse"`
	LogOffResult bool     `xml:"logOffResult"`
}

// Logs a Reading System off a Service.
// A return value of false or a Fault both indicate that the operation was not successful.
func (c *Client) LogOff() (bool, error) {
	action := "logOff"
	req := logOff{}
	resp := logOffResponse{}
	if err := c.call(action, req, &resp); err != nil {
		return false, fmt.Errorf("%v operation: %w", action, err)
	}
	c.httpClient.CloseIdleConnections()
	return resp.LogOffResult, nil
}

type getServiceAttributes struct {
	XMLName xml.Name `xml:"http://www.daisy.org/ns/daisy-online/ getServiceAttributes"`
}

type getServiceAttributesResponse struct {
	XMLName           xml.Name          `xml:"http://www.daisy.org/ns/daisy-online/ getServiceAttributesResponse"`
	ServiceAttributes ServiceAttributes `xml:"serviceAttributes"`
}

// Retrieves Service properties, including information on which optional Operations the Service supports.
// A Reading System must call this operation as part of the Session Initialization Sequence and may call the operation to retrieve information on possible changes to Service properties at any other time during a Session.
func (c *Client) GetServiceAttributes() (*ServiceAttributes, error) {
	action := "getServiceAttributes"
	req := getServiceAttributes{}
	resp := getServiceAttributesResponse{}
	if err := c.call(action, req, &resp); err != nil {
		return nil, fmt.Errorf("%v operation: %w", action, err)
	}
	return &resp.ServiceAttributes, nil
}

type setReadingSystemAttributes struct {
	XMLName                 xml.Name                 `xml:"http://www.daisy.org/ns/daisy-online/ setReadingSystemAttributes"`
	ReadingSystemAttributes *ReadingSystemAttributes `xml:"readingSystemAttributes"`
}

type setReadingSystemAttributesResponse struct {
	XMLName                          xml.Name `xml:"http://www.daisy.org/ns/daisy-online/ setReadingSystemAttributesResponse"`
	SetReadingSystemAttributesResult bool     `xml:"setReadingSystemAttributesResult"`
}

// Sends Reading System properties to a Service.
// A Reading System must call this operation as part of the Session Initialization Sequence. The operation may be called additional times during a Session to record dynamic changes in a Reading System's properties.
func (c *Client) SetReadingSystemAttributes(readingSystemAttributes *ReadingSystemAttributes) (bool, error) {
	action := "setReadingSystemAttributes"
	req := setReadingSystemAttributes{ReadingSystemAttributes: readingSystemAttributes}
	resp := setReadingSystemAttributesResponse{}
	if err := c.call(action, req, &resp); err != nil {
		return false, fmt.Errorf("%v operation: %w", action, err)
	}
	return resp.SetReadingSystemAttributesResult, nil
}

type getContentList struct {
	XMLName   xml.Name `xml:"http://www.daisy.org/ns/daisy-online/ getContentList"`
	ID        string   `xml:"id"`
	FirstItem int32    `xml:"firstItem"`
	LastItem  int32    `xml:"lastItem"`
}

type getContentListResponse struct {
	XMLName     xml.Name    `xml:"http://www.daisy.org/ns/daisy-online/ getContentListResponse"`
	ContentList ContentList `xml:"contentList"`
}

// Retrieves a list of Content items.
// The list returned by the Service can be pre-composed, in which case it is retrieved by passing one of the three reserved values defined in the id parameter below. (Refer to 4, Protocol Fundamentals for information on the contexts in which these reserved values are used.)
// The list can also be dynamic (e.g., the result of a dynamic menu search operation sequence). In this case, the id value used to refer to the list is provided in the return value of a previous call to getQuestions. (Refer to the questions type for more information.)
func (c *Client) GetContentList(id string, firstItem int32, lastItem int32) (*ContentList, error) {
	action := "getContentList"
	req := getContentList{
		ID:        id,
		FirstItem: firstItem,
		LastItem:  lastItem,
	}
	resp := getContentListResponse{}
	if err := c.call(action, req, &resp); err != nil {
		return nil, fmt.Errorf("%v operation: %w", action, err)
	}
	return &resp.ContentList, nil
}

type getContentMetadata struct {
	XMLName   xml.Name `xml:"http://www.daisy.org/ns/daisy-online/ getContentMetadata"`
	ContentID string   `xml:"contentID"`
}

type getContentMetadataResponse struct {
	XMLName         xml.Name        `xml:"http://www.daisy.org/ns/daisy-online/ getContentMetadataResponse"`
	ContentMetadata ContentMetadata `xml:"contentMetadata"`
}

// Retrieves the contentMetadata of the specified Content item.
// This operation must be called as part of the Content Retrieval Sequence.
func (c *Client) GetContentMetadata(contentID string) (*ContentMetadata, error) {
	action := "getContentMetadata"
	req := getContentMetadata{ContentID: contentID}
	resp := getContentMetadataResponse{}
	if err := c.call(action, req, &resp); err != nil {
		return nil, fmt.Errorf("%v operation: %w", action, err)
	}
	return &resp.ContentMetadata, nil
}

type getContentResources struct {
	XMLName   xml.Name `xml:"http://www.daisy.org/ns/daisy-online/ getContentResources"`
	ContentID string   `xml:"contentID"`
}

type getContentResourcesResponse struct {
	XMLName   xml.Name  `xml:"http://www.daisy.org/ns/daisy-online/ getContentResourcesResponse"`
	Resources Resources `xml:"resources"`
}

// Retrieves the resources list for the specified Content item.
// The Content item must be issued before this operation is called. If not, the Service shall respond with an invalidParameter Fault.
func (c *Client) GetContentResources(contentID string) (*Resources, error) {
	action := "getContentResources"
	req := getContentResources{ContentID: contentID}
	resp := getContentResourcesResponse{}
	if err := c.call(action, req, &resp); err != nil {
		return nil, fmt.Errorf("%v operation: %w", action, err)
	}
	return &resp.Resources, nil
}

type issueContent struct {
	XMLName   xml.Name `xml:"http://www.daisy.org/ns/daisy-online/ issueContent"`
	ContentID string   `xml:"contentID"`
}

type issueContentResponse struct {
	XMLName            xml.Name `xml:"http://www.daisy.org/ns/daisy-online/ issueContentResponse"`
	IssueContentResult bool     `xml:"issueContentResult"`
}

// Requests a Service to issue the specified Content item.
func (c *Client) IssueContent(contentID string) (bool, error) {
	action := "issueContent"
	req := issueContent{ContentID: contentID}
	resp := issueContentResponse{}
	if err := c.call(action, req, &resp); err != nil {
		return false, fmt.Errorf("%v operation: %w", action, err)
	}
	return resp.IssueContentResult, nil
}

type returnContent struct {
	XMLName   xml.Name `xml:"http://www.daisy.org/ns/daisy-online/ returnContent"`
	ContentID string   `xml:"contentID"`
}

type returnContentResponse struct {
	XMLName             xml.Name `xml:"http://www.daisy.org/ns/daisy-online/ returnContentResponse"`
	ReturnContentResult bool     `xml:"returnContentResult"`
}

// Notifies the Service that the specified Content item has been deleted from the Reading System.
// The specified Content item is no longer issued to the User after a successful call to this operation.
// A Reading System must not call this function for a Content item that has a requiresReturn attribute with a value of false.
// A Reading System must delete the Content item before calling returnContent. A Reading System must not call returnContent for a Content item that was not issued to the User on that Reading System.
func (c *Client) ReturnContent(contentID string) (bool, error) {
	action := "returnContent"
	req := returnContent{ContentID: contentID}
	resp := returnContentResponse{}
	if err := c.call(action, req, &resp); err != nil {
		return false, fmt.Errorf("%v operation: %w", action, err)
	}
	return resp.ReturnContentResult, nil
}

type getQuestions struct {
	XMLName       xml.Name       `xml:"http://www.daisy.org/ns/daisy-online/ getQuestions"`
	UserResponses *UserResponses `xml:"userResponses"`
}

type getQuestionsResponse struct {
	XMLName   xml.Name  `xml:"http://www.daisy.org/ns/daisy-online/ getQuestionsResponse"`
	Questions Questions `xml:"questions"`
}

// Retrieves a question from the series of questions that comprise the dynamic menu system.
func (c *Client) GetQuestions(userResponses *UserResponses) (*Questions, error) {
	action := "getQuestions"
	req := getQuestions{UserResponses: userResponses}
	resp := getQuestionsResponse{}
	if err := c.call(action, req, &resp); err != nil {
		return nil, fmt.Errorf("%v operation: %w", action, err)
	}
	return &resp.Questions, nil
}

type getServiceAnnouncements struct {
	XMLName xml.Name `xml:"http://www.daisy.org/ns/daisy-online/ getServiceAnnouncements"`
}

type getServiceAnnouncementsResponse struct {
	XMLName       xml.Name      `xml:"http://www.daisy.org/ns/daisy-online/ getServiceAnnouncementsResponse"`
	Announcements Announcements `xml:"announcements"`
}

// Retrieves any announcements from the Service that a User has not yet read.
func (c *Client) GetServiceAnnouncements() (*Announcements, error) {
	action := "getServiceAnnouncements"
	req := getServiceAnnouncements{}
	resp := getServiceAnnouncementsResponse{}
	if err := c.call(action, req, &resp); err != nil {
		return nil, fmt.Errorf("%v operation: %w", action, err)
	}
	return &resp.Announcements, nil
}

type setBookmarks struct {
	XMLName     xml.Name     `xml:"http://www.daisy.org/ns/daisy-online/ setBookmarks"`
	ContentID   string       `xml:"contentID"`
	BookmarkSet *BookmarkSet `xml:"bookmarkSet"`
}

type setBookmarksResponse struct {
	XMLName            xml.Name `xml:"http://www.daisy.org/ns/daisy-online/ setBookmarksResponse"`
	SetBookmarksResult bool     `xml:"setBookmarksResult"`
}

// Requests that a Service store the supplied bookmarks for a Content item.
// This operation only supports the storage of bookmarks for one Content item at a time.
func (c *Client) SetBookmarks(contentID string, bookmarkSet *BookmarkSet) (bool, error) {
	action := "setBookmarks"
	req := setBookmarks{
		ContentID:   contentID,
		BookmarkSet: bookmarkSet,
	}
	resp := setBookmarksResponse{}
	if err := c.call(action, req, &resp); err != nil {
		return false, fmt.Errorf("%v operation: %w", action, err)
	}
	return resp.SetBookmarksResult, nil
}

type getBookmarks struct {
	XMLName   xml.Name `xml:"http://www.daisy.org/ns/daisy-online/ getBookmarks"`
	ContentID string   `xml:"contentID"`
}

type getBookmarksResponse struct {
	XMLName     xml.Name    `xml:"http://www.daisy.org/ns/daisy-online/ getBookmarksResponse"`
	BookmarkSet BookmarkSet `xml:"bookmarkSet"`
}

// Retrieves the bookmarks for a Content item from a Service.
func (c *Client) GetBookmarks(contentID string) (*BookmarkSet, error) {
	action := "getBookmarks"
	req := getBookmarks{ContentID: contentID}
	resp := getBookmarksResponse{}
	if err := c.call(action, req, &resp); err != nil {
		return nil, fmt.Errorf("%v operation: %w", action, err)
	}
	return &resp.BookmarkSet, nil
}

type markAnnouncementsAsRead struct {
	XMLName xml.Name `xml:"http://www.daisy.org/ns/daisy-online/ markAnnouncementsAsRead"`
	Read    *Read    `xml:"read"`
}

type markAnnouncementsAsReadResponse struct {
	XMLName                       xml.Name `xml:"http://www.daisy.org/ns/daisy-online/ markAnnouncementsAsReadResponse"`
	MarkAnnouncementsAsReadResult bool     `xml:"markAnnouncementsAsReadResult"`
}

// Marks the specified announcement(s) as read.
// This operation is only valid if a previous call to  getServiceAnnouncements  has been made during the Session.
func (c *Client) MarkAnnouncementsAsRead(read *Read) (bool, error) {
	action := "markAnnouncementsAsRead"
	req := markAnnouncementsAsRead{Read: read}
	resp := markAnnouncementsAsReadResponse{}
	if err := c.call(action, req, &resp); err != nil {
		return false, fmt.Errorf("%v operation: %w", action, err)
	}
	return resp.MarkAnnouncementsAsReadResult, nil
}
