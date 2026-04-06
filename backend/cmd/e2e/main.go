package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type apiResp struct {
	Code      int             `json:"code"`
	Message   string          `json:"message"`
	Data      json.RawMessage `json:"data"`
	DataTotal int64           `json:"data_total"`
}

type echoLog struct {
	ID         int64  `json:"id"`
	Substat1   int64  `json:"substat1"`
	Substat2   int64  `json:"substat2"`
	Substat3   int64  `json:"substat3"`
	Substat4   int64  `json:"substat4"`
	Substat5   int64  `json:"substat5"`
	SubstatAll int64  `json:"substat_all"`
	S1Desc     string `json:"s1_desc"`
	S2Desc     string `json:"s2_desc"`
	S3Desc     string `json:"s3_desc"`
	S4Desc     string `json:"s4_desc"`
	S5Desc     string `json:"s5_desc"`
	Clazz      string `json:"clazz"`
	UserID     int64  `json:"user_id"`
	OperatorID int64  `json:"operator_id"`
}

func main() {
	baseURL := getenv("E2E_BASE_URL", "http://127.0.0.1:8888")
	token := os.Getenv("E2E_TOKEN")
	if token == "" {
		fail("E2E_TOKEN is required")
	}
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		fail("DATABASE_URL is required")
	}

	client := &http.Client{Timeout: 5 * time.Second}
	user := struct {
		ID          int64    `json:"id"`
		Name        string   `json:"name"`
		Permissions []string `json:"permissions"`
	}{}
	requestRawJSON(client, http.MethodGet, baseURL+"/auth/me", token, nil, &user)
	if user.ID <= 0 {
		fail("auth/me returned invalid user id")
	}

	ws, err := dialWebsocket(strings.Replace(baseURL, "http://", "ws://", 1) + "/ws?operator_id=" + fmt.Sprint(user.ID))
	if err != nil {
		fail("websocket dial failed: %v", err)
	}
	defer ws.Close()

	e2eUserID := time.Now().Unix()
	createPayload := map[string]any{
		"user_id":     e2eUserID,
		"clazz":       "沉日劫明",
		"substat1":    0,
		"substat2":    0,
		"substat3":    0,
		"substat4":    0,
		"substat5":    0,
		"substat_all": 0,
	}
	var created echoLog
	mustDecode(requestJSON(client, http.MethodPost, baseURL+"/echo_log", token, createPayload), &created)
	if created.ID <= 0 {
		fail("create echo_log failed")
	}
	defer cleanup(databaseURL, created.ID)

	updatePayload := map[string]any{
		"id":          created.ID,
		"user_id":     e2eUserID,
		"clazz":       "沉日劫明",
		"substat1":    8193,
		"substat_all": 1,
		"s1_desc":     "暴击 6.3%",
	}
	var updated echoLog
	mustDecode(requestJSON(client, http.MethodPatch, baseURL+"/echo_log", token, updatePayload), &updated)
	if updated.Substat1 != 8193 {
		fail("patch echo_log did not persist substat1")
	}

	tunePayload := map[string]any{
		"id":          created.ID,
		"user_id":     e2eUserID,
		"clazz":       "沉日劫明",
		"substat1":    8193,
		"substat2":    16386,
		"substat_all": 3,
		"s1_desc":     "暴击 6.3%",
		"s2_desc":     "暴击伤害 12.6%",
		"position":    1,
		"substat":     1,
		"value":       0,
	}
	requestJSON(client, http.MethodPost, baseURL+"/echo_log/tune", token, tunePayload)
	message, err := ws.readUntilContains(5*time.Second, "tune_echo_log")
	if err != nil {
		fail("websocket did not receive tune message: %v %s", err, message)
	}

	getResp := requestJSON(client, http.MethodGet, baseURL+"/echo_log/"+fmt.Sprint(created.ID), token, nil)
	var getData map[string]any
	mustDecode(getResp, &getData)
	if int64(getData["id"].(float64)) != created.ID {
		fail("get echo_log returned unexpected id")
	}

	findResp := requestJSON(client, http.MethodPost, baseURL+"/echo_log/find?page_size=20", token, map[string]any{
		"id":       created.ID,
		"user_id":  e2eUserID,
		"clazz":    "沉日劫明",
		"keyword":  "",
		"substat1": 8193,
	})
	var found []echoLog
	mustDecode(findResp, &found)
	if len(found) == 0 {
		fail("find echo_log returned empty result")
	}

	requestJSON(client, http.MethodGet, baseURL+"/echo_logs/analysis?size=20&user_id="+fmt.Sprint(e2eUserID)+"&target_bits=3", token, nil)
	requestJSON(client, http.MethodGet, baseURL+"/tune_stats?size=20&user_id="+fmt.Sprint(e2eUserID), token, nil)
	requestJSON(client, http.MethodGet, baseURL+"/counts/echo_dcrit", token, nil)
	requestJSON(client, http.MethodPost, baseURL+"/analyze_echo?resonator=暗主&cost=4C", token, map[string]any{
		"substat1":    8193,
		"substat2":    16386,
		"substat_all": 3,
	})
	requestJSON(client, http.MethodPost, baseURL+"/predict/echo_substat", token, map[string]any{
		"substat1":    8193,
		"substat_all": 1,
	})

	requestJSON(client, http.MethodDelete, baseURL+"/echo_log/"+fmt.Sprint(created.ID), token, nil)
	fmt.Println("E2E OK")
}

func cleanup(databaseURL string, echoID int64) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, _ = db.ExecContext(ctx, "delete from wuwa_tune_log where echo_id = $1", echoID)
	_, _ = db.ExecContext(ctx, "delete from wuwa_echo_log where id = $1", echoID)
}

func requestJSON(client *http.Client, method, endpoint, token string, payload any) json.RawMessage {
	raw := doRequest(client, method, endpoint, token, payload)
	var decoded apiResp
	if err := json.Unmarshal(raw, &decoded); err != nil {
		fail("decode api response for %s %s: %v body=%s", method, endpoint, err, string(raw))
	}
	if decoded.Code != 200 {
		fail("api returned non-200 code for %s %s: %s", method, endpoint, string(raw))
	}
	return decoded.Data
}

func requestRawJSON(client *http.Client, method, endpoint, token string, payload any, target any) {
	raw := doRequest(client, method, endpoint, token, payload)
	if err := json.Unmarshal(raw, target); err != nil {
		fail("decode raw json for %s %s: %v body=%s", method, endpoint, err, string(raw))
	}
}

func doRequest(client *http.Client, method, endpoint, token string, payload any) []byte {
	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			fail("marshal payload: %v", err)
		}
		body = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, endpoint, body)
	if err != nil {
		fail("build request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := client.Do(req)
	if err != nil {
		fail("request %s %s failed: %v", method, endpoint, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		fail("request %s %s failed: status=%d body=%s", method, endpoint, resp.StatusCode, string(raw))
	}
	return raw
}

func mustDecode(data json.RawMessage, target any) {
	if err := json.Unmarshal(data, target); err != nil {
		fail("decode payload: %v body=%s", err, string(data))
	}
}

type wsClient struct {
	conn net.Conn
	br   *bufio.Reader
}

func dialWebsocket(endpoint string) (*wsClient, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	host := u.Host
	if !strings.Contains(host, ":") {
		host += ":80"
	}
	conn, err := net.DialTimeout("tcp", host, 5*time.Second)
	if err != nil {
		return nil, err
	}
	keyBytes := make([]byte, 16)
	if _, err := rand.Read(keyBytes); err != nil {
		conn.Close()
		return nil, err
	}
	key := base64.StdEncoding.EncodeToString(keyBytes)
	path := u.RequestURI()
	req := "GET " + path + " HTTP/1.1\r\n" +
		"Host: " + u.Host + "\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Version: 13\r\n" +
		"Sec-WebSocket-Key: " + key + "\r\n\r\n"
	if _, err := conn.Write([]byte(req)); err != nil {
		conn.Close()
		return nil, err
	}
	br := bufio.NewReader(conn)
	statusLine, err := br.ReadString('\n')
	if err != nil {
		conn.Close()
		return nil, err
	}
	if !strings.Contains(statusLine, "101") {
		conn.Close()
		return nil, fmt.Errorf("unexpected status line: %s", statusLine)
	}
	for {
		line, err := br.ReadString('\n')
		if err != nil {
			conn.Close()
			return nil, err
		}
		if line == "\r\n" {
			break
		}
	}
	return &wsClient{conn: conn, br: br}, nil
}

func (c *wsClient) readText(timeout time.Duration) (string, error) {
	_ = c.conn.SetReadDeadline(time.Now().Add(timeout))
	defer c.conn.SetReadDeadline(time.Time{})
	first, err := c.br.ReadByte()
	if err != nil {
		return "", err
	}
	second, err := c.br.ReadByte()
	if err != nil {
		return "", err
	}
	opcode := first & 0x0f
	length := int(second & 0x7f)
	switch length {
	case 126:
		b1, _ := c.br.ReadByte()
		b2, _ := c.br.ReadByte()
		length = int(b1)<<8 | int(b2)
	case 127:
		var size uint64
		for i := 0; i < 8; i++ {
			b, _ := c.br.ReadByte()
			size = (size << 8) | uint64(b)
		}
		length = int(size)
	}
	payload := make([]byte, length)
	if _, err := io.ReadFull(c.br, payload); err != nil {
		return "", err
	}
	if opcode != 1 {
		return "", fmt.Errorf("unexpected opcode %d", opcode)
	}
	return string(payload), nil
}

func (c *wsClient) readUntilContains(timeout time.Duration, needle string) (string, error) {
	deadline := time.Now().Add(timeout)
	last := ""
	for time.Now().Before(deadline) {
		message, err := c.readText(time.Until(deadline))
		if err != nil {
			return last, err
		}
		last = message
		if strings.Contains(message, needle) {
			return message, nil
		}
	}
	return last, fmt.Errorf("timeout waiting for %q", needle)
}

func (c *wsClient) Close() error { return c.conn.Close() }

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
