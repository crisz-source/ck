package notify

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// PodEvent representa um evento de pod pra notificação
type PodEvent struct {
	PodName      string
	Namespace    string
	EventType    string // "RESTART", "CRASHLOOP", "FATAL", "OOM_KILLED"
	RestartCount int32
	Reason       string
	Timestamp    time.Time
}

func SendEmail(event PodEvent) error {
	connStr := viper.GetString("notify.email.connection_string")
	from := viper.GetString("notify.email.from")
	to := viper.GetString("notify.email.to")

	if connStr == "" || from == "" || to == "" {
		return fmt.Errorf("email não configurado no ~/.ck.yaml (connection_string, from, to)")
	}

	// Parseia a connection string pra extrair endpoint e accesskey
	endpoint, accessKey, err := parseConnectionString(connStr)
	if err != nil {
		return fmt.Errorf("connection string inválida: %w", err)
	}

	// Monta o corpo do email
	subject, body := formatEmail(event)

	// Cria o payload da API
	payload := map[string]interface{}{
		"senderAddress": from,
		"recipients": map[string]interface{}{
			"to": []map[string]string{
				{"address": to},
			},
		},
		"content": map[string]string{
			"subject":   subject,
			"plainText": body,
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("erro ao criar JSON: %w", err)
	}

	// Envia via REST API com autenticação HMAC
	apiURL := endpoint + "/emails:send?api-version=2023-03-31"
	err = sendWithHMAC(apiURL, accessKey, jsonPayload)
	if err != nil {
		return fmt.Errorf("erro ao enviar email: %w", err)
	}

	return nil
}

// ────────────────────────────────────────────────────────────
// AUTENTICAÇÃO HMAC-SHA256 (padrão Azure Communication Services)
// ────────────────────────────────────────────────────────────
//
// Azure CS não usa Bearer token nem API key simples.
// Usa HMAC-SHA256: assina cada request com a access key.
// É o mesmo padrão do Azure Storage.
//
// O fluxo:
// 1. Calcula SHA256 do body (content hash)
// 2. Monta string-to-sign com data, host e content hash
// 3. Assina com HMAC-SHA256 usando a access key
// 4. Coloca tudo nos headers

func sendWithHMAC(apiURL string, accessKey string, jsonPayload []byte) error {
	// Parseia a URL pra extrair host e path
	parsedURL, err := url.Parse(apiURL)
	if err != nil {
		return fmt.Errorf("URL inválida: %w", err)
	}

	host := parsedURL.Host
	pathAndQuery := parsedURL.RequestURI()

	// 1. Content hash (SHA256 do body em base64)
	contentHash := computeContentHash(jsonPayload)

	// 2. Timestamp no formato RFC1123 (exigido pelo Azure)
	timestamp := time.Now().UTC().Format(http.TimeFormat)

	// 3. String-to-sign (o que vai ser assinado)
	//    Formato: "VERB\npath-and-query\ndate;host;content-hash"
	stringToSign := fmt.Sprintf("POST\n%s\n%s;%s;%s",
		pathAndQuery, timestamp, host, contentHash)

	// 4. Assina com HMAC-SHA256
	signature, err := computeSignature(stringToSign, accessKey)
	if err != nil {
		return err
	}

	// 5. Monta o request com todos os headers
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-ms-date", timestamp)
	req.Header.Set("x-ms-content-sha256", contentHash)
	req.Header.Set("Authorization",
		fmt.Sprintf("HMAC-SHA256 SignedHeaders=x-ms-date;host;x-ms-content-sha256&Signature=%s", signature))

	// 6. Envia
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("erro na requisição: %w", err)
	}
	defer resp.Body.Close()

	// 7. Verifica resposta (202 = aceito pra envio)
	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Azure API retornou status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// computeContentHash calcula o SHA256 do body e retorna em base64
func computeContentHash(content []byte) string {
	hash := sha256.Sum256(content)
	return base64.StdEncoding.EncodeToString(hash[:])
}

// computeSignature assina a string com HMAC-SHA256 usando a access key
func computeSignature(stringToSign string, accessKey string) (string, error) {
	// A access key vem em base64, precisa decodificar
	decodedKey, err := base64.StdEncoding.DecodeString(accessKey)
	if err != nil {
		return "", fmt.Errorf("erro ao decodificar access key: %w", err)
	}

	// HMAC-SHA256
	mac := hmac.New(sha256.New, decodedKey)
	mac.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return signature, nil
}

// ────────────────────────────────────────────────────────────
// HELPERS
// ────────────────────────────────────────────────────────────

// parseConnectionString extrai endpoint e accesskey da connection string do Azure
// Formato: "endpoint=https://xxx.communication.azure.com/;accesskey=BASE64KEY"
func parseConnectionString(connStr string) (string, string, error) {
	var endpoint, accessKey string

	parts := strings.Split(connStr, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "endpoint=") {
			endpoint = strings.TrimPrefix(part, "endpoint=")
			endpoint = strings.TrimRight(endpoint, "/")
		} else if strings.HasPrefix(part, "accesskey=") {
			accessKey = strings.TrimPrefix(part, "accesskey=")
		}
	}

	if endpoint == "" || accessKey == "" {
		return "", "", fmt.Errorf("connection string deve ter endpoint= e accesskey=")
	}

	return endpoint, accessKey, nil
}

// formatEmail cria o subject e body do email de alerta
func formatEmail(event PodEvent) (string, string) {
	emoji := "⚠️"
	switch event.EventType {
	case "CRASHLOOP":
		emoji = "🔴"
	case "OOM_KILLED":
		emoji = "💀"
	case "RESTART":
		emoji = "🔄"
	case "FATAL":
		emoji = "❌"
	}

	subject := fmt.Sprintf("%s CK Alert: %s - %s/%s",
		emoji, event.EventType, event.Namespace, event.PodName)

	body := fmt.Sprintf(
		"CK ALERT\n"+
			"════════════════════════════════\n\n"+
			"Evento:     %s\n"+
			"Pod:        %s\n"+
			"Namespace:  %s\n"+
			"Restarts:   %d\n"+
			"Motivo:     %s\n"+
			"Hora:       %s\n\n"+
			"════════════════════════════════\n"+
			"Enviado por ck watch",
		event.EventType,
		event.PodName,
		event.Namespace,
		event.RestartCount,
		event.Reason,
		event.Timestamp.Format("02/01/2006 15:04:05"),
	)

	return subject, body
}

// formatMessage cria o texto formatado pra terminal
func formatMessage(event PodEvent) string {
	emoji := "⚠️"
	switch event.EventType {
	case "CRASHLOOP":
		emoji = "🔴"
	case "OOM_KILLED":
		emoji = "💀"
	case "RESTART":
		emoji = "🔄"
	case "FATAL":
		emoji = "❌"
	}

	return fmt.Sprintf(
		"%s CK ALERT\n\n"+
			"Evento:     %s\n"+
			"Pod:        %s\n"+
			"Namespace:  %s\n"+
			"Restarts:   %d\n"+
			"Motivo:     %s\n"+
			"Hora:       %s",
		emoji,
		event.EventType,
		event.PodName,
		event.Namespace,
		event.RestartCount,
		event.Reason,
		event.Timestamp.Format("02/01/2006 15:04:05"),
	)
}

// PrintAlert mostra o alerta no terminal (sempre funciona, mesmo sem email)
func PrintAlert(event PodEvent) {
	msg := formatMessage(event)
	fmt.Println(msg)
	fmt.Println("─────────────────────────────────")
}
