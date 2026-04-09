package goapp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func (a *App) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func loadEnv(path string) {
	content, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, raw := range strings.Split(string(content), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" || os.Getenv(key) != "" {
			continue
		}
		if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'')) {
			value = value[1 : len(value)-1]
		}
		_ = os.Setenv(key, value)
	}
}

func (a *App) withPermission(required string, next func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		info, err := a.validateRequest(r, required)
		if err != nil {
			a.writeAuthError(w, err)
			return
		}
		ctx := context.WithValue(r.Context(), authInfoKey, info)
		next(w, r.WithContext(ctx))
	}
}

func (a *App) validateRequest(r *http.Request, required string) (*AuthInfo, error) {
	token := extractToken(r)
	if token == "" {
		return nil, statusError{status: http.StatusUnauthorized, detail: authInvalidDetail}
	}
	payload := map[string]any{"token": token}
	if required != "" {
		payload["permission"] = required
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, a.authURL+"/api/validate", bytes.NewReader(body))
	if err != nil {
		return nil, statusError{status: http.StatusServiceUnavailable, detail: authUnavailableDetail}
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, statusError{status: http.StatusServiceUnavailable, detail: authUnavailableDetail}
	}
	defer resp.Body.Close()

	var result struct {
		Valid       bool     `json:"valid"`
		Reason      string   `json:"reason"`
		Permissions []string `json:"permissions"`
		ID          *int64   `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, statusError{status: http.StatusServiceUnavailable, detail: authUnavailableDetail}
	}
	if resp.StatusCode != http.StatusOK {
		return nil, statusError{status: http.StatusServiceUnavailable, detail: authUnavailableDetail}
	}
	if !result.Valid {
		if strings.EqualFold(result.Reason, "forbidden") {
			return nil, statusError{status: http.StatusForbidden, detail: authForbiddenDetail}
		}
		return nil, statusError{status: http.StatusUnauthorized, detail: authInvalidDetail}
	}
	if required != "" && !hasPermission(result.Permissions, required) {
		return nil, statusError{status: http.StatusForbidden, detail: authForbiddenDetail}
	}
	return &AuthInfo{Permissions: result.Permissions, OperatorID: result.ID}, nil
}

type statusError struct {
	status int
	detail string
}

func (e statusError) Error() string { return e.detail }

func (a *App) writeAuthError(w http.ResponseWriter, err error) {
	var se statusError
	if errors.As(err, &se) {
		writeJSONWithStatus(w, se.status, map[string]string{"detail": se.detail})
		return
	}
	writeJSONWithStatus(w, http.StatusServiceUnavailable, map[string]string{"detail": authUnavailableDetail})
}

func extractToken(r *http.Request) string {
	if auth := r.Header.Get("Authorization"); auth != "" {
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") && strings.TrimSpace(parts[1]) != "" {
			return strings.TrimSpace(parts[1])
		}
	}
	if token := strings.TrimSpace(r.Header.Get("X-Token")); token != "" {
		return token
	}
	return strings.TrimSpace(r.URL.Query().Get("token"))
}

func hasPermission(permissions []string, required string) bool {
	for _, p := range permissions {
		if p == "manage" || p == required {
			return true
		}
	}
	return false
}

func authInfoFromContext(ctx context.Context) *AuthInfo {
	info, _ := ctx.Value(authInfoKey).(*AuthInfo)
	return info
}

func operatorIDFromContext(ctx context.Context) *int64 {
	info := authInfoFromContext(ctx)
	if info == nil {
		return nil
	}
	return info.OperatorID
}

func canManage(ctx context.Context) bool {
	info := authInfoFromContext(ctx)
	if info == nil {
		return false
	}
	return hasPermission(info.Permissions, "manage")
}

func writeJSON(w http.ResponseWriter, payload any) {
	writeJSONWithStatus(w, http.StatusOK, payload)
}

func writeJSONWithStatus(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func success(message string, data any) SuccessResponse {
	return SuccessResponse{Code: 200, Message: message, Data: data}
}

func appError(message string, code int) ErrorResponse {
	return ErrorResponse{Code: code, Message: message}
}

func page(message string, data any, total int64, pageNum, pageSize int) PageResponse {
	pageTotal := 0
	if pageSize > 0 {
		pageTotal = int(total/int64(pageSize)) + 1
	}
	return PageResponse{Code: 200, Message: message, Data: data, DataTotal: total, Page: pageNum, PageSize: pageSize, PageTotal: pageTotal}
}

func readJSON(r *http.Request, target any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}

func parseIntDefault(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	if value, err := strconv.Atoi(raw); err == nil {
		return value
	}
	return fallback
}

func parseInt64Default(raw string, fallback int64) int64 {
	if raw == "" {
		return fallback
	}
	if value, err := strconv.ParseInt(raw, 10, 64); err == nil {
		return value
	}
	return fallback
}

func parseStatsWindow(raw string) statsWindow {
	switch strings.TrimSpace(raw) {
	case "", "all":
		return statsWindow{Name: "all"}
	case "last_100":
		return statsWindow{Name: "last_100", Limit: 100}
	case "last_500":
		return statsWindow{Name: "last_500", Limit: 500}
	case "last_1000":
		return statsWindow{Name: "last_1000", Limit: 1000}
	case "day_7":
		return statsWindow{Name: "day_7", Days: 7}
	case "day_30":
		return statsWindow{Name: "day_30", Days: 30}
	default:
		return statsWindow{Name: "all"}
	}
}

func (w statsWindow) isAll() bool {
	return w.Name == "" || w.Name == "all"
}

func (w statsWindow) applyLimit(size int) int {
	if w.Limit > 0 && size > 0 {
		if size < w.Limit {
			return size
		}
		return w.Limit
	}
	if w.Limit > 0 {
		return w.Limit
	}
	return size
}

func (w statsWindow) sinceTime() *time.Time {
	if w.Days <= 0 {
		return nil
	}
	t := time.Now().AddDate(0, 0, -w.Days)
	return &t
}
