package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/Daniel1212649/LinksHelperBot/lib/e"
)

const (
	getUpdatesMethod      = "getUpdates"
	sendMessageMethod     = "sendMessage"
	editMessageTextMethod = "editMessageText"
	answerCallbackMethod  = "answerCallbackQuery"
)

type Client struct {
	host     string
	basePath string
	client   http.Client
}

func New(host string, token string, timeout time.Duration) *Client {
	return &Client{
		host:     host,
		basePath: newBasePath(token),
		client: http.Client{
			Timeout: timeout,
		},
	}
}

func newBasePath(token string) string {
	return "bot" + token
}

func (c *Client) Updates(ctx context.Context, offset int, limit int) ([]Update, error) {
	q := url.Values{}
	q.Add("offset", strconv.Itoa(offset))
	q.Add("limit", strconv.Itoa(limit))
	q.Add("timeout", "25")

	data, err := c.doGet(ctx, getUpdatesMethod, q)
	if err != nil {
		return nil, err
	}

	var res UpdateResponse
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, err
	}
	if !res.Ok {
		return nil, telegramAPIError(res.ErrorCode, res.Description)
	}

	return res.Result, nil
}

func (c *Client) SendMessage(ctx context.Context, chatID int64, text string, markup *InlineKeyboardMarkup) error {
	form := url.Values{}
	form.Add("chat_id", strconv.FormatInt(chatID, 10))
	form.Add("text", text)
	form.Add("disable_web_page_preview", "false")
	if err := addReplyMarkup(form, markup); err != nil {
		return err
	}

	data, err := c.doPost(ctx, sendMessageMethod, form)
	if err != nil {
		return e.Wrap("can't send message", err)
	}

	return decodeAPIResponse(data, "can't send message")
}

func (c *Client) EditMessageText(ctx context.Context, chatID int64, messageID int, text string, markup *InlineKeyboardMarkup) error {
	form := url.Values{}
	form.Add("chat_id", strconv.FormatInt(chatID, 10))
	form.Add("message_id", strconv.Itoa(messageID))
	form.Add("text", text)
	form.Add("disable_web_page_preview", "false")
	if err := addReplyMarkup(form, markup); err != nil {
		return err
	}

	data, err := c.doPost(ctx, editMessageTextMethod, form)
	if err != nil {
		return e.Wrap("can't edit message", err)
	}

	return decodeAPIResponse(data, "can't edit message")
}

func (c *Client) AnswerCallbackQuery(ctx context.Context, callbackQueryID string, text string) error {
	form := url.Values{}
	form.Add("callback_query_id", callbackQueryID)
	if text != "" {
		form.Add("text", text)
	}

	data, err := c.doPost(ctx, answerCallbackMethod, form)
	if err != nil {
		return e.Wrap("can't answer callback query", err)
	}

	return decodeAPIResponse(data, "can't answer callback query")
}

func addReplyMarkup(form url.Values, markup *InlineKeyboardMarkup) error {
	if markup == nil {
		return nil
	}

	raw, err := json.Marshal(markup)
	if err != nil {
		return e.Wrap("can't encode reply markup", err)
	}
	form.Add("reply_markup", string(raw))
	return nil
}

func decodeAPIResponse(data []byte, wrapMsg string) error {
	var res APIResponse
	if err := json.Unmarshal(data, &res); err != nil {
		return e.Wrap(wrapMsg+", can't decode response", err)
	}
	if !res.Ok {
		return e.Wrap(wrapMsg, telegramAPIError(res.ErrorCode, res.Description))
	}
	return nil
}

func (c *Client) doGet(ctx context.Context, method string, query url.Values) ([]byte, error) {
	u := c.methodURL(method)
	u.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	return c.doRequest(req)
}

func (c *Client) doPost(ctx context.Context, method string, form url.Values) ([]byte, error) {
	u := c.methodURL(method)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return c.doRequest(req)
}

func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, e.Wrap("can't do request", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("telegram returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

func (c *Client) methodURL(method string) url.URL {
	return url.URL{
		Scheme: "https",
		Host:   c.host,
		Path:   path.Join(c.basePath, method),
	}
}

func telegramAPIError(code int, description string) error {
	if description == "" {
		return fmt.Errorf("telegram API error code %d", code)
	}
	return fmt.Errorf("telegram API error code %d: %s", code, description)
}
