package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/convox/rack/api/httperr"
	"github.com/convox/rack/api/models"
)

func init() {
}

func SystemReleaseList(rw http.ResponseWriter, r *http.Request) *httperr.Error {
	rack, err := models.GetSystem()

	if awsError(err) == "ValidationError" {
		return httperr.Errorf(404, "no such stack: %s", rack)
	}

	if err != nil {
		return httperr.Server(err)
	}

	releases, err := models.ListReleases(rack.Name)

	if err != nil {
		return httperr.Server(err)
	}

	return RenderJson(rw, releases)
}

func SystemShow(rw http.ResponseWriter, r *http.Request) *httperr.Error {
	rack, err := models.GetSystem()

	if awsError(err) == "ValidationError" {
		return httperr.Errorf(404, "no such stack: %s", rack)
	}

	if err != nil {
		return httperr.Server(err)
	}

	return RenderJson(rw, rack)
}

func SystemUpdate(rw http.ResponseWriter, r *http.Request) *httperr.Error {
	rack, err := models.GetSystem()

	if err != nil {
		return httperr.Server(err)
	}

	notifyData := map[string]string{}

	if count := GetForm(r, "count"); count != "" {
		count, err := strconv.Atoi(count)

		if err != nil {
			return httperr.Server(err)
		}

		rack.Count = count

		notifyData["count"] = strconv.Itoa(count)
	}

	if t := GetForm(r, "type"); t != "" {
		rack.Type = t

		notifyData["type"] = t
	}

	if version := GetForm(r, "version"); version != "" {
		rack.Version = version

		notifyData["version"] = version
	}

	err = rack.Save()

	if awsError(err) == "ValidationError" {
		switch {
		case strings.Index(err.Error(), "No updates are to be performed") > -1:
			return httperr.Errorf(403, "no system updates are to be performed")
		case strings.Index(err.Error(), "can not be updated") > -1:
			return httperr.Errorf(403, "system is already updating")
		}
	}

	if err != nil {
		return httperr.Server(err)
	}

	rack, err = models.GetSystem()

	if err != nil {
		return httperr.Server(err)
	}

	models.NotifySuccess("rack:update", notifyData)

	return RenderJson(rw, rack)
}
