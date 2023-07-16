package main

import (
	"context"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/joho/godotenv"
	"googlemaps.github.io/maps"
)

type SAParams struct {
	Temperature int     // initial temperature
	M           int     // iterations
	N           int     // iterations
	Alpha       float64 // cooling rate
}

var starting_point = "Akmerkez"
var initial_route = []string{
	"Kanyon",
	"Cevahir",
}

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
	gmaps_client, err := maps.NewClient(maps.WithAPIKey(os.Getenv("GOOGLE_API_KEY")))
	if err != nil {
		panic(err)
	}
	matrix_request := &maps.DistanceMatrixRequest{
		Origins:      initial_route,
		Destinations: initial_route,
		Mode:         maps.TravelModeDriving,
		Units:        maps.UnitsMetric,
		Language:     "en",
	}
	matrix, err := gmaps_client.DistanceMatrix(context.Background(), matrix_request)
	if err != nil {
		panic(err)
	}
	for _, row := range matrix.Rows {
		for _, element := range row.Elements {
			if element.Status != "OK" {
				panic(element.Status)
			}
		}
	}
	initial_total_distance := getTotalDistance(matrix)
	println("Initial route: ", initial_route, "\nInitial total distance: ", initial_total_distance)
	result_route, total_distance := tsp(
		matrix,
		&SAParams{
			Temperature: 20000,
			M:           2000,
			N:           50,
			Alpha:       0.95,
		},
	)
	println("Final route: ", result_route, "\nFinal total distance: ", total_distance)
}

func sum(arr []int) int {
	var total int = 0
	for _, element := range arr {
		total += element
	}
	return total
}

func getTotalDistance(matrix *maps.DistanceMatrixResponse) int {
	initial_distances := []int{}
	for i := 0; i < len(initial_route); i++ {
		initial_distances = append(initial_distances, matrix.Rows[i].Elements[i+1].Distance.Meters)
	}
	initial_distances = append(initial_distances, matrix.Rows[len(initial_route)-1].Elements[0].Distance.Meters)
	return sum(initial_distances)
}

func tsp(matrix *maps.DistanceMatrixResponse, params *SAParams) ([]string, int) {
	initial_distances := []int{}
	for i := 0; i < params.M; i++ {
		for j := 0; j < params.N; j++ {
			rand.Seed(time.Now().UnixNano())
			rand_0 := rand.Intn(len(initial_route))
			rand_1 := rand.Intn(len(initial_route))
			swap_dest_0 := initial_route[rand_0]
			swap_dest_1 := initial_route[rand_1]
			counter := 0
			temp_routes := []string{}
			for counter < len(initial_route) {
				if initial_route[counter] == swap_dest_0 {
					temp_routes = append(temp_routes, swap_dest_1)
				} else if initial_route[counter] == swap_dest_1 {
					temp_routes = append(temp_routes, swap_dest_0)
				} else {
					temp_routes = append(temp_routes, initial_route[counter])
				}
				counter++
			}
			counter = 0
			for i := 0; i < len(initial_route); i++ {
				distance := matrix.Rows[i].Elements[i+1].Distance.Meters
				initial_distances = append(initial_distances, distance)
			}
			initial_distances = append(initial_distances, matrix.Rows[len(temp_routes)-1].Elements[0].Distance.Meters)
			initial_total_distance := sum(initial_distances)
			temp_distances := []int{}
			counter = 0
			for i := 0; i < len(temp_routes); i++ {
				distance := matrix.Rows[i].Elements[i+1].Distance.Meters
				temp_distances = append(temp_distances, distance)
			}
			temp_distances = append(temp_distances, matrix.Rows[len(temp_routes)-1].Elements[0].Distance.Meters)
			temp_total_distance := sum(temp_distances)
			formula := 1 / (1 + math.Exp((float64(temp_total_distance)-float64(initial_total_distance))/float64(params.Temperature)))
			rand.Seed(time.Now().UnixNano())
			rand_2 := rand.Float64()
			if temp_total_distance < initial_total_distance || rand_2 < formula {
				initial_route = temp_routes
			}
			params.Temperature = int(float64(params.Temperature) * params.Alpha)
		}
	}
	return initial_route, sum(initial_distances)
}
