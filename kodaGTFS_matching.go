package main

import (
	"time"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"

	"encoding/csv"
	"math"
	"os"
	"strconv"
	//"io/ioutil"
	//	"strings"
	// "github.com/kniren/gota/dataframe"
)

const (
	GPSLAT            = 34.8636752
	GPSLON            = 137.1621358
	EQUATORIAL_RADIUS = 6378137.0            // 赤道半径 GRS80
	POLAR_RADIUS      = 6356752.314          // 極半径 GRS80
	ECCENTRICITY      = 0.081819191042815790 // 第一離心率 GRS80
	TIMEFORMAT        = "15:04:05"
)

type Coord struct {
	Latitude  float64
	Longitude float64
}

type shapes struct {
	shape_id       string
	shape_lat      float64
	shape_lon      float64
	shape_sequence int
}

type trips struct {
	trip_id  string
	shape_id string
}

type stops struct {
	stop_id  string
	stop_lat float64
	stop_lon float64
}

type stoptimes struct {
	trip_id        string
	arrival_time   string
	departure_time string
	stop_id        string
}

func LoadShapeCsv(filename string) []shapes {
	var line []string
	var shape shapes
	var lat float64
	var lon float64
	var sequence int
	ary := make([]shapes, 0)

	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	for {
		line, err = reader.Read()
		if err != nil {
			break
		}

		lat, err = strconv.ParseFloat(line[1], 64)
		if err != nil {
			continue
		}
		lon, err = strconv.ParseFloat(line[2], 64)
		if err != nil {
			continue
		}
		sequence, err = strconv.Atoi(line[3])
		if err != nil {
			continue
		}
		shape.shape_id = line[0]
		shape.shape_lat = lat
		shape.shape_lon = lon
		shape.shape_sequence = sequence

		ary = append(ary, shape)
	}
	return ary
}

func LoadTripCsv(filename string) []trips {
	var trip trips
	var line []string
	p := 0
	ary := make([]trips, 0)

	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	reader := csv.NewReader(transform.NewReader(file, japanese.ShiftJIS.NewDecoder()))
	for {
		line, err = reader.Read()
		//fmt.Print(line)
		if err != nil {
			break
		}
		if p == 0 {
			p = 1
			continue
		}
		trip.trip_id = line[2]
		trip.shape_id = line[5]
		ary = append(ary, trip)

	}
	return ary
}

func LoadStopsCsv(filename string) []stops {
	var stop stops
	var line []string
	var lat float64
	var lon float64
	ary := make([]stops, 0)

	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	reader := csv.NewReader(transform.NewReader(file, japanese.ShiftJIS.NewDecoder()))
	for {
		line, err = reader.Read()
		//fmt.Print(line)
		if err != nil {
			break
		}
		lat, err = strconv.ParseFloat(line[4], 64)
		if err != nil {
			continue
		}
		lon, err = strconv.ParseFloat(line[5], 64)
		if err != nil {
			continue
		}

		stop.stop_id = line[0]
		stop.stop_lat = lat
		stop.stop_lon = lon
		ary = append(ary, stop)

	}
	return ary
}

func LoadStopTimesCsv(filename string) []stoptimes {
	var line []string
	var stoptime stoptimes
	ary := make([]stoptimes, 0)
	p := 0
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	reader := csv.NewReader(transform.NewReader(file, japanese.ShiftJIS.NewDecoder()))
	for {
		line, err = reader.Read()
		//fmt.Printf("%v\n", line)
		if err != nil {
			break
		}

		if p == 0 {
			p = 1
			continue
		} //pass first loop

		stoptime.trip_id = line[0]
		stoptime.arrival_time = line[1]
		stoptime.departure_time = line[2]
		stoptime.stop_id = line[3]
		ary = append(ary, stoptime)
	}
	return ary
}

func degree2radian(x float64) float64 {
	return x * math.Pi / 180
}

func Power2(x float64) float64 {
	return math.Pow(x, 2)
}

func hubenyDistance(src Coord, dst Coord) float64 {
	dx := degree2radian(dst.Longitude - src.Longitude)
	dy := degree2radian(dst.Latitude - src.Latitude)
	my := degree2radian((src.Latitude + dst.Latitude) / 2)

	W := math.Sqrt(1 - (Power2(ECCENTRICITY) * Power2(math.Sin(my)))) // 卯酉線曲率半径の分母
	m_numer := EQUATORIAL_RADIUS * (1 - Power2(ECCENTRICITY))         // 子午線曲率半径の分子

	M := m_numer / math.Pow(W, 3) // 子午線曲率半径
	N := EQUATORIAL_RADIUS / W    // 卯酉線曲率半径

	d := math.Sqrt(Power2(dy*M) + Power2(dx*N*math.Cos(my)))

	return d
}

func GetIdHeader(ary []shapes, gpsCoord Coord) string {
	type id struct {
		id_header int
		number    int
	}
	var dists []float64
	var dist float64

	for _, v := range ary {
		shapeCoord := Coord{v.shape_lat, v.shape_lon}
		dist = hubenyDistance(gpsCoord, shapeCoord)
		dists = append(dists, dist)
	} //shapeの各点への距離を計算

	min := dists[0]
	min_index := 0
	for i, v := range dists {
		if min > v {
			min = v
			min_index = i
		}
	} //最も近い点のindexを取得
	target_id := ary[min_index].shape_id

	return target_id[0:2]
}

func GetTripIdList(ary []trips, id_header string) []trips {
	target_trips := make([]trips, 0)
	for _, v := range ary {
		id := v.shape_id
		if id[0:2] == id_header {
			target_trips = append(target_trips, v)
		}
	}
	return target_trips
}

func GetTripId(tripIdList []trips, aryStops []stops, aryStopTimes []stoptimes, gpsCoord Coord, gpsTime time.Time) string {
	var dists []float64
	var dist float64
	aryTargetStopTimes := make([]stoptimes, 0)
	zerotime := 0 * time.Second
	min_dist_time := 12 * time.Hour

	for _, v := range aryStops { //GPS座標に一番近いバス停を検索
		stopsCoord := Coord{v.stop_lat, v.stop_lon}
		dist = hubenyDistance(gpsCoord, stopsCoord)
		dists = append(dists, dist)
	}
	min := dists[0]
	min_index := 0
	for i, v := range dists {
		if min > v {
			min = v
			min_index = i
		}
	}
	target_stop_id := aryStops[min_index].stop_id

	for _, v := range tripIdList { //対象ルートのtripIDから上記のバス停の出発時刻と現在時刻と比較して最も時間が近いtripIDを取得
		for _, w := range aryStopTimes {
			if v.trip_id == w.trip_id {
				aryTargetStopTimes = append(aryTargetStopTimes, w)
			}
		}
	}
	for i, v := range aryTargetStopTimes {
		if v.stop_id == target_stop_id {
			t, _ := time.Parse(TIMEFORMAT, v.departure_time)

			dist_time := t.Sub(gpsTime)
			if dist_time < zerotime {
				dist_time = -dist_time
			}
			if dist_time < min_dist_time {
				min_dist_time = dist_time
				min_index = i
			}
		}
	}
	target_trip_id := aryTargetStopTimes[min_index].trip_id

	return target_trip_id
}

func GetBusNumber(gpsCoord Coord, gpsTime time.Time) string {

	aryShapes := LoadShapeCsv("gtfs_csv/shapes.csv")
	aryTrips := LoadTripCsv("gtfs_csv/trip.csv")
	aryStops := LoadStopsCsv("gtfs_csv/stops.csv")
	aryStopTimes := LoadStopTimesCsv(("gtfs_csv/stop_time.csv"))

	idHeader := GetIdHeader(aryShapes, gpsCoord)                                //点マッチングでルートの特定(shape_idの頭2文字)
	tripIdList := GetTripIdList(aryTrips, idHeader)                             //対象ルートのtrip_idをスライスで取得
	trip_id := GetTripId(tripIdList, aryStops, aryStopTimes, gpsCoord, gpsTime) //どのバス(trip_id)かを特定
	return trip_id
}
