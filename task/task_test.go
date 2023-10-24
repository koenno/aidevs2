package task

import (
	"errors"
	"net/http"
	"testing"

	"github.com/koenno/aidevs2/task/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type FetcherTestSuite struct {
	suite.Suite
	clientMock     *mocks.Client
	reqCreatorMock *mocks.TaskRequestCreator
	sut            Fetcher
	taskName       string
}

func TestFetcherTestSuite(t *testing.T) {
	suite.Run(t, new(FetcherTestSuite))
}

func (s *FetcherTestSuite) SetupTest() {
	s.clientMock = mocks.NewClient(s.T())
	s.reqCreatorMock = mocks.NewTaskRequestCreator(s.T())
	s.taskName = "someTask"
	s.sut = Fetcher{
		ApiKey:  "jfhfiwu763gw",
		Client:  s.clientMock,
		Creator: s.reqCreatorMock,
	}
}

type TestTaskData struct {
	Response
	Token string
	Name  string
}

func (p *TestTaskData) GetCode() int {
	return p.Code
}

func (p *TestTaskData) GetMsg() string {
	return p.Msg
}

func (p *TestTaskData) SetToken(token string) {
	p.Token = token
}

func (s *FetcherTestSuite) TestShouldReturnErrorWhenAuthRequestCreationFails() {
	// given
	var payload TestTaskData

	expectedErr := errors.New("fatal failure")
	s.reqCreatorMock.EXPECT().Authenticate(s.sut.ApiKey, s.taskName).Return(nil, expectedErr).Once()

	// when
	err := s.sut.Fetch(s.taskName, &payload)

	// then
	s.ErrorIs(err, errAuth)
	s.Zero(payload)
}

func (s *FetcherTestSuite) TestShouldReturnErrorWhenSendingRequestFails() {
	// given
	var payload TestTaskData

	req, _ := http.NewRequest(http.MethodGet, "", nil)
	s.reqCreatorMock.EXPECT().Authenticate(s.sut.ApiKey, s.taskName).Return(req, nil).Once()

	expectedErr := errors.New("fatal failure")
	s.clientMock.EXPECT().Send(req, mock.Anything).Return(expectedErr).Once()

	// when
	err := s.sut.Fetch(s.taskName, &payload)

	// then
	s.ErrorIs(err, errAuth)
	s.Zero(payload)
}

func (s *FetcherTestSuite) TestShouldReturnErrorWhenAuthorizationResponseCodeIsNotZero() {
	// given
	var payload TestTaskData

	req, _ := http.NewRequest(http.MethodGet, "", nil)
	s.reqCreatorMock.EXPECT().Authenticate(s.sut.ApiKey, s.taskName).Return(req, nil).Once()

	s.clientMock.EXPECT().Send(req, mock.Anything).Run(func(r *http.Request, respPayload interface{}) {
		respPayload = Response{
			Code: -1,
			Msg:  "failure",
		}
	}).Return(nil).Once()

	// when
	err := s.sut.Fetch(s.taskName, &payload)

	// then
	s.ErrorIs(err, errAuth)
	s.Zero(payload)
}

func (s *FetcherTestSuite) TestShouldReturnErrorWhenAuthorizationResponseTokenIsEmpty() {
	// given
	var payload TestTaskData

	req, _ := http.NewRequest(http.MethodGet, "", nil)
	s.reqCreatorMock.EXPECT().Authenticate(s.sut.ApiKey, s.taskName).Return(req, nil).Once()

	s.clientMock.EXPECT().Send(req, mock.Anything).Run(func(r *http.Request, respPayload interface{}) {
		resp, ok := respPayload.(*AuthorizationResponse)
		s.True(ok)
		resp.Response = Response{
			Code: 0,
			Msg:  "OK",
		}
		resp.Token = ""
	}).Return(nil).Once()

	// when
	err := s.sut.Fetch(s.taskName, &payload)

	// then
	s.ErrorIs(err, errAuth)
	s.Zero(payload)
}

func (s *FetcherTestSuite) TestShouldFetchTask() {
	// given
	var payload TestTaskData
	token := "rghetrgt67ih"

	reqAuth, _ := http.NewRequest(http.MethodPost, "", nil)
	s.reqCreatorMock.EXPECT().Authenticate(s.sut.ApiKey, s.taskName).Return(reqAuth, nil).Once()

	s.clientMock.EXPECT().Send(reqAuth, mock.Anything).Run(func(r *http.Request, respPayload interface{}) {
		resp, ok := respPayload.(*AuthorizationResponse)
		s.True(ok)
		resp.Response = Response{
			Code: 0,
			Msg:  "OK",
		}
		resp.Token = token
	}).Return(nil).Once()

	reqTask, _ := http.NewRequest(http.MethodGet, "", nil)
	s.reqCreatorMock.EXPECT().Task(token).Return(reqTask, nil).Once()

	expectedTaskData := TestTaskData{
		Response: Response{
			Code: 0,
			Msg:  "OK",
		},
		Name:  "some name",
		Token: token,
	}
	s.clientMock.EXPECT().Send(reqTask, mock.Anything).Run(func(r *http.Request, respPayload interface{}) {
		resp, ok := respPayload.(*TestTaskData)
		s.True(ok)
		resp.Response = expectedTaskData.Response
		resp.Name = expectedTaskData.Name
	}).Return(nil).Once()

	// when
	err := s.sut.Fetch(s.taskName, &payload)

	// then
	s.NoError(err)
	s.Equal(expectedTaskData, payload)
}
