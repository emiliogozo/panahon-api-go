package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	db "github.com/emiliogozo/panahon-api-go/db/sqlc"
	"github.com/emiliogozo/panahon-api-go/util"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

type stationObsResponse struct {
	ID        int64              `json:"id"`
	Pres      util.NullFloat4    `json:"pres"`
	Rr        util.NullFloat4    `json:"rr"`
	Rh        util.NullFloat4    `json:"rh"`
	Temp      util.NullFloat4    `json:"temp"`
	Td        util.NullFloat4    `json:"td"`
	Wdir      util.NullFloat4    `json:"wdir"`
	Wspd      util.NullFloat4    `json:"wspd"`
	Wspdx     util.NullFloat4    `json:"wspdx"`
	Srad      util.NullFloat4    `json:"srad"`
	Mslp      util.NullFloat4    `json:"mslp"`
	Hi        util.NullFloat4    `json:"hi"`
	StationID int64              `json:"station_id"`
	Timestamp pgtype.Timestamptz `json:"timestamp"`
	Wchill    util.NullFloat4    `json:"wchill"`
	QcLevel   int32              `json:"qc_level"`
} //@name StationObservationResponse

func newStationObsResponse(obs db.ObservationsObservation) stationObsResponse {
	return stationObsResponse{
		ID:        obs.ID,
		StationID: obs.StationID,
		Pres:      obs.Pres,
		Rr:        obs.Rr,
		Rh:        obs.Rh,
		Temp:      obs.Temp,
		Td:        obs.Td,
		Wdir:      obs.Wdir,
		Wspd:      obs.Wspd,
		Wspdx:     obs.Wspdx,
		Srad:      obs.Srad,
		Mslp:      obs.Mslp,
		Hi:        obs.Hi,
		Wchill:    obs.Wchill,
		Timestamp: obs.Timestamp,
		QcLevel:   obs.QcLevel,
	}
}

type listStationObsUri struct {
	StationID int64 `uri:"station_id" binding:"required,min=1"`
}

type listStationObsReq struct {
	Page      int32  `form:"page,default=1" binding:"omitempty,min=1"`            // page number
	PerPage   int32  `form:"per_page,default=5" binding:"omitempty,min=1,max=30"` // limit
	StartDate string `form:"start_date" binding:"omitempty,date_time"`
	EndDate   string `form:"end_date" binding:"omitempty,date_time"`
} //@name ListStationObservationsParams

type listStationObsRes struct {
	Page    int32                `json:"page"`
	PerPage int32                `json:"per_page"`
	Total   int64                `json:"total"`
	Data    []stationObsResponse `json:"data"`
} //@name ListStationObservationsResponse

// ListStationObservations
//
//	@Summary	List station observations
//	@Tags		observations
//	@Accept		json
//	@Produce	json
//	@Param		station_id	path		int					true	"Station ID"
//	@Param		req			query		listStationObsReq	false	"List station observations parameters"
//	@Success	200			{object}	listStationObsRes
//	@Router		/stations/{station_id}/observations [get]
func (s *Server) ListStationObservations(ctx *gin.Context) {
	var uri listStationObsUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	var req listStationObsReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	startDate, isStartDate := util.ParseDateTime(req.StartDate)
	endDate, isEndDate := util.ParseDateTime(req.EndDate)

	offset := (req.Page - 1) * req.PerPage
	arg := db.ListStationObservationsParams{
		StationID: uri.StationID,
		Limit: util.NullInt4{
			Int4: pgtype.Int4{
				Int32: req.PerPage,
				Valid: true,
			},
		},
		Offset:      offset,
		IsStartDate: isStartDate,
		StartDate: pgtype.Timestamptz{
			Time:  startDate,
			Valid: !startDate.IsZero(),
		},
		IsEndDate: isEndDate,
		EndDate: pgtype.Timestamptz{
			Time:  endDate,
			Valid: !endDate.IsZero(),
		},
	}

	observations, err := s.store.ListStationObservations(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	numObs := len(observations)
	obsRes := make([]stationObsResponse, numObs)
	for i, observation := range observations {
		obsRes[i] = newStationObsResponse(observation)
	}

	totalObs, err := s.store.CountStationObservations(ctx, db.CountStationObservationsParams{
		StationID:   arg.StationID,
		IsStartDate: arg.IsStartDate,
		StartDate:   arg.StartDate,
		IsEndDate:   arg.IsEndDate,
		EndDate:     arg.EndDate,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := listStationObsRes{
		Page:    req.Page,
		PerPage: req.PerPage,
		Total:   totalObs,
		Data:    obsRes,
	}

	ctx.JSON(http.StatusOK, rsp)
}

type getStationObsReq struct {
	StationID int64 `uri:"station_id" binding:"required,min=1"`
	ID        int64 `uri:"id" binding:"required,min=1"`
}

// GetStationObservation
//
//	@Summary	Get station observation
//	@Tags		observations
//	@Accept		json
//	@Produce	json
//	@Param		station_id	path		int	true	"Station ID"
//	@Param		id			path		int	true	"Station Observation ID"
//	@Success	200			{object}	stationObsResponse
//	@Router		/stations/{station_id}/observations/{id} [get]
func (s *Server) GetStationObservation(ctx *gin.Context) {
	var req getStationObsReq
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.GetStationObservationParams{
		StationID: req.StationID,
		ID:        req.ID,
	}

	obs, err := s.store.GetStationObservation(ctx, arg)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(errors.New("station observation not found")))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, newStationObsResponse(obs))
}

type createStationObsUri struct {
	StationID int64 `uri:"station_id" binding:"required,min=1"`
}

type createStationObsReq struct {
	Pres      util.NullFloat4    `json:"pres" binding:"omitempty,numeric"`
	Rr        util.NullFloat4    `json:"rr" binding:"omitempty,numeric"`
	Rh        util.NullFloat4    `json:"rh" binding:"omitempty,numeric"`
	Temp      util.NullFloat4    `json:"temp" binding:"omitempty,numeric"`
	Td        util.NullFloat4    `json:"td" binding:"omitempty,numeric"`
	Wdir      util.NullFloat4    `json:"wdir" binding:"omitempty,numeric"`
	Wspd      util.NullFloat4    `json:"wspd" binding:"omitempty,numeric"`
	Wspdx     util.NullFloat4    `json:"wspdx" binding:"omitempty,numeric"`
	Srad      util.NullFloat4    `json:"srad" binding:"omitempty,numeric"`
	Mslp      util.NullFloat4    `json:"mslp" binding:"omitempty,numeric"`
	Hi        util.NullFloat4    `json:"hi" binding:"omitempty,numeric"`
	Wchill    util.NullFloat4    `json:"wchill" binding:"omitempty,numeric"`
	QcLevel   int32              `json:"qc_level" binding:"omitempty,numeric"`
	Timestamp pgtype.Timestamptz `json:"timestamp" binding:"omitempty,numeric"`
} //@name CreateStationObservationParams

// CreateStationObservation
//
//	@Summary	Create station observation
//	@Tags		observations
//	@Accept		json
//	@Produce	json
//	@Param		station_id	path	int					true	"Station ID"
//	@Param		stnObs		body	createStationObsReq	true	"Create station observation parameters"
//	@Security	BearerAuth
//	@Success	201	{object}	stationObsResponse
//	@Router		/stations/{station_id}/observations [post]
func (s *Server) CreateStationObservation(ctx *gin.Context) {
	var uri createStationObsUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var req createStationObsReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.CreateStationObservationParams{
		StationID: uri.StationID,
		Pres:      req.Pres,
		Rr:        req.Rr,
		Rh:        req.Rh,
		Temp:      req.Temp,
		Td:        req.Td,
		Wdir:      req.Wdir,
		Wspd:      req.Wspd,
		Wspdx:     req.Wspdx,
		Srad:      req.Srad,
		Mslp:      req.Mslp,
		Hi:        req.Hi,
		Wchill:    req.Wchill,
		Timestamp: req.Timestamp,
		QcLevel:   req.QcLevel,
	}

	obs, err := s.store.CreateStationObservation(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusCreated, newStationObsResponse(obs))
}

type updateStationObsUri struct {
	StationID int64 `uri:"station_id" binding:"required,min=1"`
	ID        int64 `uri:"id" binding:"required,min=1"`
}

type updateStationObsReq struct {
	Pres      util.NullFloat4    `json:"pres" binding:"omitempty,numeric"`
	Rr        util.NullFloat4    `json:"rr" binding:"omitempty,numeric"`
	Rh        util.NullFloat4    `json:"rh" binding:"omitempty,numeric"`
	Temp      util.NullFloat4    `json:"temp" binding:"omitempty,numeric"`
	Td        util.NullFloat4    `json:"td" binding:"omitempty,numeric"`
	Wdir      util.NullFloat4    `json:"wdir" binding:"omitempty,numeric"`
	Wspd      util.NullFloat4    `json:"wspd" binding:"omitempty,numeric"`
	Wspdx     util.NullFloat4    `json:"wspdx" binding:"omitempty,numeric"`
	Srad      util.NullFloat4    `json:"srad" binding:"omitempty,numeric"`
	Mslp      util.NullFloat4    `json:"mslp" binding:"omitempty,numeric"`
	Hi        util.NullFloat4    `json:"hi" binding:"omitempty,numeric"`
	Wchill    util.NullFloat4    `json:"wchill" binding:"omitempty,numeric"`
	QcLevel   util.NullInt4      `json:"qc_level" binding:"omitempty,numeric"`
	Timestamp pgtype.Timestamptz `json:"timestamp" binding:"omitempty,numeric"`
} //@name UpdateStationObservationParams

// UpdateStationObservation
//
//	@Summary	Update station observation
//	@Tags		observations
//	@Produce	json
//	@Param		station_id	path	int					true	"Station ID"
//	@Param		id			path	int					true	"Station Observation ID"
//	@Param		stnObs		body	updateStationObsReq	true	"Update station observation parameters"
//	@Security	BearerAuth
//	@Success	200	{object}	stationObsResponse
//	@Router		/stations/{station_id}/observations/{id} [put]
func (s *Server) UpdateStationObservation(ctx *gin.Context) {
	var uri updateStationObsUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var req updateStationObsReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.UpdateStationObservationParams{
		ID:        uri.ID,
		StationID: uri.StationID,
		Pres:      req.Pres,
		Rr:        req.Rr,
		Rh:        req.Rh,
		Temp:      req.Temp,
		Td:        req.Td,
		Wdir:      req.Wdir,
		Wspd:      req.Wspd,
		Wspdx:     req.Wspdx,
		Srad:      req.Srad,
		Mslp:      req.Mslp,
		Hi:        req.Hi,
		Wchill:    req.Wchill,
		Timestamp: req.Timestamp,
		QcLevel:   req.QcLevel,
	}

	obs, err := s.store.UpdateStationObservation(ctx, arg)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(errors.New("station not found")))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, newStationObsResponse(obs))
}

type deleteStationObsReq struct {
	StationID int64 `uri:"station_id" binding:"required,min=1"`
	ID        int64 `uri:"id" binding:"required,min=1"`
}

// DeleteStationObservation
//
//	@Summary	Delete station observation
//	@Tags		observations
//	@Accept		json
//	@Produce	json
//	@Param		station_id	path	int	true	"Station ID"
//	@Param		id			path	int	true	"Station Observation ID"
//	@Security	BearerAuth
//	@Success	204
//	@Router		/stations/{station_id}/observations/{id} [delete]
func (s *Server) DeleteStationObservation(ctx *gin.Context) {
	var req deleteStationObsReq
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.DeleteStationObservationParams{
		ID:        req.ID,
		StationID: req.StationID,
	}

	err := s.store.DeleteStationObservation(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}

type listObservationsReq struct {
	Page       int32  `form:"page,default=1" binding:"omitempty,min=1"`            // page number
	PerPage    int32  `form:"per_page,default=5" binding:"omitempty,min=1,max=30"` // limit
	StationIDs string `form:"station_ids" binding:"omitempty"`
	StartDate  string `form:"start_date" binding:"omitempty,date_time"`
	EndDate    string `form:"end_date" binding:"omitempty,date_time"`
} //@name ListObservationsParams

// ListObservations
//
//	@Summary	list station observation
//	@Tags		observations
//	@Produce	json
//	@Param		req	query		listObservationsReq	false	"List observations parameters"
//	@Success	200	{object}	listStationObsRes
//	@Router		/observations [get]
func (s *Server) ListObservations(ctx *gin.Context) {
	var req listObservationsReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var stationIDs []int64
	if len(req.StationIDs) == 0 {
		stations, err := s.store.ListStations(ctx, db.ListStationsParams{
			Limit:  10,
			Offset: 0,
		})

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}

		for i := range stations {
			stationIDs = append(stationIDs, stations[i].ID)
		}
	} else {
		stnIDStrs := strings.Split(req.StationIDs, ",")
		for i := range stnIDStrs {
			stnID, err := strconv.ParseInt(stnIDStrs[i], 10, 64)
			if err != nil {
				continue
			}
			stationIDs = append(stationIDs, stnID)
		}
	}

	startDate, isStartDate := util.ParseDateTime(req.StartDate)
	endDate, isEndDate := util.ParseDateTime(req.EndDate)

	offset := (req.Page - 1) * req.PerPage
	arg := db.ListObservationsParams{
		StationIds: stationIDs,
		Limit: util.NullInt4{
			Int4: pgtype.Int4{
				Int32: req.PerPage,
				Valid: true,
			},
		},
		Offset:      offset,
		IsStartDate: isStartDate,
		StartDate: pgtype.Timestamptz{
			Time:  startDate,
			Valid: !startDate.IsZero(),
		},
		IsEndDate: isEndDate,
		EndDate: pgtype.Timestamptz{
			Time:  endDate,
			Valid: !endDate.IsZero(),
		},
	}

	obs, err := s.store.ListObservations(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	numObs := len(obs)
	obsRes := make([]stationObsResponse, numObs)
	for i, observation := range obs {
		obsRes[i] = newStationObsResponse(observation)
	}

	totalObs, err := s.store.CountObservations(ctx, db.CountObservationsParams{
		StationIds:  arg.StationIds,
		IsStartDate: arg.IsStartDate,
		StartDate:   arg.StartDate,
		IsEndDate:   arg.IsEndDate,
		EndDate:     arg.EndDate,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := listStationObsRes{
		Page:    req.Page,
		PerPage: req.PerPage,
		Total:   totalObs,
		Data:    obsRes,
	}

	ctx.JSON(http.StatusOK, rsp)
}

type latestObservationRes struct {
	Name      string                   `json:"name"`
	Lat       util.NullFloat4          `json:"lat"`
	Lon       util.NullFloat4          `json:"lon"`
	Elevation util.NullFloat4          `json:"elevation"`
	Address   util.NullString          `json:"address"`
	Obs       db.MvObservationsCurrent `json:"obs"`
} //@name LatestObservation

func newLatestObservationResponse(data any) latestObservationRes {
	switch d := data.(type) {
	case db.ListLatestObservationsRow:
		return latestObservationRes{
			Name:      d.Name,
			Lat:       d.Lat,
			Lon:       d.Lon,
			Elevation: d.Elevation,
			Address:   d.Address,
			Obs:       d.MvObservationsCurrent,
		}
	case db.GetLatestStationObservationRow:
		return latestObservationRes{
			Name:      d.Name,
			Lat:       d.Lat,
			Lon:       d.Lon,
			Elevation: d.Elevation,
			Address:   d.Address,
			Obs:       d.MvObservationsCurrent,
		}
	default:
		return latestObservationRes{}
	}
}

// ListLatestObservations
//
//	@Summary	list latest observation
//	@Tags		observations
//	@Produce	json
//	@Success	200	{array}	latestObservationRes
//	@Router		/observations/latest [get]
func (s *Server) ListLatestObservations(ctx *gin.Context) {
	_obsSlice, err := s.store.ListLatestObservations(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	obsSlice := make([]latestObservationRes, len(_obsSlice))

	for i := range obsSlice {
		obsSlice[i] = newLatestObservationResponse(_obsSlice[i])
	}

	ctx.JSON(http.StatusOK, obsSlice)
}

type getLatestStationObsReq struct {
	StationID int64 `uri:"station_id" binding:"required,min=1"`
}

// GetLatestStationObservation
//
//	@Summary	Get latest station observation
//	@Tags		observations
//	@Accept		json
//	@Produce	json
//	@Param		station_id	path		int	true	"Station ID"
//	@Success	200			{object}	latestObservationRes
//	@Router		/stations/{station_id}/observations/latest [get]
func (s *Server) GetLatestStationObservation(ctx *gin.Context) {
	var req getLatestStationObsReq
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	obs, err := s.store.GetLatestStationObservation(ctx, req.StationID)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(errors.New("station observation not found")))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, newLatestObservationResponse(obs))
}
