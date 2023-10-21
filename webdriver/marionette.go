package webdriver

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"

	cm "github.com/lanseg/golang-commons/common"
	"github.com/lanseg/golang-commons/optional"
)

type Response struct {
	MessageId int
	Error     *DriverError
	Result    interface{}
}

func getStringValue(r *Response) string {
	if r == nil {
		return ""
	}
	result, ok := r.Result.(map[string]interface{})["value"]
	if !ok || result == nil {
		return ""
	}
	return fmt.Sprintf("%v", result)
}

type DriverError struct {
	Error      string `json:"error"`
	Message    string `json:"message"`
	Stacktrace string `json:"stacktrace"`
}

func parseError(raw interface{}) *DriverError {
	if raw == nil {
		return nil
	}
	result := raw.(map[string]interface{})
	return &DriverError{
		Error:      fmt.Sprintf("%v", result["error"]),
		Message:    fmt.Sprintf("%v", result["message"]),
		Stacktrace: fmt.Sprintf("%v", result["stacktrace"]),
	}
}

type DriverVersion struct {
	ApplicationType    string `json:"applicationType"`
	MarionetteProtocol int    `json:"marionetteVersion"`
}

type SessionParams struct {
	SessionId    string            `json:"sessionId"`
	Capabilities map[string]string `json:"capabilities"`
}

type Marionette struct {
	WebDriver

	logger *cm.Logger

	pendingCommands map[int](chan optional.Optional[*Response])
	driverVersion   *DriverVersion
	sessionParams   *SessionParams
	messageCounter  int
	connection      net.Conn
}

func (m *Marionette) reader() {
	conn := m.connection
	for {
		onebyte := make([]byte, 1)
		prefix := []byte{}
		for {
			bytes, err := conn.Read(onebyte)
			if bytes == 0 {
				continue
			}
			if err != nil {
				return
			}
			if onebyte[0] == ':' {
				break
			}
			prefix = append(prefix, onebyte...)
		}
		payloadSize, err := strconv.Atoi(string(prefix))
		m.logger.Debugf("Payload size: %03d", payloadSize)
		if err != nil {
			m.logger.Errorf("Incorrect payload size: %s", err)
			break
		}
		payload := []byte{}
		payloadBuffer := make([]byte, 65536)
		for len(payload) < payloadSize {
			bytesRead, err := conn.Read(payloadBuffer)
			if err != nil {
				m.logger.Errorf("Error while reading payload: %s", err)
				break
			}
			payload = append(payload, payloadBuffer[:bytesRead]...)
		}
		m.logger.Infof("Read [%d of %d] bytes of payload", len(payload), payloadSize)
		m.onDataRead(payload)
	}
}

func (m *Marionette) onDataRead(data []byte) {
	if m.driverVersion == nil {
		m.driverVersion = &DriverVersion{}
		json.Unmarshal(data, m.driverVersion)
		m.logger.Infof("Driver version: %v", m.driverVersion)
		return
	}
	var result []interface{}
	json.Unmarshal(data, &result)

	if m.sessionParams == nil {
		m.sessionParams = &SessionParams{
			Capabilities: map[string]string{},
		}
		sessionParams := result[3].(map[string]interface{})
		m.sessionParams.SessionId = sessionParams["sessionId"].(string)
		for k, v := range sessionParams["capabilities"].(map[string]interface{}) {
			m.sessionParams.Capabilities[k] = fmt.Sprintf("%v", v)
		}
		m.pendingCommands[1] <- optional.OfNullable[Response](nil)
		return
	}

	id := int(result[1].(float64))
	ch := m.pendingCommands[id]
	if err := parseError(result[2]); err != nil {
		ch <- optional.OfErrorNullable[Response](nil, fmt.Errorf("%v", err))
		return
	}
	response := &Response{MessageId: id, Result: result[3]}
	m.logger.Infof("Response id:%d err:%v result:%p", response.MessageId, response.Error, result[3])
	ch <- optional.OfNullable(response)
}

func (m *Marionette) write(command string, params any) optional.Optional[*Response] {
	ch := make(chan optional.Optional[*Response])
	m.pendingCommands[m.messageCounter] = ch
	optional.OfError(json.Marshal([]any{0, m.messageCounter, command, params})).
		IfPresent(func(bytes []byte) {
			msg := fmt.Sprintf("%d:%s", len(bytes), string(bytes))
			m.logger.Infof("Writing json: %s", msg)
			m.connection.Write([]byte(msg))
		})
	m.messageCounter++
	result := <-ch
	close(ch)
	return result
}

func (m *Marionette) NewSession() {
	m.write("WebDriver:NewSession", struct{}{})
}

func (m *Marionette) Navigate(url string) {
	m.write("WebDriver:Navigate", struct {
		Url string `json:"url"`
	}{Url: url})
}

func (m *Marionette) GetPageSource() optional.Optional[string] {
	return optional.Map(m.write("WebDriver:GetPageSource", struct{}{}), getStringValue)
}

func (m *Marionette) TakeScreenshot() optional.Optional[string] {
	return optional.Map(m.write("WebDriver:TakeScreenshot", struct{}{}), getStringValue)
}

func (m *Marionette) Print() optional.Optional[string] {
	return optional.Map(m.write("WebDriver:Print", struct{}{}), getStringValue)
}

func (m *Marionette) ExecuteScript(script string) optional.Optional[string] {
	return optional.Map(m.write("WebDriver:ExecuteScript",
		struct {
			Script string `json:"script"`
		}{Script: script}),
		getStringValue)
}

func (m *Marionette) Close() {
	m.connection.Close()
	m.logger.Infof("Disconnected")
}

func ConnectMarionette(host string, port int) optional.Optional[WebDriver] {
	return optional.Map(
		optional.OfError[net.Conn](net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))),
		func(c net.Conn) WebDriver {
			logger := cm.NewLogger("Marionette")
			logger.Infof("Connected to Marionette at %s:%s", host, port)
			marionette := &Marionette{
				messageCounter:  1,
				logger:          logger,
				connection:      c,
				pendingCommands: map[int](chan optional.Optional[*Response]){},
			}
			go marionette.reader()
			return marionette
		})
}
