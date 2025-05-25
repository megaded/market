package manager

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/megaded/market/cmd/internal/dto"
)

type AccrualClient struct {
	client  *http.Client
	baseURL string
}

type AccrualInfo struct {
	Response dto.Accrual
	Retry    int
	Error    error
}

func CreateAccrualClient(baseURL string) AccrualClient {
	return AccrualClient{baseURL: baseURL, client: &http.Client{}}
}

func (c *AccrualClient) GetOrderStatus(ctx context.Context, orderNumber string) AccrualInfo {
	accrualResponse := AccrualInfo{}
	path, err := url.JoinPath(c.baseURL, "api", "orders", orderNumber)
	if err != nil {
		accrualResponse.Error = err
		return accrualResponse
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		accrualResponse.Error = err
		return accrualResponse
	}

	response, err := c.client.Do(request)
	if err != nil {
		accrualResponse.Error = err
		return accrualResponse
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusTooManyRequests {
		retryAfterStr := response.Header.Get("Retry-After")
		seconds, err := strconv.Atoi(retryAfterStr)
		if err != nil {
			accrualResponse.Error = err
			return accrualResponse
		}
		accrualResponse.Retry = seconds
		return accrualResponse
	}

	if response.StatusCode == http.StatusNoContent {
		return accrualResponse
	}

	if response.StatusCode != http.StatusOK {
		accrualResponse.Error = errors.New("failed to get order info")
		return accrualResponse
	}

	var accrualInfo *dto.Accrual
	if err = json.NewDecoder(response.Body).Decode(&accrualInfo); err != nil {
		accrualResponse.Error = err
		return accrualResponse
	}
	accrualResponse.Response = *accrualInfo

	return accrualResponse
}
