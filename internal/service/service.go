package service

import (
	"pt/third_party/http"

	"github.com/google/wire"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(
	NewGreeterService,
	NewTrackerService,
	wire.Bind(new(http.RouteAppender), new(*TrackerService)),
)
