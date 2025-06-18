package vendors

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"

	"github.com/devopsext/utils"
)

//go:embed slack.tmpl
var msgTemplate string

const slackBaseURL = "https://slack.com/api/"

const (
	slackFilesUpload           = "files.upload"
	slackChatPostMessage       = "chat.postMessage"
	slackReactionsAdd          = "reactions.add"
	slackUsersLookupByEmail    = "users.lookupByEmail"
	slackUsergroupsUsersUpdate = "usergroups.users.update"
	slackConversationsHistory  = "conversations.history"
)

type SlackOptions struct {
	Timeout  int
	Insecure bool
	Token    string
}

type SlackMessageOptions struct {
	Channel     string
	Thread      string
	Title       string
	Text        string
	Attachments string
	Blocks      string
}

type SlackFileOptions struct {
	Channel string
	Thread  string
	Title   string
	Text    string
	Name    string
	Content string
	Type    string
}

type SlackReactionOptions struct {
	Channel string
	Thread  string
	Name    string
}

type SlackMessageBlock struct {
	Type    string `json:"type"`
	BlockID string `json:"block_id"`
}

type SlackMessage struct {
	Subtype string               `json:"subtype"`
	Text    string               `json:"text"`
	Type    string               `json:"type"`
	TS      string               `json:"ts"`
	BotID   string               `json:"bot_id,omitempty"`
	UserID  string               `json:"user_id,omitempty"`
	Blocks  []*SlackMessageBlock `json:"blocks,omitempty"`
}

type SlackMessageResponse struct {
	OK      bool          `json:"ok"`
	Channel string        `json:"channel"`
	TS      string        `json:"ts"`
	Message *SlackMessage `json:"message,omitempty"`
}

type GetConversationHistoryParameters struct {
	ChannelID          string
	Cursor             string
	Inclusive          bool
	Latest             string
	Limit              int
	Oldest             string
	IncludeAllMetadata bool
}

type GetConversationHistoryResponse struct {
	Ok       bool `json:"ok"`
	Messages []struct {
		Type string `json:"type"`
		User string `json:"user"`
		Text string `json:"text"`
		Ts   string `json:"ts"`
	} `json:"messages"`
	HasMore          bool `json:"has_more"`
	PinCount         int  `json:"pin_count"`
	ResponseMetadata struct {
		NextCursor string `json:"next_cursor"`
	} `json:"response_metadata"`
}

type SlackOutputOptions struct {
	Output      string // path to output if empty to stdout
	OutputQuery string
}

/*type SlackMessage struct {
	Token       string
	Channel     string
	ParentTS    string
	Title       string
	Message     string
	ImageURL    string
	FileName    string
	FileContent string
	QuoteColor  string
}*/

type SlackUserEmail struct {
	Email string
}

type SlackUsergroupUsers struct {
	Usergroup string   `json:"usergroup"`
	Users     []string `json:"users"`
}

type Slack struct {
	client  *http.Client
	options SlackOptions
}

func (s *Slack) apiURL(cmd string) string {
	return slackBaseURL + cmd
}

func (s *Slack) getAuth(opts SlackOptions) string {

	auth := ""
	if !utils.IsEmpty(opts.Token) {
		auth = fmt.Sprintf("Bearer %s", opts.Token)
		return auth
	}
	return auth
}

/*
	func (s *Slack) Send() ([]byte, error) {
		m := SlackMessage{
			Token:       s.options.Token,
			Channel:     s.options.Channel,
			Title:       s.options.Title,
			Message:     s.options.Message,
			FileContent: s.options.File,
			ParentTS:    s.options.ParentTS,
			QuoteColor:  s.options.QuoteColor,
		}
		return s.SendCustom(m)
	}

	func (s *Slack) SendFile() ([]byte, error) {
		m := SlackMessage{
			Token:       s.options.Token,
			Channel:     s.options.Channel,
			ParentTS:    s.options.ParentTS,
			Title:       s.options.Title,
			Message:     s.options.Message,
			ImageURL:    s.options.ImageURL,
			FileName:    s.options.FileName,
			FileContent: s.options.File,
		}
		return s.SendCustomFile(m)
	}

	func (s *Slack) SendMessage() ([]byte, error) {
		m := SlackMessage{
			Token:       s.options.Token,
			Channel:     s.options.Channel,
			ParentTS:    s.options.ParentTS,
			Title:       s.options.Title,
			Message:     s.options.Message,
			ImageURL:    s.options.ImageURL,
			FileName:    s.options.FileName,
			FileContent: s.options.File,
			QuoteColor:  s.options.QuoteColor,
		}
		return s.sendMessage(m)
	}

	func (s *Slack) SendCustomMessage(m SlackMessage) ([]byte, error) {
		return s.sendMessage(m)
	}

func (s *Slack) sendMessage(m SlackMessage) ([]byte, error) {

		if m.Message == "" {
			return nil, errors.New("slack message is empty")
		}

		if m.Title == "" {
			// find the first nonempty line
			lines := strings.Split(m.Message, "\n")
			for _, line := range lines {
				if line != "" {
					m.Title = line
					break
				}
			}

			// if still empty, use the first line
			if m.Title == "" {
				m.Title = "No title"
			}

			m.Title = common.TruncateString(m.Title, 150)
		}
		jsonMsg, err := s.prepareMessage(m)
		if err != nil {
			return nil, err
		}
		q := url.Values{}
		b, err := s.post(m.Token, s.apiURL(slackChatPostMessage), q, "application/json; charset=utf-8", *jsonMsg)
		if err != nil {
			return nil, err
		}
		return b, nil
	}

func (s *Slack) SendCustom(m SlackMessage) ([]byte, error) {

		var body bytes.Buffer
		w := multipart.NewWriter(&body)
		defer func() {
			w.Close()
		}()

		if err := w.WriteField("initial_comment", m.Message); err != nil {
			return nil, err
		}

		if err := w.WriteField("title", m.Title); err != nil {
			return nil, err
		}

		if err := w.WriteField("content", m.FileContent); err != nil {
			return nil, err
		}

		if err := w.Close(); err != nil {
			return nil, err
		}

		q := url.Values{}
		q.Add("channels", s.options.Channel)

		return s.post(m.Token, s.apiURL(slackFilesUpload), q, w.FormDataContentType(), body)
	}

func (s *Slack) SendCustomFile(m SlackMessage) ([]byte, error) {

		var body bytes.Buffer
		w := multipart.NewWriter(&body)
		defer func() {
			w.Close()
		}()

		if !utils.IsEmpty(m.Message) {
			if err := w.WriteField("initial_comment", m.Message); err != nil {
				return nil, err
			}
		}

		if !utils.IsEmpty(m.Title) {
			if err := w.WriteField("title", m.Title); err != nil {
				return nil, err
			}
		}

		if !utils.IsEmpty(m.ParentTS) {
			if err := w.WriteField("thread_ts", m.ParentTS); err != nil {
				return nil, err
			}
		}

		fw, err := w.CreateFormFile("file", m.FileName)
		if err != nil {
			return nil, err
		}

		if _, err := fw.Write([]byte(m.FileContent)); err != nil {
			return nil, err
		}

		if err := w.Close(); err != nil {
			return nil, err
		}

		q := url.Values{}
		q.Add("channels", m.Channel)

		return s.post(m.Token, s.apiURL(slackFilesUpload), q, w.FormDataContentType(), body)
	}

	func (s *Slack) prepareMessage(m SlackMessage) (*bytes.Buffer, error) {
		t, err := template.New("slack").Parse(msgTemplate)
		if err != nil {
			return nil, err
		}

		ts := strings.ReplaceAll(m.Message, "\r", "")
		m.Message = strings.ReplaceAll(ts, "\n", "\\n")

		b := &bytes.Buffer{}

		if err := t.Execute(b, m); err != nil {
			return nil, err
		}
		return b, nil
	}

func (s *Slack) post(token string, URL string, query url.Values, contentType string, body bytes.Buffer) ([]byte, error) {

		reader := bytes.NewReader(body.Bytes())

		req, err := http.NewRequest("POST", URL, reader)
		if err != nil {
			return nil, err
		}

		if token == "" {
			token = s.options.Token
		}
		req.Header.Set("Content-Type", contentType)
		req.Header.Set("Authorization", "Bearer "+token)
		req.URL.RawQuery = query.Encode()

		resp, err := s.client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return b, nil
	}
*/

func (s *Slack) CustomSendMessage(slackOptions SlackOptions, messageOptions SlackMessageOptions) ([]byte, error) {

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	defer func() {
		w.Close()
	}()

	if err := w.WriteField("channel", messageOptions.Channel); err != nil {
		return nil, err
	}

	if !utils.IsEmpty(messageOptions.Thread) {
		if err := w.WriteField("thread_ts", messageOptions.Thread); err != nil {
			return nil, err
		}
	}

	if !utils.IsEmpty(messageOptions.Text) {
		if err := w.WriteField("text", messageOptions.Text); err != nil {
			return nil, err
		}
	}

	if !utils.IsEmpty(messageOptions.Attachments) {
		if err := w.WriteField("attachments", messageOptions.Attachments); err != nil {
			return nil, err
		}
	}

	if !utils.IsEmpty(messageOptions.Blocks) {
		if err := w.WriteField("blocks", messageOptions.Blocks); err != nil {
			return nil, err
		}
	}

	if err := w.WriteField("title", messageOptions.Title); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	return utils.HttpPostRaw(s.client, s.apiURL(slackChatPostMessage), w.FormDataContentType(), s.getAuth(slackOptions), body.Bytes())
}

func (s *Slack) SendMessage(messageOptions SlackMessageOptions) ([]byte, error) {
	return s.CustomSendMessage(s.options, messageOptions)
}

func (s *Slack) CustomSendFile(slackOptions SlackOptions, fileOptions SlackFileOptions) ([]byte, error) {

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	defer func() {
		w.Close()
	}()

	if err := w.WriteField("channels", fileOptions.Channel); err != nil {
		return nil, err
	}

	if !utils.IsEmpty(fileOptions.Thread) {
		if err := w.WriteField("thread_ts", fileOptions.Thread); err != nil {
			return nil, err
		}
	}

	if err := w.WriteField("initial_comment", fileOptions.Text); err != nil {
		return nil, err
	}

	if err := w.WriteField("title", fileOptions.Title); err != nil {
		return nil, err
	}

	if !utils.IsEmpty(fileOptions.Type) {
		if err := w.WriteField("filetype", fileOptions.Type); err != nil {
			return nil, err
		}
	}

	fw, err := w.CreateFormFile("file", fileOptions.Name)
	if err != nil {
		return nil, err
	}

	if _, err := fw.Write([]byte(fileOptions.Content)); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	return utils.HttpPostRaw(s.client, s.apiURL(slackFilesUpload), w.FormDataContentType(), s.getAuth(slackOptions), body.Bytes())
}

func (s *Slack) SendFile(fileOptions SlackFileOptions) ([]byte, error) {
	return s.CustomSendFile(s.options, fileOptions)
}

func (s *Slack) CustomAddReaction(slackOptions SlackOptions, reactionOptions SlackReactionOptions) ([]byte, error) {

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	defer func() {
		w.Close()
	}()

	if err := w.WriteField("channel", reactionOptions.Channel); err != nil {
		return nil, err
	}

	if err := w.WriteField("name", reactionOptions.Name); err != nil {
		return nil, err
	}

	if !utils.IsEmpty(reactionOptions.Thread) {
		if err := w.WriteField("timestamp", reactionOptions.Thread); err != nil {
			return nil, err
		}
	}

	if err := w.Close(); err != nil {
		return nil, err
	}
	return utils.HttpPostRaw(s.client, s.apiURL(slackReactionsAdd), w.FormDataContentType(), s.getAuth(slackOptions), body.Bytes())
}

func (s *Slack) AddReaction(options SlackReactionOptions) ([]byte, error) {
	return s.CustomAddReaction(s.options, options)
}

func (s *Slack) CustomGetUser(slackOptions SlackOptions, slackUser SlackUserEmail) ([]byte, error) {
	params := make(url.Values)
	params.Add("email", slackUser.Email)

	u, err := url.Parse(s.apiURL(slackUsersLookupByEmail))
	if err != nil {
		return nil, err
	}

	u.RawQuery = params.Encode()
	return utils.HttpGetRaw(s.client, u.String(), "application/x-www-form-urlencoded", s.getAuth(slackOptions))
}

func (s *Slack) GetUser(options SlackUserEmail) ([]byte, error) {
	return s.CustomGetUser(s.options, options)
}

func (s *Slack) CustomUpdateUsergroup(slackOptions SlackOptions, slackUpdateUsergroup SlackUsergroupUsers) ([]byte, error) {

	body := &SlackUsergroupUsers{
		Usergroup: slackUpdateUsergroup.Usergroup,
		Users:     slackUpdateUsergroup.Users,
	}

	req, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return utils.HttpPostRaw(s.client, s.apiURL(slackUsergroupsUsersUpdate), "application/json", s.getAuth(slackOptions), req)
}

func (s *Slack) UpdateUsergroup(options SlackUsergroupUsers) ([]byte, error) {
	return s.CustomUpdateUsergroup(s.options, options)
}

func (s *Slack) GetConversationHistory(options GetConversationHistoryParameters) ([]byte, error) {
	return s.CustomGetConversationHistory(s.options, options)
}

func (s *Slack) CustomGetConversationHistory(slackOptions SlackOptions, getConversationHistoryParameters GetConversationHistoryParameters) ([]byte, error) {

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	defer func() {
		w.Close()
	}()

	if !utils.IsEmpty(getConversationHistoryParameters.Cursor) {
		if err := w.WriteField("thread_ts", getConversationHistoryParameters.Cursor); err != nil {
			return nil, err
		}
	}

	if getConversationHistoryParameters.Cursor != "" {
		if err := w.WriteField("cursor", getConversationHistoryParameters.Cursor); err != nil {
			return nil, err
		}
	}
	if getConversationHistoryParameters.Inclusive {
		if err := w.WriteField("inclusive", "1"); err != nil {
			return nil, err
		}
	} else {
		if err := w.WriteField("inclusive", "0"); err != nil {
			return nil, err
		}
	}
	if getConversationHistoryParameters.Latest != "" {
		if err := w.WriteField("latest", getConversationHistoryParameters.Latest); err != nil {
			return nil, err
		}
	}
	if getConversationHistoryParameters.Limit != 0 {
		if err := w.WriteField("limit", strconv.Itoa(getConversationHistoryParameters.Limit)); err != nil {
			return nil, err
		}
	}
	if getConversationHistoryParameters.Oldest != "" {
		if err := w.WriteField("oldest", getConversationHistoryParameters.Oldest); err != nil {
			return nil, err
		}
	}
	if getConversationHistoryParameters.IncludeAllMetadata {
		if err := w.WriteField("include_all_metadata", "1"); err != nil {
			return nil, err
		}
	} else {
		if err := w.WriteField("include_all_metadata", "0"); err != nil {
			return nil, err
		}
	}

	if err := w.Close(); err != nil {
		return nil, err
	}
	return utils.HttpPostRaw(s.client, s.apiURL(slackChatPostMessage), w.FormDataContentType(), s.getAuth(slackOptions), body.Bytes())
	return utils.HttpPostRaw(s.client, s.apiURL(slackConversationsHistory), w.FormDataContentType(), s.getAuth(slackOptions), body.Bytes())

}

func NewSlack(options SlackOptions) *Slack {

	slack := &Slack{
		client:  utils.NewHttpClient(options.Timeout, options.Insecure),
		options: options,
	}
	return slack
}
