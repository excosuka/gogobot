package telegram

import (
	"encoding/json"
	"gogobot/events/telegram/keyboards"
	"gogobot/lib/e"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
)

type Client struct {
	host     string
	basePath string
	client   http.Client
}

const (
	getUpdatesMethod  = "getUpdates"
	sendMessageMethod = "sendMessage"
)

func newBasePath(token string) string {
	return "bot" + token
}

func New(host, token string) *Client {
	return &Client{
		host:     host,
		basePath: newBasePath(token),
		client:   http.Client{},
	}
}

func (c *Client) Updates(offset int, limit int) (updates []Update, err error) {
	defer func() { err = e.WrapIfErr("can`t get updates", err) }()

	q := url.Values{}
	q.Add("offset", strconv.Itoa(offset))
	q.Add("limit", strconv.Itoa(limit))

	data, err := c.doRequest(getUpdatesMethod, q)
	if err != nil {
		return nil, err
	}

	var res UpdatesResponse

	if err := json.Unmarshal(data, &res); err != nil {
		return nil, err
	}
	return res.Result, nil
}

func (c *Client) SendMessage(chatID int, text string) error {
	q := url.Values{}
	q.Add("chat_id", strconv.Itoa(chatID))
	q.Add("text", text)

	_, err := c.doRequest(sendMessageMethod, q)
	if err != nil {
		return e.Wrap("can`t send message", err)
	}

	return nil
}

type SendMessageResponse struct {
	Ok     bool `json:"ok"`
	Result struct {
		MessageID int `json:"message_id"`
	} `json:"result"`
}

func (c *Client) SendMessageWithKeyboard(chatID int, text string, keyboard keyboards.ReplyMenuKeyboard) (int, error) {
	q := url.Values{}
	keyboardToJson, err := json.Marshal(keyboard)
	if err != nil {
		return 0, e.Wrap("can't marshal keyboard to json", err)
	}

	q.Add("chat_id", strconv.Itoa(chatID))
	q.Add("text", text)
	q.Add("reply_markup", string(keyboardToJson))

	data, err := c.doRequest(sendMessageMethod, q)
	if err != nil {
		return 0, e.Wrap("can't send message", err)
	}

	var resp SendMessageResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return 0, e.Wrap("can't parse sendMessage response", err)
	}

	return resp.Result.MessageID, nil
}

func (c *Client) doRequest(method string, query url.Values) (data []byte, err error) {
	defer func() { err = e.WrapIfErr("can`t do request", err) }()

	u := url.URL{
		Scheme: "https",
		Host:   c.host,
		Path:   path.Join(c.basePath, method),
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)

	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = query.Encode()

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	return body, nil

}
func (c *Client) EditMessageReplyMarkup(chatID int, messageID int) error {
	q := url.Values{}
	q.Add("chat_id", strconv.Itoa(chatID))
	q.Add("message_id", strconv.Itoa(messageID))

	_, err := c.doRequest("editMessageReplyMarkup", q)
	if err != nil {
		return e.Wrap("can't edit message reply markup", err)
	}
	return nil
}
