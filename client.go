package daisyonline

import (
	"bytes"
	"encoding/xml"
	"errors"
	"net/http"
	"net/http/cookiejar"
)

// Supported input types
const (
	TEXT_NUMERIC = "TEXT_NUMERIC"
	TEXT_ALPHANUMERIC = "TEXT_ALPHANUMERIC"
	AUDIO = "AUDIO"
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

type GetServiceAnnouncements struct {
	XMLName xml.Name `xml:"getServiceAnnouncements"`
}

type GetServiceAnnouncementsResponse struct {
	XMLName       xml.Name      `xml:"getServiceAnnouncementsResponse"`
	Announcements Announcements `xml:"announcements"`
}

type GetQuestions struct {
	XMLName       xml.Name       `xml:"getQuestions"`
	UserResponses *UserResponses `xml:"userResponses"`
}

type GetQuestionsResponse struct {
	XMLName   xml.Name  `xml:"getQuestionsResponse"`
	Questions Questions `xml:"questions"`
}

type ReturnContent struct {
	XMLName   xml.Name `xml:"returnContent"`
	ContentID string   `xml:"contentID"`
}

type ReturnContentResponse struct {
	XMLName             xml.Name `xml:"returnContentResponse"`
	ReturnContentResult bool     `xml:"returnContentResult"`
}

type IssueContent struct {
	XMLName   xml.Name `xml:"issueContent"`
	ContentID string   `xml:"contentID"`
}

type IssueContentResponse struct {
	XMLName            xml.Name `xml:"issueContentResponse"`
	IssueContentResult bool     `xml:"issueContentResult"`
}

type GetContentResources struct {
	XMLName   xml.Name `xml:"getContentResources"`
	ContentID string   `xml:"contentID"`
}

type GetContentResourcesResponse struct {
	XMLName   xml.Name  `xml:"getContentResourcesResponse"`
	Resources Resources `xml:"resources"`
}

type GetContentMetadata struct {
	XMLName   xml.Name `xml:"getContentMetadata"`
	ContentID string   `xml:"contentID"`
}

type GetContentMetadataResponse struct {
	XMLName         xml.Name        `xml:"getContentMetadataResponse"`
	ContentMetadata ContentMetadata `xml:"contentMetadata"`
}

type GetContentList struct {
	XMLName   xml.Name `xml:"getContentList"`
	ID        string   `xml:"id"`
	FirstItem int      `xml:"firstItem"`
	LastItem  int      `xml:"lastItem"`
}

type GetContentListResponse struct {
	XMLName     xml.Name    `xml:"getContentListResponse"`
	ContentList ContentList `xml:"contentList"`
}

type LogOn struct {
	XMLName  xml.Name `xml:"logOn"`
	Username string   `xml:"username"`
	Password string   `xml:"password"`
}

type LogOnResponse struct {
	XMLName     xml.Name `xml:"logOnResponse"`
	LogOnResult bool     `xml:"logOnResult"`
}

type GetServiceAttributes struct {
	XMLName xml.Name `xml:"getServiceAttributes"`
}

type GetServiceAttributesResponse struct {
	XMLName           xml.Name          `xml:"getServiceAttributesResponse"`
	ServiceAttributes ServiceAttributes `xml:"serviceAttributes"`
}

type SetReadingSystemAttributes struct {
	XMLName                 xml.Name                 `xml:"setReadingSystemAttributes"`
	ReadingSystemAttributes *ReadingSystemAttributes `xml:"readingSystemAttributes"`
}

type SetReadingSystemAttributesResponse struct {
	XMLName                          xml.Name `xml:"setReadingSystemAttributesResponse"`
	SetReadingSystemAttributesResult bool     `xml:"setReadingSystemAttributesResult"`
}

type LogOff struct {
	XMLName xml.Name `xml:"logOff"`
}

type LogOffResponse struct {
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
	Body    envelopeBody
}

type envelopeBody struct {
	XMLName xml.Name
	Content interface{}
}

type Client struct {
	url        string
	httpClient http.Client
}

func NewClient(url string) *Client {
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic("Invalid cookie jar")
	}

	return &Client{
		url:        url,
		httpClient: http.Client{Jar: jar},
	}
}

// Performs the session initialization sequence
func Authentication(client *Client, readingSystemAttributes *ReadingSystemAttributes, username, password string) (*ServiceAttributes, error) {
	_, err := client.LogOn(username, password)
	if err != nil {
		return nil, err
	}

	serviceAttributes, err := client.GetServiceAttributes()
	if err != nil {
		return nil, err
	}

	_, err = client.SetReadingSystemAttributes(readingSystemAttributes)
	if err != nil {
		return nil, err
	}

	return serviceAttributes, nil
}

func (c *Client) call(method string, args, rs interface{}) error {
	env := envelope{
		NS: "http://schemas.xmlsoap.org/soap/envelope/",
		Body: envelopeBody{
			XMLName: xml.Name{"http://www.daisy.org/ns/daisy-online/", "SOAP-ENV:Body"},
			Content: args,
		}}

	XMLDoc, err := xml.Marshal(env)
	if err != nil {
		panic(err)
	}
	XMLDoc = append([]byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n"), XMLDoc...)

	req, err := http.NewRequest(http.MethodPost, c.url, bytes.NewReader(XMLDoc))
	if err != nil {
		return err
	}

	req.ContentLength = int64(len(XMLDoc))
	req.Header.Add("Content-Type", "text/xml; charset=utf-8")
	req.Header.Add("Accept", "text/xml")
	req.Header.Add("SOAPAction", "/"+method)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}

	d := xml.NewDecoder(resp.Body)
	envresp := envelopeResponse{}

	if err := d.Decode(&envresp); err != nil {
		return err
	}

	return xml.Unmarshal(envresp.Body.Content, rs)
}

// Logs a Reading System on to a Service.
func (c *Client) LogOn(username, password string) (bool, error) {
	req := LogOn{
		Username: username,
		Password: password,
	}

	resp := LogOnResponse{}
	if err := c.call("logOn", req, &resp); err != nil {
		return false, err
	}
	return resp.LogOnResult, nil
}

// Logs a Reading System off a Service.
// A return value of false or a Fault both indicate that the operation was not successful.
func (c *Client) LogOff() (bool, error) {
	req := LogOff{}
	resp := LogOffResponse{}
	if err := c.call("logOff", req, &resp); err != nil {
		return false, err
	}
	return resp.LogOffResult, nil
}

// Retrieves Service properties, including information on which optional Operations the Service supports.
// A Reading System must call this operation as part of the Session Initialization Sequence and may call the operation to retrieve information on possible changes to Service properties at any other time during a Session.
func (c *Client) GetServiceAttributes() (*ServiceAttributes, error) {
	req := GetServiceAttributes{}
	resp := GetServiceAttributesResponse{}
	if err := c.call("getServiceAttributes", req, &resp); err != nil {
		return nil, err
	}
	return &resp.ServiceAttributes, nil
}

// Sends Reading System properties to a Service.
// A Reading System must call this operation as part of the Session Initialization Sequence. The operation may be called additional times during a Session to record dynamic changes in a Reading System’s properties.
func (c *Client) SetReadingSystemAttributes(readingSystemAttributes *ReadingSystemAttributes) (bool, error) {
	req := SetReadingSystemAttributes{ReadingSystemAttributes: readingSystemAttributes}
	resp := SetReadingSystemAttributesResponse{}
	if err := c.call("setReadingSystemAttributes", req, &resp); err != nil {
		return false, err
	}
	return resp.SetReadingSystemAttributesResult, nil
}

// Retrieves a list of Content items.
// The list returned by the Service can be pre-composed, in which case it is retrieved by passing one of the three reserved values defined in the id parameter below. (Refer to 4, Protocol Fundamentals for information on the contexts in which these reserved values are used.)
// The list can also be dynamic (e.g., the result of a dynamic menu search operation sequence). In this case, the id value used to refer to the list is provided in the return value of a previous call to getQuestions. (Refer to the questions type for more information.)
func (c *Client) GetContentList(id string, firstItem int, lastItem int) (*ContentList, error) {
	req := GetContentList{
		ID:        id,
		FirstItem: firstItem,
		LastItem:  lastItem,
	}

	resp := GetContentListResponse{}
	if err := c.call("getContentList", req, &resp); err != nil {
		return nil, err
	}
	return &resp.ContentList, nil
}

// Retrieves the contentMetadata of the specified Content item.
// This operation must be called as part of the Content Retrieval Sequence.
func (c *Client) GetContentMetadata(contentID string) (*ContentMetadata, error) {
	req := GetContentMetadata{ContentID: contentID}
	resp := GetContentMetadataResponse{}
	if err := c.call("getContentMetadata", req, &resp); err != nil {
		return nil, err
	}
	return &resp.ContentMetadata, nil
}

// Retrieves the resources list for the specified Content item.
// The Content item must be issued before this operation is called. If not, the Service shall respond with an invalidParameter Fault.
func (c *Client) GetContentResources(contentID string) (*Resources, error) {
	req := GetContentResources{ContentID: contentID}
	resp := GetContentResourcesResponse{}
	if err := c.call("getContentResources", req, &resp); err != nil {
		return nil, err
	}
	return &resp.Resources, nil
}

// Requests a Service to issue the specified Content item.
func (c *Client) IssueContent(contentID string) (bool, error) {
	req := IssueContent{ContentID: contentID}
	resp := IssueContentResponse{}
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
	req := ReturnContent{ContentID: contentID}
	resp := ReturnContentResponse{}
	if err := c.call("returnContent", req, &resp); err != nil {
		return false, err
	}
	return resp.ReturnContentResult, nil
}

// Retrieves a question from the series of questions that comprise the dynamic menu system.
func (c *Client) GetQuestions(userResponses *UserResponses) (*Questions, error) {
	req := GetQuestions{UserResponses: userResponses}
	resp := GetQuestionsResponse{}
	if err := c.call("getQuestions", req, &resp); err != nil {
		return nil, err
	}
	return &resp.Questions, nil
}

// Retrieves any announcements from the Service that a User has not yet read.
func (c *Client) GetServiceAnnouncements() (*Announcements, error) {
	req := GetServiceAnnouncements{}
	resp := GetServiceAnnouncementsResponse{}
	if err := c.call("getServiceAnnouncements", req, &resp); err != nil {
		return nil, err
	}
	return &resp.Announcements, nil
}
