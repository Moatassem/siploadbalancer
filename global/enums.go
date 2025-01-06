package global

import "slices"

type Method int

const (
	UNKNOWN Method = iota
	INVITE
	ReINVITE
	REFER
	ACK
	CANCEL
	BYE
	OPTIONS
	NOTIFY
	UPDATE
	PRACK
	INFO
	REGISTER
	SUBSCRIBE
	MESSAGE
	PUBLISH
	NEGOTIATE
)

func (m Method) String() string {
	return methods[m]
}

func MethodFromName(nm string) Method {
	idx := slices.IndexFunc(methods[:], func(m string) bool { return m == nm })
	if idx == -1 {
		return UNKNOWN
	}
	return Method(idx)
}

// ==============================================================
type BodyType int

const (
	None BodyType = iota
	SDP
	DTMF
	DTMFRelay
	SIPFragment
	SimpleMsgSummary
	PlainText
	AppJson

	MultipartAlternative
	MultipartFormData
	MultipartMixed
	MultipartRelated

	ISUP
	QSIG

	PIDFXML
	MSCPXML
	MSCXML
	VndEtsiPstnXML
	VndOrangeInData
	ResourceListXML
	AnyXML
	Invalid
)

type Direction int

const (
	INBOUND Direction = iota
	OUTBOUND
)

func (d Direction) String() string {
	return directions[d]
}

// ==============================================================
type MessageType int

const (
	REQUEST MessageType = iota
	RESPONSE
	INVALID
)

func (mt MessageType) String() string {
	return messageTypes[mt]
}

type timeFormat int

const (
	Signaling timeFormat = iota
	Tracing
	Version
	DateOnly
	TimeOnly
	DateTimeOnly
	Session
	HTML
	DateTimeLocal
	JsonDateTime
	JsonDateTimeMS
	HTMLDateOnly
	SimpleDT
)

type FieldPattern int

const (
	NumberOnly FieldPattern = iota
	NameAndNumber
	ReplaceNumberOnly
	RequestStartLinePattern
	INVITERURI
	ResponseStartLinePattern
	ViaBranchPattern
	ViaTransport
	MediaPayloadTypes
	MediaPayloadTypeDefinition
	SDPOriginLine
	MediaDirective
	ConnectionAddress
	FullHeader
	MediaLine
	HostIPPort
	FQDNPort
	TransportProtocol
	ViaIPv4Socket
	IP6
	IP4
	HeaderParameter
	SDPTime
	Header
	URIFull
	URIParameters
	HTTPRequestStartLine
	URIParameter
	CSeqHeader
	HandlingProfile
	ErrorStack
	SDPALineReMapping
	SDPPTDefinition
	CodecPolicy
	ObjectName
	SignalDTMF
	DurationDTMF
	ConfDropPart
	Tag
	RAckHeader
)

type NewSessionType int

const (
	Unset NewSessionType = iota
	ValidRequest
	DuplicateRequest
	InvalidRequest
	UnsupportedURIScheme
	UnsupportedBody
	ForbiddenRequest
	WithRequireHeader
	Response
	UnExpectedMessage
	NoAllowedAudioCodecs
	TooLowMaxForwards
	RegistrarOff
	ServerOff
	EndpointNotRegistered
	ExceededRouteCAC
	DisposedSession
	CallLegTransactionNotExist
	RouteBlocked
	UCLimitReached
	DuplicateMessage
	RouteOutboundOnly
	RouteBlackhole
	ExceededCallRate
	UnknownEndPoint
)

// ==============================================================

type HeaderEnum int

const (
	Accept HeaderEnum = iota
	Accept_Contact
	Accept_Encoding
	Accept_Language
	Accept_Resource_Priority
	Alert_Info
	Allow
	Allow_Events
	Answer_Mode
	Authentication_Info
	Authorization
	Call_ID
	Call_Info
	Compression
	Contact
	Content_Disposition
	Content_Encoding
	Content_Language
	Content_Length
	Content_Type
	Cisco_Guid
	CSeq
	Custom_CLI
	Date
	Diversion
	Error_Info
	Event
	Expires
	Feature_Caps
	Flow_Timer
	From
	Geolocation
	Geolocation_Error
	Geolocation_Routing
	History_Info
	Identity
	Identity_Info
	In_Reply_To
	Info_Package
	Join
	Max_Breadth
	Max_Expires
	Max_Forwards
	MIME_Version
	Min_Expires
	Min_SE
	Organization
	P_Access_Network_Info
	P_Answer_State
	P_Asserted_Identity
	P_Asserted_Service
	P_Associated_URI
	P_Called_Party_ID
	P_Charging_Function_Addresses
	P_Charging_Vector
	P_DCS_Billing_Info
	P_DCS_LAES
	P_DCS_OSPS
	P_DCS_Redirect
	P_DCS_Trace_Party_ID
	P_Early_Media
	P_Media_Authorization
	P_Preferred_Identity
	P_Preferred_Service
	P_Private_Network_Indication
	P_Profile_Key
	P_Refused_URI_List
	P_Served_User
	P_User_Database
	P_Visited_Network_ID
	Path
	Permission_Missing
	Policy_Contact
	Policy_ID
	Priority
	Priv_Answer_Mode
	Privacy
	Proxy_Authenticate
	Proxy_Authorization
	Proxy_Require
	RAck
	Reason
	Reason_Phrase
	Record_Route
	Recv_Info
	Refer_Events_At
	Refer_Sub
	Refer_To
	Referred_By
	Reject_Contact
	Replaces
	Reply_To
	Request_Disposition
	Require
	Resource_Priority
	Retry_After
	Route
	RSeq
	Security_Client
	Security_Server
	Security_Verify
	Server
	Service_Route
	Session_Expires
	Session_ID
	SIP_ETag
	SIP_If_Match
	Subject
	Subscription_State
	Subscription_Expires
	Supported
	Suppress_If_Match
	Target_Dialog
	Timestamp
	To
	Trigger_Consent
	Unsupported
	User_Agent
	User_to_User
	Via
	Warning
	WWW_Authenticate
)
