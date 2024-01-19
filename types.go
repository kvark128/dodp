package dodp

import (
	"encoding/xml"
)

// Properties of the Service.
// The properties specified must be constant for the duration of the Session.
type ServiceAttributes struct {
	XMLName                          xml.Name `xml:"serviceAttributes"`
	ServiceProvider                  ServiceProvider
	Service                          Service
	SupportedContentSelectionMethods SupportedContentSelectionMethods
	SupportsServerSideBack           bool `xml:"supportsServerSideBack"`
	SupportsSearch                   bool `xml:"supportsSearch"`
	SupportedUplinkAudioCodecs       SupportedUplinkAudioCodecs
	SupportsAudioLabels              bool `xml:"supportsAudioLabels"`
	SupportedOptionalOperations      SupportedOptionalOperations
}

// Specifies which (if any) of the  optional  operations are supported by the Service.
type SupportedOptionalOperations struct {
	XMLName   xml.Name `xml:"supportedOptionalOperations"`
	Operation []string `xml:"operation"`
}

// A list of the audio codecs (if any) supported in  userResponses  in addition to  [ RIFF WAVE ] .
type SupportedUplinkAudioCodecs struct {
	XMLName xml.Name `xml:"supportedUplinkAudioCodecs"`
	Codec   []string `xml:"codec"`
}

// A list of  Content Selection Methods  supported by this Service. A Service must support at least one of the two methods.
type SupportedContentSelectionMethods struct {
	XMLName xml.Name `xml:"supportedContentSelectionMethods"`
	Method  []string `xml:"method"`
}

// The identity of the Service.
type Service struct {
	XMLName xml.Name `xml:"service"`
	ID      string   `xml:"id,attr"`
	Label   Label
}

// The identity of the Service Provider.
type ServiceProvider struct {
	XMLName xml.Name `xml:"serviceProvider"`
	ID      string   `xml:"id,attr"`
	Label   Label
}

// A multi-purpose label, containing text and optionally audio.
// To achieve maximum interoperability, Services should support the provision of audio labels, as Reading Systems may require them in order to render Service messages to the user.
type Label struct {
	XMLName xml.Name `xml:"label"`
	Lang    string   `xml:"lang,attr"`
	Dir     string   `xml:"dir,attr"`
	Text    string   `xml:"text"`
	Audio   Audio
}

type Audio struct {
	XMLName    xml.Name `xml:"audio"`
	URI        string   `xml:"uri,attr"`
	RangeBegin int64    `xml:"rangeBegin,attr"`
	RangeEnd   int64    `xml:"rangeEnd,attr"`
	Size       int64    `xml:"size,attr"`
}

// Specifies Reading System properties.
// The properties specified are valid until the end of the Session.
type ReadingSystemAttributes struct {
	XMLName      xml.Name `xml:"readingSystemAttributes"`
	Manufacturer string   `xml:"manufacturer"`
	Model        string   `xml:"model"`
	Version      string   `xml:"version"`
	Config       Config
}

type Config struct {
	XMLName                           xml.Name `xml:"config"`
	SupportsMultipleSelections        bool     `xml:"supportsMultipleSelections"`
	PreferredUILanguage               string   `xml:"preferredUILanguage"`
	SupportedContentFormats           SupportedContentFormats
	SupportedContentProtectionFormats SupportedContentProtectionFormats
	SupportedMimeTypes                SupportedMimeTypes
	SupportedInputTypes               SupportedInputTypes
	RequiresAudioLabels               bool `xml:"requiresAudioLabels"`
}

// Specifies which Content protection (Digital Rights Management) standards the Reading System supports, if any.
type SupportedContentProtectionFormats struct {
	XMLName          xml.Name `xml:"supportedContentProtectionFormats"`
	ProtectionFormat []string `xml:"protectionFormat"`
}

// Specifies which Content formats the Reading System supports. A Service may use this information to choose which formats to offer to the Reading System. This document does not specify the behavior of the Service if this list is empty.
type SupportedContentFormats struct {
	XMLName       xml.Name `xml:"supportedContentFormats"`
	ContentFormat []string `xml:"contentFormat"`
}

type SupportedMimeTypes struct {
	XMLName  xml.Name   `xml:"supportedMimeTypes"`
	MimeType []MimeType `xml:"mimeType"`
}

type MimeType struct {
	XMLName xml.Name `xml:"mimeType"`
	Type    string   `xml:"type,attr"`
}

type SupportedInputTypes struct {
	XMLName xml.Name `xml:"supportedInputTypes"`
	Input   []Input  `xml:"input"`
}

type Input struct {
	XMLName xml.Name `xml:"input"`
	Type    string   `xml:"type,attr"`
}

type ContentList struct {
	XMLName      xml.Name `xml:"contentList"`
	TotalItems   int32    `xml:"totalItems,attr"`
	FirstItem    int32    `xml:"firstItem,attr"`
	LastItem     int32    `xml:"lastItem,attr"`
	ID           string   `xml:"id,attr"`
	Label        Label
	ContentItems []ContentItem `xml:"contentItem"`
}

type ContentItem struct {
	XMLName          xml.Name `xml:"contentItem"`
	ID               string   `xml:"id,attr"`
	LastModifiedDate string   `xml:"lastModifiedDate,attr"`
	Label            Label
}

type ContentMetadata struct {
	XMLName        xml.Name `xml:"contentMetadata"`
	Category       string   `xml:"category,attr"`
	RequiresReturn bool     `xml:"requiresReturn,attr"`
	Sample         Sample
	Metadata       Metadata
}

type Metadata struct {
	XMLName     xml.Name `xml:"metadata"`
	Title       string   `xml:"title"`
	Identifier  string   `xml:"identifier"`
	Publisher   string   `xml:"publisher"`
	Format      string   `xml:"format"`
	Date        string   `xml:"date"`
	Source      string   `xml:"source"`
	Type        []string `xml:"type"`
	Subject     []string `xml:"subject"`
	Rights      []string `xml:"rights"`
	Relation    []string `xml:"relation"`
	Language    []string `xml:"language"`
	Description []string `xml:"description"`
	Creator     []string `xml:"creator"`
	Coverage    []string `xml:"coverage"`
	Contributor []string `xml:"contributor"`
	Narrator    []string `xml:"narrator"`
	Size        int64    `xml:"size"`
	Meta        []Meta   `xml:"meta"`
}

type Meta struct {
	XMLName xml.Name `xml:"meta"`
	Name    string   `xml:"name,attr"`
	Content string   `xml:"content,attr"`
}

// A sample of the Content item that the User may retrieve without the Content item being issued.
type Sample struct {
	XMLName xml.Name `xml:"sample"`
	ID      string   `xml:"id,attr"`
}

// A list of all the resources that constitute the Content item.
type Resources struct {
	XMLName          xml.Name   `xml:"resources"`
	ReturnBy         string     `xml:"returnBy,attr"`
	LastModifiedDate string     `xml:"lastModifiedDate,attr"`
	Resources        []Resource `xml:"resource"`
}

type Resource struct {
	XMLName          xml.Name `xml:"resource"`
	URI              string   `xml:"uri,attr"`
	MimeType         string   `xml:"mimeType,attr"`
	Size             int64    `xml:"size,attr"`
	LocalURI         string   `xml:"localURI,attr"`
	LastModifiedDate string   `xml:"lastModifiedDate,attr"`
}

// A set of User responses to  questions  provided by the Service.
type UserResponses struct {
	XMLName      xml.Name       `xml:"userResponses"`
	UserResponse []UserResponse `xml:"userResponse"`
}

type UserResponse struct {
	XMLName    xml.Name `xml:"userResponse"`
	QuestionID string   `xml:"questionID,attr"`
	Value      string   `xml:"value,attr,omitempty"`
	Data       string   `xml:"data,omitempty"`
}

type Questions struct {
	XMLName                xml.Name                 `xml:"questions"`
	MultipleChoiceQuestion []MultipleChoiceQuestion `xml:"multipleChoiceQuestion"`
	InputQuestion          []InputQuestion          `xml:"inputQuestion"`
	ContentListRef         string                   `xml:"contentListRef"`
	Label                  Label
}

type MultipleChoiceQuestion struct {
	XMLName                 xml.Name `xml:"multipleChoiceQuestion"`
	ID                      string   `xml:"id,attr"`
	AllowMultipleSelections bool     `xml:"allowMultipleSelections,attr"`
	Label                   Label
	Choices                 Choices
}

type Choices struct {
	XMLName xml.Name `xml:"choices"`
	Choice  []Choice `xml:"choice"`
}

type Choice struct {
	XMLName xml.Name `xml:"choice"`
	ID      string   `xml:"id,attr"`
	Label   Label
}

type InputQuestion struct {
	XMLName    xml.Name `xml:"inputQuestion"`
	ID         string   `xml:"id,attr"`
	InputTypes InputTypes
	Label      Label
}

type InputTypes struct {
	XMLName xml.Name `xml:"inputTypes"`
	Input   []Input  `xml:"input"`
}

type Announcements struct {
	XMLName      xml.Name       `xml:"announcements"`
	Announcement []Announcement `xml:"announcement"`
}

type Announcement struct {
	XMLName  xml.Name `xml:"announcement"`
	ID       string   `xml:"id,attr"`
	Type     string   `xml:"type,attr"`
	Priority int32    `xml:"priority,attr"`
	Label    Label
}
