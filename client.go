package dodp

import (
	"bytes"
	"compress/gzip"
	"encoding/xml"
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

type getServiceAnnouncements struct {
	XMLName xml.Name `xml:"getServiceAnnouncements"`
}

type getServiceAnnouncementsResponse struct {
	XMLName       xml.Name      `xml:"getServiceAnnouncementsResponse"`
	Announcements Announcements `xml:"announcements"`
}

type getQuestions struct {
	XMLName       xml.Name       `xml:"getQuestions"`
	UserResponses *UserResponses `xml:"userResponses"`
}

type getQuestionsResponse struct {
	XMLName   xml.Name  `xml:"getQuestionsResponse"`
	Questions Questions `xml:"questions"`
}

type returnContent struct {
	XMLName   xml.Name `xml:"returnContent"`
	ContentID string   `xml:"contentID"`
}

type returnContentResponse struct {
	XMLName             xml.Name `xml:"returnContentResponse"`
	ReturnContentResult bool     `xml:"returnContentResult"`
}

type issueContent struct {
	XMLName   xml.Name `xml:"issueContent"`
	ContentID string   `xml:"contentID"`
}

type issueContentResponse struct {
	XMLName            xml.Name `xml:"issueContentResponse"`
	IssueContentResult bool     `xml:"issueContentResult"`
}

type getContentResources struct {
	XMLName   xml.Name `xml:"getContentResources"`
	ContentID string   `xml:"contentID"`
}

type getContentResourcesResponse struct {
	XMLName   xml.Name  `xml:"getContentResourcesResponse"`
	Resources Resources `xml:"resources"`
}

type getContentMetadata struct {
	XMLName   xml.Name `xml:"getContentMetadata"`
	ContentID string   `xml:"contentID"`
}

type getContentMetadataResponse struct {
	XMLName         xml.Name        `xml:"getContentMetadataResponse"`
	ContentMetadata ContentMetadata `xml:"contentMetadata"`
}

type getContentList struct {
	XMLName   xml.Name `xml:"getContentList"`
	ID        string   `xml:"id"`
	FirstItem int      `xml:"firstItem"`
	LastItem  int      `xml:"lastItem"`
}

type getContentListResponse struct {
	XMLName     xml.Name    `xml:"getContentListResponse"`
	ContentList ContentList `xml:"contentList"`
}

type logOn struct {
	XMLName  xml.Name `xml:"logOn"`
	Username string   `xml:"username"`
	Password string   `xml:"password"`
}

type logOnResponse struct {
	XMLName     xml.Name `xml:"logOnResponse"`
	LogOnResult bool     `xml:"logOnResult"`
}

type getServiceAttributes struct {
	XMLName xml.Name `xml:"getServiceAttributes"`
}

type getServiceAttributesResponse struct {
	XMLName           xml.Name          `xml:"getServiceAttributesResponse"`
	ServiceAttributes ServiceAttributes `xml:"serviceAttributes"`
}

type setReadingSystemAttributes struct {
	XMLName                 xml.Name                 `xml:"setReadingSystemAttributes"`
	ReadingSystemAttributes *ReadingSystemAttributes `xml:"readingSystemAttributes"`
}

type setReadingSystemAttributesResponse struct {
	XMLName                          xml.Name `xml:"setReadingSystemAttributesResponse"`
	SetReadingSystemAttributesResult bool     `xml:"setReadingSystemAttributesResult"`
}

type logOff struct {
	XMLName xml.Name `xml:"logOff"`
}

type logOffResponse struct {
	XMLName      xml.Name `xml:"logOffResponse"`
	LogOffResult bool     `xml:"logOffResult"`
}

type envelopeResponse struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		XMLName xml.Name `xml:"Body"`
		Content []byte   `xml:",innerxml"`
	}
}

type envelope struct {
	XMLName xml.Name `xml:"SOAP-ENV:Envelope"`
	NS      string   `xml:"xmlns:SOAP-ENV,attr"`
	Body    body
}

type body struct {
	XMLName xml.Name `xml:"SOAP-ENV:Body"`
	NS      string   `xml:"xmlns,attr"`
	Content any
}

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
}

// Creates an instance of a new DAISY Online client with the specified service URL.
// Timeout limits the execution time of each HTTP request for this client.
// Zero timeout means no timeout.
func NewClient(url string, timeout time.Duration) *Client {
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
	}
}

func (c *Client) call(method string, args any, rs any) error {
	env := envelope{
		NS: "http://schemas.xmlsoap.org/soap/envelope/",
		Body: body{
			NS:      "http://www.daisy.org/ns/daisy-online/",
			Content: args,
		},
	}

	buf := bytes.NewBufferString(xml.Header)
	if err := xml.NewEncoder(buf).Encode(env); err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, c.url, buf)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "text/xml; charset=utf-8")
	req.Header.Add("Accept", "text/xml")
	req.Header.Add("Accept-Encoding", "gzip")
	req.Header.Add("SOAPAction", "/"+method)

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

	var envresp envelopeResponse
	if err := xml.NewDecoder(reader).Decode(&envresp); err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		fault := &Fault{}
		if err := xml.Unmarshal(envresp.Body.Content, fault); err != nil {
			return err
		}
		return fault
	}
	return xml.Unmarshal(envresp.Body.Content, rs)
}

// Logs a Reading System on to a Service.
func (c *Client) LogOn(username, password string) (bool, error) {
	req := logOn{
		Username: username,
		Password: password,
	}

	resp := logOnResponse{}
	if err := c.call("logOn", req, &resp); err != nil {
		return false, err
	}
	return resp.LogOnResult, nil
}

// Logs a Reading System off a Service.
// A return value of false or a Fault both indicate that the operation was not successful.
func (c *Client) LogOff() (bool, error) {
	req := logOff{}
	resp := logOffResponse{}
	if err := c.call("logOff", req, &resp); err != nil {
		return false, err
	}
	c.httpClient.CloseIdleConnections()
	return resp.LogOffResult, nil
}

// Retrieves Service properties, including information on which optional Operations the Service supports.
// A Reading System must call this operation as part of the Session Initialization Sequence and may call the operation to retrieve information on possible changes to Service properties at any other time during a Session.
func (c *Client) GetServiceAttributes() (*ServiceAttributes, error) {
	req := getServiceAttributes{}
	resp := getServiceAttributesResponse{}
	if err := c.call("getServiceAttributes", req, &resp); err != nil {
		return nil, err
	}
	return &resp.ServiceAttributes, nil
}

// Sends Reading System properties to a Service.
// A Reading System must call this operation as part of the Session Initialization Sequence. The operation may be called additional times during a Session to record dynamic changes in a Reading System’s properties.
func (c *Client) SetReadingSystemAttributes(readingSystemAttributes *ReadingSystemAttributes) (bool, error) {
	req := setReadingSystemAttributes{ReadingSystemAttributes: readingSystemAttributes}
	resp := setReadingSystemAttributesResponse{}
	if err := c.call("setReadingSystemAttributes", req, &resp); err != nil {
		return false, err
	}
	return resp.SetReadingSystemAttributesResult, nil
}

// Retrieves a list of Content items.
// The list returned by the Service can be pre-composed, in which case it is retrieved by passing one of the three reserved values defined in the id parameter below. (Refer to 4, Protocol Fundamentals for information on the contexts in which these reserved values are used.)
// The list can also be dynamic (e.g., the result of a dynamic menu search operation sequence). In this case, the id value used to refer to the list is provided in the return value of a previous call to getQuestions. (Refer to the questions type for more information.)
func (c *Client) GetContentList(id string, firstItem int, lastItem int) (*ContentList, error) {
	req := getContentList{
		ID:        id,
		FirstItem: firstItem,
		LastItem:  lastItem,
	}

	resp := getContentListResponse{}
	if err := c.call("getContentList", req, &resp); err != nil {
		return nil, err
	}
	return &resp.ContentList, nil
}

// Retrieves the contentMetadata of the specified Content item.
// This operation must be called as part of the Content Retrieval Sequence.
func (c *Client) GetContentMetadata(contentID string) (*ContentMetadata, error) {
	req := getContentMetadata{ContentID: contentID}
	resp := getContentMetadataResponse{}
	if err := c.call("getContentMetadata", req, &resp); err != nil {
		return nil, err
	}
	return &resp.ContentMetadata, nil
}

// Retrieves the resources list for the specified Content item.
// The Content item must be issued before this operation is called. If not, the Service shall respond with an invalidParameter Fault.
func (c *Client) GetContentResources(contentID string) (*Resources, error) {
	req := getContentResources{ContentID: contentID}
	resp := getContentResourcesResponse{}
	if err := c.call("getContentResources", req, &resp); err != nil {
		return nil, err
	}
	return &resp.Resources, nil
}

// Requests a Service to issue the specified Content item.
func (c *Client) IssueContent(contentID string) (bool, error) {
	req := issueContent{ContentID: contentID}
	resp := issueContentResponse{}
	if err := c.call("issueContent", req, &resp); err != nil {
		return false, err
	}
	return resp.IssueContentResult, nil
}

// Notifies the Service that the specified Content item has been deleted from the Reading System.
// The specified Content item is no longer issued to the User after a successful call to this operation.
// A Reading System must not call this function for a Content item that has a requiresReturn attribute with a value of false.
// A Reading System must delete the Content item before calling returnContent. A Reading System must not call returnContent for a Content item that was not issued to the User on that Reading System.
func (c *Client) ReturnContent(contentID string) (bool, error) {
	req := returnContent{ContentID: contentID}
	resp := returnContentResponse{}
	if err := c.call("returnContent", req, &resp); err != nil {
		return false, err
	}
	return resp.ReturnContentResult, nil
}

// Retrieves a question from the series of questions that comprise the dynamic menu system.
func (c *Client) GetQuestions(userResponses *UserResponses) (*Questions, error) {
	req := getQuestions{UserResponses: userResponses}
	resp := getQuestionsResponse{}
	if err := c.call("getQuestions", req, &resp); err != nil {
		return nil, err
	}
	return &resp.Questions, nil
}

// Retrieves any announcements from the Service that a User has not yet read.
func (c *Client) GetServiceAnnouncements() (*Announcements, error) {
	req := getServiceAnnouncements{}
	resp := getServiceAnnouncementsResponse{}
	if err := c.call("getServiceAnnouncements", req, &resp); err != nil {
		return nil, err
	}
	return &resp.Announcements, nil
}
