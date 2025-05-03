package ingestwebapp

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Kaese72/device-store/ingestmodels"
	"github.com/Kaese72/device-store/internal/logging"
	"github.com/Kaese72/device-store/internal/persistence"
	"github.com/Kaese72/huemie-lib/liberrors"
	"github.com/pkg/errors"
)

type webApp struct {
	persistence persistence.IngestPersistenceDB
}

func NewWebApp(persistence persistence.IngestPersistenceDB) webApp {
	return webApp{
		persistence: persistence,
	}
}

func serveHTTPError(err error, ctx context.Context, writer http.ResponseWriter) {
	var apiError *liberrors.ApiError
	if !errors.As(err, &apiError) {
		apiError = liberrors.NewApiError(liberrors.InternalError, errors.New("unknown internal error occured"))
	}

	if apiError.Reason <= 599 && apiError.Reason >= 500 {
		// Internal server error, log error
		logging.ErrorErr(err, ctx)
	} else {
		// For everything else, log an info. This may be excessive, but it only happens on errors
		logging.ErrorErr(err, ctx)
	}
	// Safe to ignore error here... kind of but not really
	resp, err := json.Marshal(struct {
		Message string `json:"message"`
	}{Message: err.Error()})
	if err != nil {
		logging.ErrorErr(errors.Wrap(err, "failed to marshal error message"), ctx)
		return
	}

	writer.WriteHeader(int(apiError.Reason))
	_, err = writer.Write(resp)
	if err != nil {
		logging.ErrorErr(errors.Wrap(err, "failed to write error message"), ctx)
	}
}

func (app webApp) PostDevice(writer http.ResponseWriter, reader *http.Request) {
	ctx := reader.Context()
	bridgeKey := reader.Header.Get("Bridge-Key")
	if bridgeKey == "" {
		http.Error(writer, "May not update devices when not identifying as a bridge", http.StatusBadRequest)
		return
	}
	device := ingestmodels.Device{}
	err := json.NewDecoder(reader.Body).Decode(&device)
	if err != nil {
		serveHTTPError(liberrors.NewApiError(liberrors.UserError, err), ctx, writer)
		return
	}
	// We do not trust the client that much. Override the bridgeKey
	device.BridgeKey = bridgeKey

	// FIXME local tests do not allow this. Replace with JWT authentication or something
	// _, err = attendant.GetAdapter(bridgeKey, ctx)
	// if err != nil {
	// 	serveHTTPError(liberrors.NewApiError(liberrors.InternalError, err), ctx, writer)
	// 	return
	// }
	_, err = app.persistence.PostDevice(ctx, device)
	if err != nil {
		serveHTTPError(liberrors.NewApiError(liberrors.InternalError, err), ctx, writer)
		return
	}
	writer.WriteHeader(http.StatusOK)
	// Update the rabbitmq queue with the changed attributes after the device has been updated
	// If this fails the device remains changed

}

func (app webApp) PostGroup(writer http.ResponseWriter, reader *http.Request) {
	ctx := reader.Context()
	bridgeKey := reader.Header.Get("Bridge-Key")
	if bridgeKey == "" {
		http.Error(writer, "May not update groups when not identifying as a bridge", http.StatusBadRequest)
		return
	}
	group := ingestmodels.Group{}
	err := json.NewDecoder(reader.Body).Decode(&group)
	if err != nil {
		serveHTTPError(liberrors.NewApiError(liberrors.UserError, err), ctx, writer)
		return
	}
	// We do not trust the client that much. Override the bridgeKey
	group.BridgeKey = bridgeKey

	err = app.persistence.PostGroup(ctx, group)
	if err != nil {
		serveHTTPError(liberrors.NewApiError(liberrors.InternalError, err), ctx, writer)
		return
	}
	writer.WriteHeader(http.StatusOK)
}
