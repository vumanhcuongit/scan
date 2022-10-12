package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"github.com/vumanhcuongit/scan/internal/config"
	"github.com/vumanhcuongit/scan/internal/services/api"
	"github.com/vumanhcuongit/scan/pkg/models"
)

type handlerSuite struct {
	suite.Suite

	config      *config.App
	mockCtrl    *gomock.Controller
	scanService *api.MockIScanService
	router      *gin.Engine
	handler     *Handler
}

func TestHandlerSuite(t *testing.T) {
	suite.Run(t, &handlerSuite{})
}

func (s *handlerSuite) SetupSuite() {
	cfg, err := config.Load("")
	if err != nil {
		s.Require().NoError(err)
	}
	s.config = cfg
}

func (s *handlerSuite) SetupTest() {
	s.router = gin.New()
	s.mockCtrl = gomock.NewController(s.T())
	s.scanService = api.NewMockIScanService(s.mockCtrl)
	s.handler = NewHandler(s.scanService)
	s.handler.Register(s.router)
	s.handler.SetScanService(s.scanService)
}

func (s *handlerSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *handlerSuite) TestListRepositories() {
	size := 20
	page := 0
	request := &api.ListRepositoriesRequest{
		Size: size,
		Page: page,
	}
	repositories := []*models.Repository{
		{
			ID: 1,
		},
	}
	s.scanService.EXPECT().ListRepositories(gomock.Any(), request).Return(repositories, nil)

	resp := performHandlerRequest(s.router, "GET", "/api/repositories", nil)
	s.Equal(200, resp.Code)
	var respBody struct {
		Data []struct {
			ID            int       `json:"id"`
			Name          string    `json:"name"`
			Owner         string    `json:"owner"`
			RepositoryURL string    `json:"repository_url"`
			CreatedAt     time.Time `json:"created_at"`
			UpdatedAt     time.Time `json:"updated_at"`
		} `json:"data"`
	}
	err := json.Unmarshal(resp.Body.Bytes(), &respBody)
	s.NoError(err)
	s.Require().Equal(1, len(respBody.Data))
}

func (s *handlerSuite) TestListRepositoriesWithError() {
	size := 20
	page := 0
	request := &api.ListRepositoriesRequest{
		Size: size,
		Page: page,
	}
	s.scanService.EXPECT().ListRepositories(gomock.Any(), request).Return(nil, errors.New("failed to list repositories"))

	resp := performHandlerRequest(s.router, "GET", "/api/repositories", nil)
	s.Equal(200, resp.Code)
	var respBody struct {
		Data  interface{} `json:"data"`
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	err := json.Unmarshal(resp.Body.Bytes(), &respBody)
	s.NoError(err)
	s.Require().Nil(respBody.Data)
	s.Require().Equal(500, respBody.Error.Code)
	s.Require().Equal("failed to list repositories", respBody.Error.Message)
}

func (s *handlerSuite) TestCreateRepository() {
	repositoryURL := "https://github.com/vumanhcuongit/scan"
	request := &api.CreateRepositoryRequest{
		RepositoryURL: repositoryURL,
	}
	bodyData, _ := json.Marshal(request)
	bodyReader := bytes.NewReader(bodyData)
	repository := &models.Repository{ID: 1}
	s.scanService.EXPECT().CreateRepository(gomock.Any(), request).Return(repository, nil)

	resp := performHandlerRequest(s.router, "POST", "/api/repositories", bodyReader)
	s.Equal(200, resp.Code)
	var respBody struct {
		Data struct {
			ID            int       `json:"id"`
			Name          string    `json:"name"`
			Owner         string    `json:"owner"`
			RepositoryURL string    `json:"repository_url"`
			CreatedAt     time.Time `json:"created_at"`
			UpdatedAt     time.Time `json:"updated_at"`
		} `json:"data"`
	}
	err := json.Unmarshal(resp.Body.Bytes(), &respBody)
	s.NoError(err)
	s.Require().Equal(1, respBody.Data.ID)
}

func (s *handlerSuite) TestCreateRepositoryWithFailedCreation() {
	repositoryURL := "https://github.com/vumanhcuongit/scan"
	request := &api.CreateRepositoryRequest{
		RepositoryURL: repositoryURL,
	}
	bodyData, _ := json.Marshal(request)
	bodyReader := bytes.NewReader(bodyData)
	s.scanService.EXPECT().CreateRepository(gomock.Any(), request).Return(nil, errors.New("failed to create repository"))

	resp := performHandlerRequest(s.router, "POST", "/api/repositories", bodyReader)
	s.Equal(200, resp.Code)
	var respBody struct {
		Data  interface{} `json:"data"`
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	err := json.Unmarshal(resp.Body.Bytes(), &respBody)
	s.NoError(err)
	s.Require().Equal(500, respBody.Error.Code)
	s.Require().Equal("failed to create repository", respBody.Error.Message)
}

func (s *handlerSuite) TestCreateRepositoryWithInvalidParams() {
	request := &api.CreateRepositoryRequest{
		RepositoryURL: "",
	}
	bodyData, _ := json.Marshal(request)
	bodyReader := bytes.NewReader(bodyData)

	resp := performHandlerRequest(s.router, "POST", "/api/repositories", bodyReader)
	s.Equal(400, resp.Code)
}

func (s *handlerSuite) TestGetRepository() {
	repository := &models.Repository{ID: 1}
	s.scanService.EXPECT().GetRepository(gomock.Any(), repository.ID).Return(repository, nil)

	resp := performHandlerRequest(s.router, "GET", fmt.Sprintf("/api/repositories/%d", repository.ID), nil)
	s.Equal(200, resp.Code)
	var respBody struct {
		Data struct {
			ID            int       `json:"id"`
			Name          string    `json:"name"`
			Owner         string    `json:"owner"`
			RepositoryURL string    `json:"repository_url"`
			CreatedAt     time.Time `json:"created_at"`
			UpdatedAt     time.Time `json:"updated_at"`
		} `json:"data"`
	}
	err := json.Unmarshal(resp.Body.Bytes(), &respBody)
	s.NoError(err)
	s.Equal(1, respBody.Data.ID)
}

func (s *handlerSuite) TestGetRepositoryWithNotFoundRecord() {
	repository := &models.Repository{ID: 1}
	s.scanService.EXPECT().GetRepository(gomock.Any(), repository.ID).Return(nil, errors.New("record not found"))

	resp := performHandlerRequest(s.router, "GET", fmt.Sprintf("/api/repositories/%d", repository.ID), nil)
	s.Equal(200, resp.Code)
	var respBody struct {
		Data  interface{} `json:"data"`
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	err := json.Unmarshal(resp.Body.Bytes(), &respBody)
	s.NoError(err)
	s.Equal(500, respBody.Error.Code)
}

func (s *handlerSuite) TestUpdateRepository() {
	repositoryURL := "https://github.com/vumanhcuongit/scan"
	request := &api.UpdateRepositoryRequest{
		RepositoryURL: repositoryURL,
	}
	bodyData, _ := json.Marshal(request)
	bodyReader := bytes.NewReader(bodyData)
	repository := &models.Repository{ID: 1}
	s.scanService.EXPECT().UpdateRepository(gomock.Any(), repository.ID, gomock.Any()).Return(repository, nil)

	resp := performHandlerRequest(s.router, "PATCH", fmt.Sprintf("/api/repositories/%d", repository.ID), bodyReader)
	s.Equal(200, resp.Code)
	var respBody struct {
		Data struct {
			ID            int       `json:"id"`
			Name          string    `json:"name"`
			Owner         string    `json:"owner"`
			RepositoryURL string    `json:"repository_url"`
			CreatedAt     time.Time `json:"created_at"`
			UpdatedAt     time.Time `json:"updated_at"`
		} `json:"data"`
	}
	err := json.Unmarshal(resp.Body.Bytes(), &respBody)
	s.NoError(err)
	s.Require().Equal(1, respBody.Data.ID)
}

func (s *handlerSuite) TestUpdateRepositoryWithFailedUpdation() {
	repositoryURL := "https://github.com/vumanhcuongit/scan"
	request := &api.UpdateRepositoryRequest{
		RepositoryURL: repositoryURL,
	}
	bodyData, _ := json.Marshal(request)
	bodyReader := bytes.NewReader(bodyData)
	repository := &models.Repository{ID: 1}
	s.scanService.EXPECT().UpdateRepository(gomock.Any(), repository.ID, gomock.Any()).
		Return(nil, errors.New("failed to update"))

	resp := performHandlerRequest(s.router, "PATCH", fmt.Sprintf("/api/repositories/%d", repository.ID), bodyReader)
	s.Equal(200, resp.Code)
	var respBody struct {
		Data  interface{} `json:"data"`
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	err := json.Unmarshal(resp.Body.Bytes(), &respBody)
	s.NoError(err)
	s.Equal(500, respBody.Error.Code)
}

func (s *handlerSuite) TestDeleteRepository() {
	repository := &models.Repository{ID: 1}
	s.scanService.EXPECT().DeleteRepository(gomock.Any(), repository.ID).Return(nil)

	resp := performHandlerRequest(s.router, "DELETE", fmt.Sprintf("/api/repositories/%d", repository.ID), nil)
	s.Equal(204, resp.Code)
}

func (s *handlerSuite) TestDeleteRepositoryWithFailedDeletion() {
	repository := &models.Repository{ID: 1}
	s.scanService.EXPECT().DeleteRepository(gomock.Any(), repository.ID).Return(errors.New("failed to delete"))

	resp := performHandlerRequest(s.router, "DELETE", fmt.Sprintf("/api/repositories/%d", repository.ID), nil)
	s.Equal(200, resp.Code)
	var respBody struct {
		Data  interface{} `json:"data"`
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	err := json.Unmarshal(resp.Body.Bytes(), &respBody)
	s.NoError(err)
	s.Equal(500, respBody.Error.Code)
}

func (s *handlerSuite) TestCreateScan() {
	request := &api.TriggerScanRequest{
		RepositoryID: 1,
	}
	bodyData, _ := json.Marshal(request)
	bodyReader := bytes.NewReader(bodyData)
	scan := &models.Scan{ID: 1}
	s.scanService.EXPECT().TriggerScan(gomock.Any(), request).Return(scan, nil)

	resp := performHandlerRequest(s.router, "POST", "/api/scans", bodyReader)
	s.Equal(200, resp.Code)
	var respBody struct {
		Data struct {
			ID int `json:"id"`
		} `json:"data"`
	}
	err := json.Unmarshal(resp.Body.Bytes(), &respBody)
	s.NoError(err)
	s.Require().Equal(1, respBody.Data.ID)
}

func (s *handlerSuite) TestCreateScanWithFailedTrigger() {
	request := &api.TriggerScanRequest{
		RepositoryID: 1,
	}
	bodyData, _ := json.Marshal(request)
	bodyReader := bytes.NewReader(bodyData)
	s.scanService.EXPECT().TriggerScan(gomock.Any(), request).Return(nil, errors.New("failed to create scan"))

	resp := performHandlerRequest(s.router, "POST", "/api/scans", bodyReader)
	s.Equal(200, resp.Code)
	var respBody struct {
		Data  interface{} `json:"data"`
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	err := json.Unmarshal(resp.Body.Bytes(), &respBody)
	s.NoError(err)
	s.Require().Equal(500, respBody.Error.Code)
	s.Require().Equal("failed to create scan", respBody.Error.Message)
}

func (s *handlerSuite) TestCreateScanWithInvalidParams() {
	request := &api.TriggerScanRequest{}
	bodyData, _ := json.Marshal(request)
	bodyReader := bytes.NewReader(bodyData)

	resp := performHandlerRequest(s.router, "POST", "/api/scans", bodyReader)
	s.Equal(400, resp.Code)
}

func (s *handlerSuite) TestListScans() {
	size := 20
	page := 0
	request := &api.ListScansRequest{
		Size: size,
		Page: page,
	}
	scans := []*models.Scan{
		{
			ID: 1,
		},
	}
	s.scanService.EXPECT().ListScans(gomock.Any(), request).Return(scans, nil)

	resp := performHandlerRequest(s.router, "GET", "/api/scans", nil)
	s.Equal(200, resp.Code)
	var respBody struct {
		Data []struct {
			ID int `json:"id"`
		} `json:"data"`
	}
	err := json.Unmarshal(resp.Body.Bytes(), &respBody)
	s.NoError(err)
	s.Require().Equal(1, len(respBody.Data))
}

func (s *handlerSuite) TestListScansWithError() {
	size := 20
	page := 0
	request := &api.ListScansRequest{
		Size: size,
		Page: page,
	}
	s.scanService.EXPECT().ListScans(gomock.Any(), request).Return(nil, errors.New("failed to list scans"))

	resp := performHandlerRequest(s.router, "GET", "/api/scans", nil)
	s.Equal(200, resp.Code)
	var respBody struct {
		Data  interface{} `json:"data"`
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	err := json.Unmarshal(resp.Body.Bytes(), &respBody)
	s.NoError(err)
	s.Require().Nil(respBody.Data)
	s.Require().Equal(500, respBody.Error.Code)
	s.Require().Equal("failed to list scans", respBody.Error.Message)
}

func performHandlerRequest(h http.Handler, method string, path string, body io.Reader) *httptest.ResponseRecorder {
	r, _ := http.NewRequest(method, path, body)
	r.Header.Add("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}
