package global

type Method string

const (
	UNKNOWN   Method = "N/A"
	INVITE    Method = "INVITE"
	ReINVITE  Method = "INVITE"
	REFER     Method = "REFER"
	ACK       Method = "ACK"
	CANCEL    Method = "CANCEL"
	BYE       Method = "BYE"
	OPTIONS   Method = "OPTIONS"
	NOTIFY    Method = "NOTIFY"
	UPDATE    Method = "UPDATE"
	PRACK     Method = "PRACK"
	INFO      Method = "INFO"
	REGISTER  Method = "REGISTER"
	SUBSCRIBE Method = "SUBSCRIBE"
	MESSAGE   Method = "MESSAGE"
	PUBLISH   Method = "PUBLISH"
	NEGOTIATE Method = "NEGOTIATE"
)

func GetMethod(hdrnm string) Method {
	switch hdrnm {
	case "INVITE": //ReINVITE included
		return INVITE
	case "REFER":
		return REFER
	case "ACK":
		return ACK
	case "CANCEL":
		return CANCEL
	case "BYE":
		return BYE
	case "OPTIONS":
		return OPTIONS
	case "NOTIFY":
		return NOTIFY
	case "UPDATE":
		return UPDATE
	case "PRACK":
		return PRACK
	case "INFO":
		return INFO
	case "REGISTER":
		return REGISTER
	case "SUBSCRIBE":
		return SUBSCRIBE
	case "MESSAGE":
		return MESSAGE
	case "PUBLISH":
		return PUBLISH
	case "NEGOTIATE":
		return NEGOTIATE
	default:
		return UNKNOWN
	}
}

// ==============================================================
type MessageType = string

const (
	REQUEST  MessageType = "REQUEST"
	RESPONSE MessageType = "RESPONSE"
	INVALID  MessageType = "INVALID"
)

type FieldPattern int

const (
	RequestStartLinePattern FieldPattern = iota
	INVITERURI
	ResponseStartLinePattern
	ViaBranchPattern
	FullHeader
	ViaIPv4Socket
	IP6
	IP4
	URIFull
	URIParameters
	URIParameter
	ErrorStack
	Tag
)

type Header = string

const (
	Call_ID        Header = "Call-ID"
	Content_Length Header = "Content-Length"
	From           Header = "From"
	To             Header = "To"
	Via            Header = "Via"
	Server         Header = "Server"
	CSeq           Header = "CSeq"
	Max_Forwards   Header = "Max-Forwards"
	Contact        Header = "Contact"
	User_Agent     Header = "User-Agent"
)
