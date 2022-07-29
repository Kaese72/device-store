package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Kaese72/device-store/config"
	"github.com/Kaese72/device-store/database"
	"github.com/Kaese72/sdup-lib/devicestoretemplates"
	"github.com/Kaese72/sdup-lib/logging"
	"github.com/gorilla/mux"
)

func ServeHTTPError(err error, writer http.ResponseWriter) {
	switch err.(type) {
	case database.NotFound:
		http.Error(writer, "Not Found", http.StatusNotFound)
	case database.UserError:
		http.Error(writer, "Bad Request", http.StatusBadRequest)
	default:
		http.Error(writer, "Internal Server", http.StatusInternalServerError)
	}
}

func PersistenceAPIListenAndServe(config config.HTTPConfig, persistence database.DevicePersistenceDB) error {
	router := mux.NewRouter()

	//Everything else (not /auth/login) should have the authentication middleware
	apiv0 := router.PathPrefix("/rest/v0/").Subrouter()

	apiv0.HandleFunc("/devices", func(writer http.ResponseWriter, reader *http.Request) {
		devices, err := persistence.FilterDevices()
		if err != nil {
			ServeHTTPError(err, writer)
			return
		}

		jsonEncoded, err := json.MarshalIndent(devices, "", "   ")
		if err != nil {
			ServeHTTPError(err, writer)
			return
		}

		writer.Write(jsonEncoded)
	}).Methods("GET")

	apiv0.HandleFunc("/devices", func(writer http.ResponseWriter, reader *http.Request) {
		bridgeKey := reader.Header.Get("Bridge-Key")
		device := devicestoretemplates.Device{}
		rDevice := devicestoretemplates.Device{}
		err := json.NewDecoder(reader.Body).Decode(&device)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if len(device.Capabilities) > 0 {
			if bridgeKey == "" {
				http.Error(writer, "May not set capabilities when not identifying as a bridge", http.StatusBadRequest)
				return
			}
			rDevice, err = persistence.UpdateDeviceAttributesAndCapabilities(device, devicestoretemplates.BridgeKey(bridgeKey))
			if err != nil {
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			rDevice, err = persistence.UpdateDeviceAttributes(device, true)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		jsonEncoded, err := json.MarshalIndent(rDevice, "", "   ")
		if err != nil {
			ServeHTTPError(err, writer)
			return
		}

		writer.Write(jsonEncoded)

	}).Methods("POST")

	apiv0.HandleFunc("/devices/{deviceID}", func(writer http.ResponseWriter, reader *http.Request) {
		vars := mux.Vars(reader)
		deviceID := vars["deviceID"]
		logging.Info(fmt.Sprintf("Getting device with identifier '%s'", deviceID))
		device, err := persistence.GetDeviceByIdentifier(deviceID, true)
		if err != nil {
			ServeHTTPError(err, writer)
			return
		}

		jsonEncoded, err := json.MarshalIndent(device, "", "   ")
		if err != nil {
			ServeHTTPError(err, writer)
			return
		}

		writer.Write(jsonEncoded)

	}).Methods("GET")

	apiv0.HandleFunc("/bridges", func(writer http.ResponseWriter, reader *http.Request) {
		apiBridge := devicestoretemplates.Bridge{}
		err := json.NewDecoder(reader.Body).Decode(&apiBridge)
		if err != nil {
			logging.Info("Failed to decode request body", map[string]string{"error": err.Error()})
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		for i := 0; ; i++ {
			logging.Info(fmt.Sprintf("Attempt %d at enrolling bridge", i))
			err = apiBridge.HealthCheck()
			if err != nil {
				if i > 4 {
					logging.Info(fmt.Sprintf("Attempt %d failed, threshold surpassed", i))
					http.Error(writer, err.Error(), http.StatusBadRequest)
					return
				}
				logging.Info(fmt.Sprintf("Attempt %d failed, retry in 5 seconds", i))
				time.Sleep(5000000000)
			} else {
				logging.Info("HealthCheck on enrolled bridge passed")
				break
			}
		}

		apiBridge, err = persistence.EnrollBridge(apiBridge)
		if err != nil {
			logging.Error("Failed enroll bridge", map[string]string{"error": err.Error()})
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonEncoded, err := json.MarshalIndent(apiBridge, "", "   ")
		if err != nil {
			logging.Info("Failed to marsha response", map[string]string{"error": err.Error()})
			ServeHTTPError(err, writer)
			return
		}

		writer.Write(jsonEncoded)

	}).Methods("POST")

	apiv0.HandleFunc("/bridges", func(writer http.ResponseWriter, reader *http.Request) {
		APIBridges, err := persistence.ListBridges()
		if err != nil {
			logging.Info("Failed to list bridges", map[string]string{"error": err.Error()})
			ServeHTTPError(err, writer)
			return
		}
		jsonEncoded, err := json.MarshalIndent(APIBridges, "", "   ")
		if err != nil {
			logging.Info("Failed to marshal response", map[string]string{"error": err.Error()})
			ServeHTTPError(err, writer)
			return
		}

		writer.Write(jsonEncoded)

	}).Methods("GET")

	apiv0.HandleFunc("/bridges/{bridgeId}", func(writer http.ResponseWriter, reader *http.Request) {
		vars := mux.Vars(reader)
		deviceId := vars["bridgeId"]
		err := persistence.ForgetBridge(devicestoretemplates.BridgeKey(deviceId))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

	}).Methods("DELETE")

	server := &http.Server{
		Handler: router,
		Addr:    fmt.Sprintf("%s:%d", config.Address, config.Port),
	}

	if err := server.ListenAndServe(); err != nil {
		logging.Error(err.Error())
		return err
	}

	return nil

}
