package moodle

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/inno-gang/goodle"
	"github.com/inno-gang/goodle/richtext"
)

type Client struct {
	baseUrl *url.URL
	token   string
	client  *http.Client
}

func NewClient(
	baseUrl string,
	token string,
	authorizedClient *http.Client,
) (*Client, error) {
	parsedUrl, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}
	return &Client{
		baseUrl: parsedUrl,
		token:   token,
		client:  authorizedClient,
	}, nil
}

func (mc *Client) GetRecentCourses() ([]goodle.Course, error) {
	data, err := mc.callWsFunc("core_course_get_recent_courses", "")
	if err != nil {
		return nil, err
	}

	var courses []*rawCourse
	err = json.NewDecoder(strings.NewReader(data)).Decode(&courses)
	if err != nil {
		return nil, err
	}

	var result []goodle.Course
	for _, course := range courses {
		result = append(result, course)
	}

	return result, nil
}

func (mc *Client) GetCourseSections(courseId int) ([]goodle.Section, error) {
	arg := fmt.Sprintf("{\"courseid\":\"%d\",\"options\":[{\"name\":\"excludemodules\",\"value\":\"0\"},{\"name\":\"excludecontents\",\"value\":\"0\"},{\"name\":\"includestealthmodules\",\"value\":\"1\"}]}", courseId)
	data, err := mc.callWsFunc("core_course_get_contents", arg)

	var rawSections []*rawSection
	err = json.NewDecoder(strings.NewReader(data)).Decode(&rawSections)
	if err != nil {
		return nil, err
	}

	sections := make([]goodle.Section, len(rawSections))
	for i, section := range rawSections {
		sections[i] = section
	}

	return sections, nil
}

type rawCourse struct {
	Id_            int    `json:"id"`
	IdNumber       string `json:"idnumber"`
	FullName       string `json:"fullname"`
	ShortName      string `json:"shortname"`
	ViewUrl        string `json:"viewurl"`
	CourseCategory string `json:"coursecategory"`
	Summary        string `json:"summary"`
	StartDate      uint64 `json:"startdate"`
	EndDate        uint64 `json:"enddate"`
}

func (rc *rawCourse) Id() int {
	return rc.Id_
}

func (rc *rawCourse) Title() string {
	return rc.FullName
}

func (rc *rawCourse) Description() *richtext.RichText {
	parsed, _ := richtext.ParseHtml(rc.Summary)
	return parsed
}

func (rc *rawCourse) MoodleUrl() string {
	return rc.ViewUrl
}

type rawSection struct {
	Id_     int         `json:"id"`
	Name    string      `json:"name"`
	Summary string      `json:"summary"`
	Modules []rawModule `json:"modules"`
}

func (rs *rawSection) Id() int {
	return rs.Id_
}

func (rs *rawSection) Title() string {
	return rs.Name
}

func (rs *rawSection) Description() *richtext.RichText {
	text, _ := richtext.ParseHtml(rs.Summary)
	return text
}

func (rs *rawSection) MoodleUrl() string {
	//TODO implement me
	return ""
}

func (rs *rawSection) Blocks() []goodle.Block {
	out := make([]goodle.Block, len(rs.Modules))
	for i, m := range rs.Modules {
		out[i] = m
	}
	return out
}

type rawModule struct {
	Id_          int              `json:"id"`
	Url_         string           `json:"url"`
	Name         string           `json:"name"`
	ModName      string           `json:"modname"`
	Dates        []rawDate        `json:"dates"`
	Contents     []rawContent     `json:"contents"`
	ContentsInfo *rawContentsInfo `json:"contentsinfo"`
}

func (rm rawModule) Id() int {
	return rm.Id_
}

func (rm rawModule) Title() string {
	return rm.Name
}

func (rm rawModule) Description() *richtext.RichText {
	//TODO implement me
	r, _ := richtext.ParseHtml("")
	return r
}

func (rm rawModule) MoodleUrl() string {
	return rm.Url_
}

func (rm rawModule) Type() goodle.BlockType {
	switch rm.ModName {
	case "url":
		return goodle.BlockTypeLink
	case "resource":
		return goodle.BlockTypeFile
	case "folder":
		return goodle.BlockTypeFolder
	case "quiz":
		return goodle.BlockTypeQuiz
	case "assign":
		return goodle.BlockTypeAssignment
	}
	return goodle.BlockTypeUnsupported
}

func (rm rawModule) SubmissionsAcceptedFrom() (time.Time, bool) {
	for _, d := range rm.Dates {
		if d.DataId == "allowsubmissionsfromdate" {
			return time.Unix(int64(d.Timestamp), 0), true
		}
	}
	return time.Time{}, false
}

func (rm rawModule) DeadlineAt() (time.Time, bool) {
	for _, d := range rm.Dates {
		if d.DataId == "duedate" {
			return time.Unix(int64(d.Timestamp), 0), true
		}
	}
	return time.Time{}, false
}

func (rm rawModule) StrictDeadlineAt() (time.Time, bool) {
	//TODO implement me
	return time.Time{}, false
}

func (rm rawModule) Url() string {
	for _, c := range rm.Contents {
		if c.Type == "url" {
			return c.FileUrl
		}
	}
	return ""
}

func (rm rawModule) DownloadUrl() string {
	for _, c := range rm.Contents {
		if c.Type == "file" {
			return c.FileUrl
		}
	}
	return ""
}

func (rm rawModule) CreatedAt() time.Time {
	for _, c := range rm.Contents {
		if c.Type == "file" {
			return time.Unix(int64(c.TimeCreated), 0)
		}
	}
	return time.Time{}
}

func (rm rawModule) LastModifiedAt() time.Time {
	for _, c := range rm.Contents {
		if c.Type == "file" {
			return time.Unix(int64(c.TimeModified), 0)
		}
	}
	return time.Time{}
}

func (rm rawModule) SizeBytes() uint64 {
	if rm.ContentsInfo == nil {
		return 0
	}
	return rm.ContentsInfo.FilesSize
}

type rawDate struct {
	Label     string `json:"label"`
	Timestamp uint64 `json:"timestamp"`
	DataId    string `json:"dataid"`
}

type rawContent struct {
	Type         string `json:"type"`
	FileName     string `json:"filename"`
	FileSize     uint64 `json:"filesize"`
	FileUrl      string `json:"fileurl"`
	TimeCreated  uint64 `json:"timecreated"`
	TimeModified uint64 `json:"timemodified"`
	MimeType     string `json:"mimetype"`
	Author       string `json:"author"`
}

type rawContentsInfo struct {
	FilesCount   uint   `json:"filescount"`
	FilesSize    uint64 `json:"filessize"`
	TimeModified uint64 `json:"timemodified"`
}
