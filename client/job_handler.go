package client

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/gurodrigues-dev/tcp-server-client/server"
)

func (c *TCPClient) StartJobHandler() {
	go func() {
		for params := range c.messageChan {
			paramsJSON, err := json.Marshal(params)
			if err != nil {
				continue
			}

			var jobParams server.JobParams
			if err := json.Unmarshal(paramsJSON, &jobParams); err != nil {
				continue
			}

			c.mu.Lock()
			c.serverNonce = jobParams.ServerNonce
			c.jobID = jobParams.JobID
			c.mu.Unlock()

			go c.calculateAndSubmit()
		}
	}()
}

func (c *TCPClient) calculateAndSubmit() {
	c.mu.Lock()
	serverNonce := c.serverNonce
	jobID := c.jobID
	c.mu.Unlock()

	clientNonce := c.generateNonce()

	result := server.ComputeSHA256(serverNonce, clientNonce)

	c.rateLimiter.Wait()

	submissionParams := server.SubmissionParams{
		JobID:       jobID,
		ClientNonce: clientNonce,
		Result:      result,
	}

	request := &server.Request{
		Method: "submit",
		Params: submissionParams,
	}

	response, err := c.SendRequestSync(request)
	if err != nil {
		fmt.Printf("Erro ao enviar submissão: %v\n", err)
		return
	}

	if response.Error != "" {
		fmt.Printf("Erro na resposta do servidor: %s\n", response.Error)
	} else {
		fmt.Printf("Submissão recebida com sucesso: %+v\n", response.Result)
	}
}

func (c *TCPClient) generateNonce() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("%x", 123456789)
	}
	return hex.EncodeToString(bytes)
}
