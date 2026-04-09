package goapp

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
)

func (wsm *wsManager) handle(operatorID string, conn *wsConn) {
	wsm.mu.Lock()
	if _, ok := wsm.connections[operatorID]; !ok {
		wsm.connections[operatorID] = map[*wsConn]struct{}{}
	}
	wsm.connections[operatorID][conn] = struct{}{}
	wsm.mu.Unlock()
	defer func() {
		wsm.mu.Lock()
		delete(wsm.connections[operatorID], conn)
		if len(wsm.connections[operatorID]) == 0 {
			delete(wsm.connections, operatorID)
		}
		wsm.mu.Unlock()
		_ = conn.conn.Close()
	}()
	for {
		message, err := conn.readLoop()
		if err != nil {
			if err != io.EOF {
				log.Printf("websocket receive: %v", err)
			}
			return
		}
		if message == "" {
			continue
		}
		wsm.handleClientMessage(operatorID, message)
	}
}

func (wsm *wsManager) send(operatorID int64, payload any) {
	wsm.sendToOperator(strconv.FormatInt(operatorID, 10), payload)
}

func (wsm *wsManager) sendToOperator(operatorID string, payload any) {
	body, err := json.Marshal(payload)
	if err != nil {
		return
	}
	wsm.mu.RLock()
	defer wsm.mu.RUnlock()
	for conn := range wsm.connections[operatorID] {
		_ = conn.writeText(string(body))
	}
}

func (wsm *wsManager) handleClientMessage(operatorID string, raw string) {
	var envelope struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal([]byte(raw), &envelope); err != nil {
		return
	}
	switch envelope.Type {
	case "score_template_changed":
		var payload struct {
			Field     string `json:"field"`
			Value     string `json:"value"`
			Resonator string `json:"resonator"`
			Cost      string `json:"cost"`
		}
		if err := json.Unmarshal(envelope.Data, &payload); err != nil {
			return
		}
		wsm.sendToOperator(operatorID, map[string]any{
			"type": "score_template_changed",
			"data": payload,
		})
	}
}

func (a *App) scanEchoLog(row pgx.Row) (*EchoLog, error) {
	var item EchoLog
	if err := row.Scan(&item.ID, &item.Substat1, &item.Substat2, &item.Substat3, &item.Substat4, &item.Substat5, &item.SubstatAll, &item.S1Desc, &item.S2Desc, &item.S3Desc, &item.S4Desc, &item.S5Desc, &item.Clazz, &item.UserID, &item.OperatorID, &item.Deleted, &item.TunedAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, err
	}
	return &item, nil
}

func (a *App) scanEchoLogs(rows pgx.Rows) ([]EchoLog, error) {
	var items []EchoLog
	for rows.Next() {
		item, err := a.scanEchoLog(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return items, rows.Err()
}

func (a *App) handleProxyLogin(w http.ResponseWriter, r *http.Request) {
	a.proxyAuthRequest(w, r, http.MethodPost, "/api/login", r.Body)
}

func (a *App) handleProxyMe(w http.ResponseWriter, r *http.Request) {
	token := extractToken(r)
	if token == "" {
		writeJSONWithStatus(w, http.StatusUnauthorized, map[string]string{"detail": authInvalidDetail})
		return
	}
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, a.authURL+"/api/me", nil)
	if err != nil {
		writeJSONWithStatus(w, http.StatusServiceUnavailable, map[string]string{"detail": authUnavailableDetail})
		return
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := a.httpClient.Do(req)
	if err != nil {
		writeJSONWithStatus(w, http.StatusServiceUnavailable, map[string]string{"detail": authUnavailableDetail})
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(body)
}

func (a *App) proxyAuthRequest(w http.ResponseWriter, r *http.Request, method, path string, body io.Reader) {
	req, err := http.NewRequestWithContext(r.Context(), method, a.authURL+path, body)
	if err != nil {
		writeJSONWithStatus(w, http.StatusServiceUnavailable, map[string]string{"detail": authUnavailableDetail})
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.httpClient.Do(req)
	if err != nil {
		writeJSONWithStatus(w, http.StatusServiceUnavailable, map[string]string{"detail": authUnavailableDetail})
		return
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(respBody)
}

func (a *App) handleWebsocket(w http.ResponseWriter, r *http.Request) {
	operatorID := strings.TrimSpace(r.URL.Query().Get("operator_id"))
	if operatorID == "" {
		writeJSON(w, appError("operator_id is required", 400))
		return
	}
	conn, err := upgradeWebsocket(w, r)
	if err != nil {
		http.Error(w, "websocket upgrade failed", http.StatusBadRequest)
		return
	}
	a.ws.handle(operatorID, conn)
}

func upgradeWebsocket(w http.ResponseWriter, r *http.Request) (*wsConn, error) {
	if !headerContainsToken(r.Header, "Connection", "Upgrade") || !strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
		return nil, errors.New("not websocket upgrade")
	}
	key := strings.TrimSpace(r.Header.Get("Sec-WebSocket-Key"))
	if key == "" {
		return nil, errors.New("missing websocket key")
	}
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		return nil, errors.New("hijack unsupported")
	}
	conn, buf, err := hijacker.Hijack()
	if err != nil {
		return nil, err
	}
	accept := websocketAccept(key)
	response := "HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: " + accept + "\r\n\r\n"
	if _, err := buf.WriteString(response); err != nil {
		_ = conn.Close()
		return nil, err
	}
	if err := buf.Flush(); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return &wsConn{conn: conn, br: bufio.NewReader(conn)}, nil
}

func websocketAccept(key string) string {
	sum := sha1.Sum([]byte(key + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
	return base64.StdEncoding.EncodeToString(sum[:])
}

func headerContainsToken(header http.Header, key, token string) bool {
	for _, value := range header.Values(key) {
		for _, part := range strings.Split(value, ",") {
			if strings.EqualFold(strings.TrimSpace(part), token) {
				return true
			}
		}
	}
	return false
}

func (c *wsConn) writeText(message string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	payload := []byte(message)
	header := []byte{0x81}
	switch {
	case len(payload) < 126:
		header = append(header, byte(len(payload)))
	case len(payload) <= 65535:
		header = append(header, 126, byte(len(payload)>>8), byte(len(payload)))
	default:
		header = append(header, 127,
			byte(uint64(len(payload))>>56), byte(uint64(len(payload))>>48), byte(uint64(len(payload))>>40), byte(uint64(len(payload))>>32),
			byte(uint64(len(payload))>>24), byte(uint64(len(payload))>>16), byte(uint64(len(payload))>>8), byte(uint64(len(payload))))
	}
	if _, err := c.conn.Write(header); err != nil {
		return err
	}
	_, err := c.conn.Write(payload)
	return err
}

func (c *wsConn) readLoop() (string, error) {
	first, err := c.br.ReadByte()
	if err != nil {
		return "", err
	}
	second, err := c.br.ReadByte()
	if err != nil {
		return "", err
	}
	opcode := first & 0x0f
	masked := second&0x80 != 0
	payloadLen := int(second & 0x7f)
	switch payloadLen {
	case 126:
		b1, err := c.br.ReadByte()
		if err != nil {
			return "", err
		}
		b2, err := c.br.ReadByte()
		if err != nil {
			return "", err
		}
		payloadLen = int(b1)<<8 | int(b2)
	case 127:
		var size uint64
		for i := 0; i < 8; i++ {
			b, err := c.br.ReadByte()
			if err != nil {
				return "", err
			}
			size = (size << 8) | uint64(b)
		}
		payloadLen = int(size)
	}
	var mask [4]byte
	if masked {
		if _, err := io.ReadFull(c.br, mask[:]); err != nil {
			return "", err
		}
	}
	payload := make([]byte, payloadLen)
	if _, err := io.ReadFull(c.br, payload); err != nil {
		return "", err
	}
	if masked {
		for i := range payload {
			payload[i] ^= mask[i%4]
		}
	}
	switch opcode {
	case 0x8:
		return "", io.EOF
	case 0x9:
		return "", c.writeControl(0xA, payload)
	case 0x1:
		return string(payload), nil
	default:
		return "", nil
	}
}

func (c *wsConn) writeControl(opcode byte, payload []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	header := []byte{0x80 | opcode, byte(len(payload))}
	if _, err := c.conn.Write(header); err != nil {
		return err
	}
	_, err := c.conn.Write(payload)
	return err
}
