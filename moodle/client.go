package moodle

import (
	"context"
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

func (mc *Client) GetRecentCourses(ctx context.Context) ([]goodle.Course, error) {
	data, err := mc.callWsFunc(ctx, "core_course_get_recent_courses", "")
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

func (mc *Client) GetCourseSections(ctx context.Context, courseId int) ([]goodle.Section, error) {
	arg := fmt.Sprintf("{\"courseid\":\"%d\",\"options\":[{\"name\":\"excludemodules\",\"value\":\"0\"},{\"name\":\"excludecontents\",\"value\":\"0\"},{\"name\":\"includestealthmodules\",\"value\":\"1\"}]}", courseId)
	data, err := mc.callWsFunc(ctx, "core_course_get_contents", arg)
	if err != nil {
		return nil, err
	}

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
	FullName       string `json:"fullname"`
	ViewUrl        string `json:"viewurl"`
	CourseCategory string `json:"coursecategory"`
	Summary        string `json:"summary"`
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

func (rs *rawSection) Blocks() []goodle.Block {
	out := make([]goodle.Block, len(rs.Modules))
	for i, m := range rs.Modules {
		out[i] = m.toBlock()
	}
	return out
}

type rawModule struct {
	Id           int              `json:"id"`
	Url          string           `json:"url"`
	Name         string           `json:"name"`
	ModName      string           `json:"modname"`
	Dates        []rawDate        `json:"dates"`
	Contents     []rawContent     `json:"contents"`
	ContentsInfo *rawContentsInfo `json:"contentsinfo"`
}

type rawDate struct {
	Label     string `json:"label"`
	Timestamp uint64 `json:"timestamp"`
	DataId    string `json:"dataid"`
}

type rawContent struct {
	Type         string `json:"type"`
	FileName     string `json:"filename"`
	FileSize     uint   `json:"filesize"`
	FileUrl      string `json:"fileurl"`
	TimeCreated  uint   `json:"timecreated"`
	TimeModified uint   `json:"timemodified"`
	MimeType     string `json:"mimetype"`
	Author       string `json:"author"`
}

type rawContentsInfo struct {
	FilesCount   uint `json:"filescount"`
	FilesSize    uint `json:"filessize"`
	TimeModified uint `json:"timemodified"`
}

func (rm *rawModule) toBlock() goodle.Block {
	m := module{
		id:        rm.Id,
		name:      rm.Name,
		moodleUrl: rm.Url,
	}
	switch rm.ModName {
	case "url":
		return &moduleLink{
			module: m,
			link:   rm.Contents[0].FileUrl,
		}
	case "resource":
		return &moduleFile{
			module:       m,
			fileName:     rm.Contents[0].FileName,
			fileSize:     rm.Contents[0].FileSize,
			fileUrl:      rm.Contents[0].FileUrl,
			mimeType:     rm.Contents[0].MimeType,
			timeCreated:  time.Unix(int64(rm.Contents[0].TimeCreated), 0),
			timeModified: time.Unix(int64(rm.Contents[0].TimeModified), 0),
		}
	case "folder":
		return &moduleFolder{
			module: m,
		}
	case "assign":
		//TODO parse all
		var submissionsFrom, due, strict time.Time
		for _, d := range rm.Dates {
			if d.DataId == "allowsubmissionsfromdate" {
				submissionsFrom = time.Unix(int64(d.Timestamp), 0)
			} else if d.DataId == "duedate" {
				due = time.Unix(int64(d.Timestamp), 0)
			}
		}
		return &moduleAssignment{
			module:              m,
			allowSubmissionFrom: submissionsFrom,
			dueDate:             due,
			strictDueDate:       strict,
		}
	case "quiz":
		//TODO parse dates
		var submissionsFrom, due time.Time
		return &moduleQuiz{
			module:               m,
			allowSubmissionsFrom: submissionsFrom,
			dueDate:              due,
		}
	}
	return &m
}

type module struct {
	id        int
	name      string
	moodleUrl string
}

func (m *module) Id() int {
	return m.id
}

func (m *module) Title() string {
	return m.name
}

func (m *module) MoodleUrl() string {
	return m.moodleUrl
}

func (m *module) Type() goodle.BlockType {
	return goodle.BlockTypeUnknown
}

type moduleLink struct {
	module
	link string
}

func (ml *moduleLink) Type() goodle.BlockType {
	return goodle.BlockTypeLink
}

func (ml *moduleLink) Url() string {
	return ml.link
}

type moduleFile struct {
	module
	fileName     string
	fileSize     uint
	fileUrl      string
	mimeType     string
	timeCreated  time.Time
	timeModified time.Time
}

func (mf *moduleFile) Type() goodle.BlockType {
	return goodle.BlockTypeFile
}

func (mf *moduleFile) DownloadUrl() string {
	return mf.fileUrl
}

func (mf *moduleFile) FileName() string {
	return mf.fileName
}

func (mf *moduleFile) SizeBytes() uint {
	return mf.fileSize
}

func (mf *moduleFile) MimeType() string {
	return mf.mimeType
}

func (mf *moduleFile) CreatedAt() time.Time {
	return mf.timeCreated
}

func (mf *moduleFile) LastModifiedAt() time.Time {
	return mf.timeModified
}

type moduleFolder struct {
	module
}

func (mf *moduleFolder) Type() goodle.BlockType {
	return goodle.BlockTypeFolder
}

type moduleAssignment struct {
	module
	allowSubmissionFrom time.Time
	dueDate             time.Time
	strictDueDate       time.Time
}

func (ma *moduleAssignment) Type() goodle.BlockType {
	return goodle.BlockTypeAssignment
}

func (ma *moduleAssignment) SubmissionsAcceptedFrom() time.Time {
	return ma.allowSubmissionFrom
}

func (ma *moduleAssignment) DeadlineAt() time.Time {
	return ma.dueDate
}

func (ma *moduleAssignment) StrictDeadlineAt() time.Time {
	return ma.strictDueDate
}

type moduleQuiz struct {
	module
	allowSubmissionsFrom time.Time
	dueDate              time.Time
}

func (m *moduleQuiz) Type() goodle.BlockType {
	return goodle.BlockTypeQuiz
}

func (m *moduleQuiz) OpensAt() time.Time {
	return m.allowSubmissionsFrom
}

func (m *moduleQuiz) ClosesAt() time.Time {
	return m.dueDate
}
