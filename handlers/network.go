package handlers

import (
	"encoding/json"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"net/http"
)

const (
	GWei = 1e9
)

// Supply godoc
// @Summary Get total supply of all native network tokens in existence
// @Tags Misc
// @Description Returns information about total supply of all native network tokens in existence (in wei).
// @Produce  json
// @Success 200 {object} types.ApiResponse{data=types.SupplyResponse} "Success"
// @Failure 400 {object} types.ApiResponse "Failure"
// @Failure 500 {object} types.ApiResponse "Server Error"
// @Router /api/v2/totalsupply [get]
func Supply(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	totalSupply, err := services.LatestTotalSupply()
	if err != nil {
		sendServerErrorResponse(w, r.URL.String(), "error processing request, please try again later")
		logger.Errorf("error getting total supply data for API %v route: %v", r.URL, err)
		return
	}

	data := types.SupplyResponse{
		TotalSupply: totalSupply,
	}

	response := &types.ApiResponse{
		Status: "OK",
		Data:   data,
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		sendServerErrorResponse(w, r.URL.String(), "could not serialize data results")
		logger.Errorf("error serializing json data for API %v route: %v", r.URL, err)
	}
}
