package main

import "time"

// Release a release. Stolen and modified from
// https://github.com/octokit/go-sdk/blob/main/pkg/github/models/release.go#L9.
type Release struct {
	Assets          []ReleaseAsset `json:"assets"`                 // The Assets property
	AssetsUrl       string         `json:"assets_url"`             // The AssetsUrl property
	Author          SimpleUser     `json:"author"`                 // A GitHub user.
	Body            *string        `json:"body,omitempty"`         // The Body property
	BodyHtml        string         `json:"body_html"`              // The BodyHtml property
	BodyText        string         `json:"body_text"`              // The BodyText property
	CreatedAt       time.Time      `json:"created_at"`             // The CreatedAt property
	DiscussionUrl   string         `json:"discussion_url"`         // The URL of the release discussion.
	Draft           bool           `json:"draft"`                  // true to create a Draft (unpublished) release, false to create a published one.
	HtmlUrl         string         `json:"html_url"`               // The HtmlUrl property
	Id              int32          `json:"id"`                     // The Id property
	MentionsCount   int32          `json:"mentions_count"`         // The MentionsCount property
	Name            *string        `json:"name,omitempty"`         // The Name property
	NodeId          string         `json:"node_id"`                // The NodeId property
	Prerelease      bool           `json:"prerelease"`             // Whether to identify the release as a Prerelease or a full release.
	PublishedAt     *time.Time     `json:"published_at,omitempty"` // The PublishedAt property
	Reactions       ReactionRollup `json:"reactions"`              // The Reactions property
	TagName         string         `json:"tag_name"`               // The Name of the tag.
	TarballUrl      *string        `json:"tarball_url,omitempty"`  // The TarballUrl property
	TargetCommitish string         `json:"target_commitish"`       // Specifies the commitish value that determines where the Git tag is created from.
	UploadUrl       string         `json:"upload_url"`             // The UploadUrl property
	Url             string         `json:"url"`                    // The Url property
	ZipballUrl      *string        `json:"zipball_url,omitempty"`  // The ZipballUrl property
}

// ReleaseAsset data related to a release.
type ReleaseAsset struct {
	BrowserDownloadUrl string      `json:"browser_download_url"` // The BrowserDownloadUrl property
	ContentType        string      `json:"content_type"`         // The ContentType property
	CreatedAt          time.Time   `json:"created_at"`           // The CreatedAt property
	DownloadCount      int32       `json:"download_count"`       // The DownloadCount property
	Id                 int32       `json:"id"`                   // The Id property
	Label              *string     `json:"label,omitempty"`      // The Label property
	Name               string      `json:"name"`                 // The file Name of the asset.
	NodeId             string      `json:"node_id"`              // The NodeId property
	Size               int32       `json:"size"`                 // The Size property
	State              int         `json:"state"`                // State of the release asset.
	UpdatedAt          time.Time   `json:"updated_at"`           // The UpdatedAt property
	Uploader           *SimpleUser `json:"uploader,omitempty"`   // A GitHub user.
	Url                string      `json:"url"`                  // The Url property
}

// SimpleUser a GitHub user.
type SimpleUser struct {
	AvatarUrl         string  `json:"avatar_url"`            // The AvatarUrl property
	Email             *string `json:"email,omitempty"`       // The Email property
	EventsUrl         string  `json:"events_url"`            // The EventsUrl property
	FollowersUrl      string  `json:"followers_url"`         // The FollowersUrl property
	FollowingUrl      string  `json:"following_url"`         // The FollowingUrl property
	GistsUrl          string  `json:"gists_url"`             // The GistsUrl property
	GravatarId        *string `json:"gravatar_id,omitempty"` // The GravatarId property
	HtmlUrl           string  `json:"html_url"`              // The HtmlUrl property
	Id                int32   `json:"id"`                    // The Id property
	Login             string  `json:"login"`                 // The Login property
	Name              *string `json:"name,omitempty"`        // The Name property
	NodeId            string  `json:"node_id"`               // The NodeId property
	OrganizationsUrl  string  `json:"organizations_url"`     // The OrganizationsUrl property
	ReceivedEventsUrl string  `json:"received_events_url"`   // The ReceivedEventsUrl property
	ReposUrl          string  `json:"repos_url"`             // The ReposUrl property
	SiteAdmin         bool    `json:"site_admin"`            // The SiteAdmin property
	StarredAt         string  `json:"starred_at"`            // The StarredAt property
	StarredUrl        string  `json:"starred_url"`           // The StarredUrl property
	SubscriptionsUrl  string  `json:"subscriptions_url"`     // The SubscriptionsUrl property
	TypeEscaped       string  `json:"type_escaped"`          // The type property
	Url               string  `json:"url"`                   // The Url property
}

// ReactionRollup a reaction rollup.
type ReactionRollup struct {
	Confused   int32  `json:"confused"`    // The Confused property
	Eyes       int32  `json:"eyes"`        // The Eyes property
	Heart      int32  `json:"heart"`       // The Heart property
	Hooray     int32  `json:"hooray"`      // The Hooray property
	Laugh      int32  `json:"laugh"`       // The Laugh property
	Minus1     int32  `json:"minus_1"`     // The Minus1 property
	Plus1      int32  `json:"plus_1"`      // The Plus1 property
	Rocket     int32  `json:"rocket"`      // The Rocket property
	TotalCount int32  `json:"total_count"` // The TotalCount property
	Url        string `json:"url"`         // The Url property
}
