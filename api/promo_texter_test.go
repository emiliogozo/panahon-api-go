package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	mockdb "github.com/emiliogozo/panahon-api-go/db/mocks"
	db "github.com/emiliogozo/panahon-api-go/db/sqlc"
	"github.com/emiliogozo/panahon-api-go/util"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPromoTexterStoreLufft(t *testing.T) {
	mobileNum := fmt.Sprintf("63%d", util.RandomInt(9000000000, 9999999999))
	lufft := util.RandomLufft()
	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recoder *httptest.ResponseRecorder, store *mockdb.MockStore)
	}{
		{
			name: "OK",
			body: gin.H{
				"number": mobileNum,
				"msg":    lufft.String(23),
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetStationByMobileNumber(mock.AnythingOfType("*gin.Context"), mock.Anything).
					Return(db.ObservationsStation{}, nil)
				store.EXPECT().CreateStationObservation(mock.AnythingOfType("*gin.Context"), mock.Anything).
					Return(db.ObservationsObservation{}, nil)
				store.EXPECT().CreateStationHealth(mock.AnythingOfType("*gin.Context"), mock.Anything).
					Return(db.ObservationsStationhealth{}, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder, store *mockdb.MockStore) {
				store.AssertExpectations(t)
				require.Equal(t, http.StatusCreated, recorder.Code)
			},
		},
		{
			name: "NotFound",
			body: gin.H{
				"number": mobileNum,
				"msg":    lufft.String(23),
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetStationByMobileNumber(mock.AnythingOfType("*gin.Context"), mock.Anything).
					Return(db.ObservationsStation{}, db.ErrRecordNotFound)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder, store *mockdb.MockStore) {
				store.AssertExpectations(t)
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"number": mobileNum,
				"msg":    lufft.String(23),
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetStationByMobileNumber(mock.AnythingOfType("*gin.Context"), mock.Anything).
					Return(db.ObservationsStation{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder, store *mockdb.MockStore) {
				store.AssertExpectations(t)
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			store := mockdb.NewMockStore(t)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprintf("%s/ptexter", server.config.APIBasePath)
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)

			tc.checkResponse(recorder, store)
		})
	}
}
