package project

import (
	"strings"

	"github.com/go-logr/logr"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/models"
	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	"github.com/pkg/errors"
)

const errorStatus string = "GetProjectQuotaError"

func (r *Reconciler) reconcileQuota(hp *goharborv1.HarborProject, log logr.Logger) error {
	projectRequest, err := r.Harbor.GetProjectRequest(hp)
	if err != nil {
		return errors.Wrapf(err, "error getting harbor project request")
	}

	var projectQuota *models.Quota

	if hp.Status.QuotaID == 0 {
		// QuotaID in custom resource still undefined. Get Quota via ProjectID
		quota, err := r.Harbor.GetQuotaByProjectID(hp.Status.ProjectID)
		if err != nil {
			err = errors.Wrapf(err, "error getting quota of harbor project")
			hp.Status.Reason = errorStatus

			return err
		}
		// set QuotaID field in custom resource and save quota for further usage
		hp.Status.QuotaID = quota.ID
		projectQuota = quota
	} else {
		quota, err := r.Harbor.GetQuotaByID(hp.Status.QuotaID)
		if err != nil {
			// reset cached quota ID if its not found
			if strings.Contains(err.Error(), "getQuotaNotFound") {
				hp.Status.QuotaID = 0
			}

			err = errors.Wrapf(err, "error getting quota of harbor project")
			hp.Status.Reason = errorStatus

			return err
		}
		projectQuota = quota
	}

	// update quota if it was changed
	if *projectRequest.StorageLimit != projectQuota.Hard["storage"] {
		log.Info("quota changed", "oldQuota", projectQuota.Hard["storage"], "newQuota", *projectRequest.StorageLimit)

		err := r.Harbor.UpdateProjectQuota(projectQuota.ID, *projectRequest.StorageLimit)
		if err != nil {
			err = errors.Wrapf(err, "error updating quota of harbor project")
			hp.Status.Reason = "UpdateProjectQuotaError"

			return err
		}
	}

	return nil
}
