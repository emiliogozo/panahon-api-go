package db

import (
	"context"
	"math"
	"testing"

	"github.com/emiliogozo/panahon-api-go/util"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/twpayne/go-geom"
)

type StationTestSuite struct {
	suite.Suite
}

func TestStationTestSuite(t *testing.T) {
	suite.Run(t, new(StationTestSuite))
}

func (ts *StationTestSuite) SetupTest() {
	err := util.RunDBMigration(testConfig.MigrationPath, testConfig.DBSource)
	require.NoError(ts.T(), err, "db migration problem")
}

func (ts *StationTestSuite) TearDownTest() {
	err := util.ReverseDBMigration(testConfig.MigrationPath, testConfig.DBSource)
	require.NoError(ts.T(), err, "reverse db migration problem")
}

func (ts *StationTestSuite) TestCreateStation() {
	createRandomStation(ts.T(), true)
}

func (ts *StationTestSuite) TestGetStation() {
	t := ts.T()
	station := createRandomStation(t, false)

	gotStation, err := testStore.GetStation(context.Background(), station.ID)

	require.NoError(t, err)
	require.NotEmpty(t, gotStation)

	require.Equal(t, gotStation.Name, station.Name)
	require.Equal(t, gotStation.MobileNumber, station.MobileNumber)
}

func (ts *StationTestSuite) TestListStations() {
	t := ts.T()
	n := 10
	for i := 0; i < n; i++ {
		createRandomStation(t, false)
	}

	arg := ListStationsParams{
		Limit:  5,
		Offset: 5,
	}
	gotStations, err := testStore.ListStations(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, gotStations, 5)

	for _, station := range gotStations {
		require.NotEmpty(t, station)
	}
}

func (ts *StationTestSuite) TestListStationsWithinRadius() {
	t := ts.T()
	cLat := util.RandomFloat(5.5, 18.6)
	cLon := util.RandomFloat(117.15, 126.6)
	cR := float32(1.0)
	n := 10
	for i := 0; i < n; i++ {
		theta := 2 * math.Pi * float64(util.RandomFloat(0.0, 1.0))
		var d float32
		if i%2 == 0 {
			d = cR * float32(math.Sqrt(float64(util.RandomFloat(2.0, 3.0))))
		} else {
			d = cR * float32(math.Sqrt(float64(util.RandomFloat(0.0, 1.0))))
		}
		lon := cLon + d*float32(math.Cos(theta))
		lat := cLat + d*float32(math.Sin(theta))
		p := geom.NewPoint(geom.XY).MustSetCoords(geom.Coord{float64(lon), float64(lat)}).SetSRID(4326)
		createRandomStation(t, util.Point{Point: p})
	}

	arg := ListStationsWithinRadiusParams{
		Cx:     cLon,
		Cy:     cLat,
		R:      cR,
		Limit:  int32(n),
		Offset: 0,
	}
	gotStations, err := testStore.ListStationsWithinRadius(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, gotStations, 5)

	for i := range gotStations {
		require.NotEmpty(t, gotStations[i])
	}
}

func (ts *StationTestSuite) TestListStationsWithinBBox() {
	t := ts.T()
	xMin, yMin, xMax, yMax := 120.0, 5.0, 122.0, 6.0
	n := 10
	for i := 0; i < n; i++ {
		var lat, lon float32
		if i%2 == 0 {
			lon = util.RandomFloat(float32(xMin), float32(xMax))
			lat = util.RandomFloat(float32(yMin), float32(yMax))
		} else {
			lon = util.RandomFloat(float32(xMax), float32(xMax+1.0))
			lat = util.RandomFloat(float32(yMax), float32(yMax+1.0))
		}
		p := geom.NewPoint(geom.XY).MustSetCoords(geom.Coord{float64(lon), float64(lat)}).SetSRID(4326)
		createRandomStation(t, util.Point{Point: p})
	}

	arg := ListStationsWithinBBoxParams{
		Xmin:   float32(xMin),
		Ymin:   float32(yMin),
		Xmax:   float32(xMax),
		Ymax:   float32(yMax),
		Limit:  int32(n),
		Offset: 0,
	}
	gotStations, err := testStore.ListStationsWithinBBox(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, gotStations, 5)

	for _, station := range gotStations {
		require.NotEmpty(t, station)
	}
}

func (ts *StationTestSuite) TestCountStations() {
	t := ts.T()
	n := 10
	for i := 0; i < n; i++ {
		createRandomStation(t, false)
	}

	numStations, err := testStore.CountStations(context.Background())
	require.NoError(t, err)
	require.Equal(t, numStations, int64(n))
}

func (ts *StationTestSuite) TestCountStationsWithinRadius() {
	t := ts.T()
	cLat := util.RandomFloat(5.5, 18.6)
	cLon := util.RandomFloat(117.15, 126.6)
	cR := float32(1.0)
	n := 10
	for i := 0; i < n; i++ {
		theta := 2 * math.Pi * float64(util.RandomFloat(0.0, 1.0))
		var d float32
		if i%2 == 0 {
			d = cR * float32(math.Sqrt(float64(util.RandomFloat(2.0, 3.0))))
		} else {
			d = cR * float32(math.Sqrt(float64(util.RandomFloat(0.0, 1.0))))
		}
		lon := cLon + d*float32(math.Cos(theta))
		lat := cLat + d*float32(math.Sin(theta))
		p := geom.NewPoint(geom.XY).MustSetCoords(geom.Coord{float64(lon), float64(lat)}).SetSRID(4326)
		createRandomStation(t, util.Point{Point: p})
	}

	arg := CountStationsWithinRadiusParams{
		Cx: cLon,
		Cy: cLat,
		R:  cR,
	}
	numStations, err := testStore.CountStationsWithinRadius(context.Background(), arg)
	require.NoError(t, err)
	require.Equal(t, numStations, int64(5))
}

func (ts *StationTestSuite) TestCountStationsWithinBBox() {
	t := ts.T()
	xMin, yMin, xMax, yMax := 120.0, 5.0, 122.0, 6.0
	n := 10
	for i := 0; i < n; i++ {
		var lat, lon float32
		if i%2 == 0 {
			lon = util.RandomFloat(float32(xMin), float32(xMax))
			lat = util.RandomFloat(float32(yMin), float32(yMax))
		} else {
			lon = util.RandomFloat(float32(xMax), float32(xMax+1.0))
			lat = util.RandomFloat(float32(yMax), float32(yMax+1.0))
		}
		p := geom.NewPoint(geom.XY).MustSetCoords(geom.Coord{float64(lon), float64(lat)}).SetSRID(4326)
		createRandomStation(t, util.Point{Point: p})
	}

	arg := CountStationsWithinBBoxParams{
		Xmin: float32(xMin),
		Ymin: float32(yMin),
		Xmax: float32(xMax),
		Ymax: float32(yMax),
	}
	numStations, err := testStore.CountStationsWithinBBox(context.Background(), arg)
	require.NoError(t, err)
	require.Equal(t, numStations, int64(5))
}

func (ts *StationTestSuite) TestUpdateStation() {
	var (
		oldStation      ObservationsStation
		newName         util.NullString
		newMobileNumber util.NullString
		newLat          util.NullFloat4
		newLon          util.NullFloat4
	)

	t := ts.T()

	testCases := []struct {
		name        string
		buildArg    func() UpdateStationParams
		checkResult func(updatedStation ObservationsStation, err error)
	}{
		{
			name: "NameOnly",
			buildArg: func() UpdateStationParams {
				oldStation = createRandomStation(t, true)
				newName = util.NullString{
					Text: pgtype.Text{
						String: util.RandomString(12),
						Valid:  true,
					},
				}

				return UpdateStationParams{
					ID:   oldStation.ID,
					Name: newName,
				}
			},
			checkResult: func(updatedStation ObservationsStation, err error) {
				require.NoError(t, err)
				require.NotEqual(t, oldStation.Name, updatedStation.Name)
				require.Equal(t, oldStation.Lat, updatedStation.Lat)
				require.Equal(t, oldStation.Lon, updatedStation.Lon)
				require.Equal(t, oldStation.Geom, updatedStation.Geom)
				require.Equal(t, oldStation.MobileNumber, updatedStation.MobileNumber)
			},
		},
		{
			name: "MobileNumberOnly",
			buildArg: func() UpdateStationParams {
				oldStation = createRandomStation(t, true)
				newMobileNumber = util.NullString{
					Text: pgtype.Text{
						String: util.RandomMobileNumber(),
						Valid:  true,
					},
				}
				return UpdateStationParams{
					ID:           oldStation.ID,
					MobileNumber: newMobileNumber,
				}
			},
			checkResult: func(updatedStation ObservationsStation, err error) {
				require.NoError(t, err)
				require.NotEqual(t, oldStation.MobileNumber, updatedStation.MobileNumber)
				require.Equal(t, oldStation.Name, updatedStation.Name)
				require.Equal(t, oldStation.Lat, updatedStation.Lat)
				require.Equal(t, oldStation.Lon, updatedStation.Lon)
				require.Equal(t, oldStation.Geom, updatedStation.Geom)
			},
		},
		{
			name: "LatLonOnly",
			buildArg: func() UpdateStationParams {
				oldStation = createRandomStation(t, true)
				newLat = util.RandomNullFloat4(-90.0, 90.0)
				newLon = util.RandomNullFloat4(0.0, 359.9)

				return UpdateStationParams{
					ID:  oldStation.ID,
					Lat: newLat,
					Lon: newLon,
				}
			},
			checkResult: func(updatedStation ObservationsStation, err error) {
				require.NoError(t, err)
				require.NotEqual(t, oldStation.Lat, updatedStation.Lat)
				require.NotEqual(t, oldStation.Lon, updatedStation.Lon)
				require.NotEqual(t, oldStation.Geom, updatedStation.Geom)
				require.Equal(t, oldStation.Name, updatedStation.Name)
				require.Equal(t, oldStation.MobileNumber, updatedStation.MobileNumber)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			updatedStation, err := testStore.UpdateStation(context.Background(), tc.buildArg())
			tc.checkResult(updatedStation, err)
		})
	}
}

func (ts *StationTestSuite) TestDeleteStation() {
	t := ts.T()
	station := createRandomStation(t, false)

	err := testStore.DeleteStation(context.Background(), station.ID)
	require.NoError(t, err)

	gotStation, err := testStore.GetStation(context.Background(), station.ID)
	require.Error(t, err)
	require.Empty(t, gotStation)
}

func createRandomStation(t *testing.T, geom any) ObservationsStation {
	mobileNum := util.RandomMobileNumber()

	arg := CreateStationParams{
		Name: util.RandomString(16),
		MobileNumber: util.NullString{
			Text: pgtype.Text{
				String: mobileNum,
				Valid:  true,
			},
		},
	}

	switch g := geom.(type) {
	case bool:
		if g {
			arg.Lat = util.RandomNullFloat4(5.5, 18.6)
			arg.Lon = util.RandomNullFloat4(117.15, 126.6)
		}
	case util.Point:
		arg.Lon = util.NullFloat4{
			Float4: pgtype.Float4{
				Float32: float32(g.X()),
				Valid:   true,
			},
		}
		arg.Lat = util.NullFloat4{
			Float4: pgtype.Float4{
				Float32: float32(g.Y()),
				Valid:   true,
			},
		}
	}

	station, err := testStore.CreateStation(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, station)

	require.Equal(t, arg.Name, station.Name)
	require.Equal(t, arg.MobileNumber, station.MobileNumber)
	require.True(t, station.UpdatedAt.Time.IsZero())
	require.True(t, station.CreatedAt.Valid)
	require.NotZero(t, station.CreatedAt.Time)

	return station
}
