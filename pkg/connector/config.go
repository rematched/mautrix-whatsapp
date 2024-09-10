package connector

import (
	_ "embed"
	"strings"
	"text/template"

	up "go.mau.fi/util/configupgrade"
	"go.mau.fi/whatsmeow/types"
	"gopkg.in/yaml.v3"
	"maunium.net/go/mautrix/event"
)

type MediaRequestMethod string

const (
	MediaRequestMethodImmediate MediaRequestMethod = "immediate"
	MediaRequestMethodLocalTime MediaRequestMethod = "local_time"
)

//go:embed example-config.yaml
var ExampleConfig string

type Config struct {
	OSName      string `yaml:"os_name"`
	BrowserName string `yaml:"browser_name"`

	Proxy          string `yaml:"proxy"`
	GetProxyURL    string `yaml:"get_proxy_url"`
	ProxyOnlyLogin bool   `yaml:"proxy_only_login"`

	DisplaynameTemplate string `yaml:"displayname_template"`

	CallStartNotices            bool          `yaml:"call_start_notices"`
	SendPresenceOnTyping        bool          `yaml:"send_presence_on_typing"`
	EnableStatusBroadcast       bool          `yaml:"enable_status_broadcast"`
	DisableStatusBroadcastSend  bool          `yaml:"disable_status_broadcast_send"`
	MuteStatusBroadcast         bool          `yaml:"mute_status_broadcast"`
	StatusBroadcastTag          event.RoomTag `yaml:"status_broadcast_tag"`
	WhatsappThumbnail           bool          `yaml:"whatsapp_thumbnail"`
	URLPreviews                 bool          `yaml:"url_previews"`
	ForceActiveDeliveryReceipts bool          `yaml:"force_active_delivery_receipts"`

	HistorySync struct {
		RequestFullSync bool `yaml:"request_full_sync"`
		FullSyncConfig  struct {
			DaysLimit    uint32 `yaml:"days_limit"`
			SizeLimit    uint32 `yaml:"size_mb_limit"`
			StorageQuota uint32 `yaml:"storage_quota_mb"`
		} `yaml:"full_sync_config"`

		MediaRequests struct {
			AutoRequestMedia bool               `yaml:"auto_request_media"`
			RequestMethod    MediaRequestMethod `yaml:"request_method"`
			RequestLocalTime int                `yaml:"request_local_time"`
			MaxAsyncHandle   int64              `yaml:"max_async_handle"`
		} `yaml:"media_requests"`
	} `yaml:"history_sync"`

	displaynameTemplate *template.Template `yaml:"-"`
}

type umConfig Config

func (c *Config) UnmarshalYAML(node *yaml.Node) error {
	err := node.Decode((*umConfig)(c))
	if err != nil {
		return err
	}

	c.displaynameTemplate, err = template.New("displayname").Parse(c.DisplaynameTemplate)
	if err != nil {
		return err
	}
	return nil
}

func upgradeConfig(helper up.Helper) {
	helper.Copy(up.Str, "os_name")
	helper.Copy(up.Str, "browser_name")

	helper.Copy(up.Str|up.Null, "proxy")
	helper.Copy(up.Str|up.Null, "get_proxy_url")
	helper.Copy(up.Bool, "proxy_only_login")

	helper.Copy(up.Str, "displayname_template")

	helper.Copy(up.Bool, "call_start_notices")
	helper.Copy(up.Bool, "send_presence_on_typing")
	helper.Copy(up.Bool, "enable_status_broadcast")
	helper.Copy(up.Bool, "disable_status_broadcast_send")
	helper.Copy(up.Bool, "mute_status_broadcast")
	helper.Copy(up.Str, "status_broadcast_tag")
	helper.Copy(up.Bool, "whatsapp_thumbnail")
	helper.Copy(up.Bool, "url_previews")

	helper.Copy(up.Bool, "history_sync", "request_full_sync")
	helper.Copy(up.Int|up.Null, "history_sync", "full_sync_config", "days_limit")
	helper.Copy(up.Int|up.Null, "history_sync", "full_sync_config", "size_mb_limit")
	helper.Copy(up.Int|up.Null, "history_sync", "full_sync_config", "storage_quota_mb")
	helper.Copy(up.Bool, "history_sync", "media_requests", "auto_request_media")
	helper.Copy(up.Str, "history_sync", "media_requests", "request_method")
	helper.Copy(up.Int, "history_sync", "media_requests", "request_local_time")
	helper.Copy(up.Int, "history_sync", "media_requests", "max_async_handle")
}

type DisplaynameParams struct {
	types.ContactInfo
	Phone string
	// Deprecated form of Phone
	JID string
}

func (c *Config) FormatDisplayname(jid types.JID, contact types.ContactInfo) string {
	var nameBuf strings.Builder
	err := c.displaynameTemplate.Execute(&nameBuf, &DisplaynameParams{
		ContactInfo: contact,
		Phone:       "+" + jid.User,
		JID:         "+" + jid.User,
	})
	if err != nil {
		panic(err)
	}
	return nameBuf.String()
}

func (wa *WhatsAppConnector) GetConfig() (string, any, up.Upgrader) {
	return ExampleConfig, &wa.Config, &up.StructUpgrader{
		SimpleUpgrader: up.SimpleUpgrader(upgradeConfig),
		Blocks: [][]string{
			{"proxy"},
			{"displayname_template"},
			{"call_start_notices"},
			{"history_sync"},
		},
		Base: ExampleConfig,
	}
}
