package v2

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/goharbor/go-client/pkg/sdk/v2.0/client/member"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/client/project"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/client/quota"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/models"
	goharborv1beta1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	utilstring "github.com/goharbor/harbor-operator/pkg/utils/strings"
	"github.com/pkg/errors"
	"github.com/spotahome/redis-operator/log"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	paginationSize int64 = 25
)

// EnsureProject ensures the specified project is on the harbor server
// If project with name is existing, then error will be nil.
func (c *Client) EnsureProject(name string) (int64, error) {
	if len(name) == 0 {
		return -1, errors.New("project name is empty")
	}

	if c.harborClient == nil {
		return -1, errors.New("nil harbor client")
	}

	// Check existence first
	p, err := c.GetProjectByName(name)
	if err == nil {
		return int64(p.ProjectID), nil
	}

	if err != nil {
		if !strings.Contains(err.Error(), "no project with name") {
			return 0, errors.Errorf("error when getting project %s: %s", name, err)
		}
	}

	fmt.Println("creating project since target project doesn't exist")

	// Create one when the project does not exist
	cparams := project.NewCreateProjectParamsWithContext(c.context).
		WithTimeout(c.timeout).
		WithProject(&models.ProjectReq{
			ProjectName: name,
			Metadata: &models.ProjectMetadata{
				Public: "false",
			},
		})

	cp, err := c.harborClient.Client.Project.CreateProject(c.context, cparams)
	if err != nil {
		return -1, fmt.Errorf("ensure project error: %w", err)
	}

	return utilstring.ExtractID(cp.Location)
}

func (c *Client) ProjectExists(name string) (bool, error) {
	headProjectOK, err := c.harborClient.Client.Project.HeadProject(c.context, project.NewHeadProjectParams().WithProjectName(name))
	// headProjectNotFound error is expected when project does not exist, throw all other errors
	if err != nil && strings.Contains(err.Error(), "headProjectNotFound") {
		err = nil
	}

	return headProjectOK != nil, err
}

// GetProjectByName gets the project data.
func (c *Client) GetProjectByName(name string) (*models.Project, error) {
	if len(name) == 0 {
		return nil, errors.New("project name is empty")
	}

	if c.harborClient == nil {
		return nil, errors.New("nil harbor client")
	}
	// Use listProject endpoint since getProject requires project id query key
	params := project.NewListProjectsParamsWithContext(c.context).
		WithTimeout(c.timeout).
		WithName(&name)

	res, err := c.harborClient.Client.Project.ListProjects(c.context, params)
	if err != nil {
		return nil, fmt.Errorf("get project error: %w", err)
	}

	if len(res.Payload) < 1 {
		return nil, errors.Errorf("no project with name %s exists", name)
	}

	return res.Payload[0], nil
}

func (c *Client) GetProjectByID(id int32) (*models.Project, error) {
	if id < 1 {
		return nil, errors.New("project id is < 1")
	}

	if c.harborClient == nil {
		return nil, errors.New("nil harbor client")
	}

	params := project.NewGetProjectParamsWithContext(c.context).WithProjectNameOrID(strconv.Itoa(int(id)))

	res, err := c.harborClient.Client.Project.GetProject(c.context, params)
	if err != nil {
		return nil, fmt.Errorf("get project by ID error: %w", err)
	}

	return res.Payload, nil
}

func (c *Client) CreateProject(hp *goharborv1beta1.HarborProject) (int32, error) {
	if c.harborClient == nil {
		return -1, errors.New("nil harbor client")
	}

	projectRequest, err := c.GetProjectRequest(hp)
	if err != nil {
		return -1, fmt.Errorf("create project error: %w", err)
	}

	params := project.NewCreateProjectParams().WithProject(projectRequest)

	res, err := c.harborClient.Client.Project.CreateProject(c.context, params)
	if err != nil {
		return -1, fmt.Errorf("create project error: %w", err)
	}

	rid, err := utilstring.ExtractID(res.Location)
	if err != nil {
		// ignore this error that should never happen
		c.log.Error(err, "location", res.Location)
	}

	if rid > 0 && rid <= math.MaxInt32 {
		return int32(rid), nil
	}

	return -1, errors.New("out of bounds project ID")
}

func (c *Client) UpdateProject(projectName string, hp *goharborv1beta1.HarborProject) error {
	if c.harborClient == nil {
		return errors.New("nil harbor client")
	}

	projectRequest, err := c.GetProjectRequest(hp)
	if err != nil {
		return fmt.Errorf("update project error: %w", err)
	}

	params := project.NewUpdateProjectParams().
		WithTimeout(c.timeout).
		WithProjectNameOrID(projectName).
		WithProject(projectRequest)

	_, err = c.harborClient.Client.Project.UpdateProject(c.context, params)
	if err != nil {
		return fmt.Errorf("update project error: %w", err)
	}

	return nil
}

// DeleteProject deletes project.
func (c *Client) DeleteProject(name string) error {
	if len(name) == 0 {
		return errors.New("project name is empty")
	}

	if c.harborClient == nil {
		return errors.New("nil harbor client")
	}

	exists, err := c.ProjectExists(name)
	if err != nil {
		return fmt.Errorf("delete project error: %w", err)
	}

	if !exists {
		return nil
	}

	// Get ID first
	p, err := c.GetProjectByName(name)
	if err != nil {
		return fmt.Errorf("error while deleting project \"%s\" (%d): %w", name, p.ProjectID, err)
	}

	params := project.NewDeleteProjectParamsWithContext(c.context).
		WithTimeout(c.timeout).
		WithProjectNameOrID(strconv.FormatInt(int64(p.ProjectID), 10))

	if _, err = c.harborClient.Client.Project.DeleteProject(c.context, params); err != nil {
		return fmt.Errorf("error while deleting project \"%s\" (%d): %w", name, p.ProjectID, err)
	}

	return nil
}

func (c *Client) GetQuotaByProjectID(projectID int32) (*models.Quota, error) {
	id := strconv.Itoa(int(projectID))

	quotas, err := c.harborClient.Client.Quota.ListQuotas(c.context, quota.NewListQuotasParams().WithReferenceID(&id))
	if err != nil {
		return nil, err
	}
	// We only expect one quota per project.
	if quotas.XTotalCount != 1 {
		return nil, errors.Errorf("unexpected quota payload length %d", quotas.XTotalCount)
	}

	return quotas.GetPayload()[0], nil
}

func (c *Client) GetQuotaByID(quotaID int64) (*models.Quota, error) {
	_quota, err := c.harborClient.Client.Quota.GetQuota(c.context, quota.NewGetQuotaParams().WithID(quotaID))
	if err != nil {
		return nil, err
	}

	return _quota.GetPayload(), nil
}

func (c *Client) UpdateProjectQuota(quotaID int64, storageLimit int64) error {
	params := quota.NewUpdateQuotaParams().
		WithID(quotaID).
		WithHard(&models.QuotaUpdateReq{
			Hard: models.ResourceList{
				"storage": storageLimit,
			},
		})

	_, err := c.harborClient.Client.Quota.UpdateQuota(c.context, params)
	if err != nil {
		return fmt.Errorf("update project quota error: %w", err)
	}

	return nil
}

func (c *Client) GetProjectMembers(hp *goharborv1beta1.HarborProject) ([]*models.ProjectMemberEntity, error) {
	var currentMemberships []*models.ProjectMemberEntity
	// handle pagination for listing current project members
	pageSize := paginationSize
	page := int64(1)
	params := member.NewListProjectMembersParams().
		WithProjectNameOrID(hp.Spec.ProjectName).
		WithPageSize(&pageSize).
		WithPage(&page)

	for {
		listResponse, err := c.harborClient.Client.Member.ListProjectMembers(c.context, params)
		if err != nil {
			return nil, err
		}

		if page == 1 {
			currentMemberships = listResponse.GetPayload()
		} else {
			currentMemberships = append(currentMemberships, listResponse.GetPayload()...)
		}

		currentMembershipsLen := len(currentMemberships)

		if currentMembershipsLen < int(listResponse.XTotalCount) {
			log.Info("handle membership pagination", "currentCount", currentMembershipsLen, "totalCount", listResponse.XTotalCount)
			page++
		} else {
			break
		}
	}

	return currentMemberships, nil
}

func (c *Client) CreateProjectMember(projectName string, newMember *models.ProjectMember) error {
	params := member.NewCreateProjectMemberParams().
		WithProjectMember(newMember).
		WithProjectNameOrID(projectName)

	_, err := c.harborClient.Client.Member.CreateProjectMember(c.context, params)
	if err != nil {
		return fmt.Errorf("create project member error: %w", err)
	}

	return nil
}

func (c *Client) UpdateProjectMember(projectName string, memberID int64, role *models.RoleRequest) error {
	params := member.NewUpdateProjectMemberParams().
		WithProjectNameOrID(projectName).
		WithMid(memberID).
		WithRole(role)

	_, err := c.harborClient.Client.Member.UpdateProjectMember(c.context, params)
	if err != nil {
		return fmt.Errorf("update project member error: %w", err)
	}

	return nil
}

func (c *Client) DeleteProjectMember(projectName string, memberID int64) error {
	params := member.NewDeleteProjectMemberParams().
		WithProjectNameOrID(projectName).
		WithMid(memberID)

	_, err := c.harborClient.Client.Member.DeleteProjectMember(c.context, params)
	if err != nil {
		return fmt.Errorf("delete project member error: %w", err)
	}

	return nil
}

func (c *Client) GetProjectRequest(hp *goharborv1beta1.HarborProject) (*models.ProjectReq, error) {
	if hp.Spec.HarborProjectMetadata == nil {
		hp.Spec.HarborProjectMetadata = &goharborv1beta1.HarborProjectMetadata{}
	}

	model := &models.ProjectReq{
		ProjectName:  hp.Spec.ProjectName,
		CVEAllowlist: &models.CVEAllowlist{},
		Metadata: &models.ProjectMetadata{
			AutoScan:                 utilstring.Bool2Str(hp.Spec.HarborProjectMetadata.AutoScan),
			EnableContentTrust:       utilstring.Bool2Str(hp.Spec.HarborProjectMetadata.EnableContentTrust),
			EnableContentTrustCosign: utilstring.Bool2Str(hp.Spec.HarborProjectMetadata.EnableContentTrustCosign),
			PreventVul:               utilstring.Bool2Str(hp.Spec.HarborProjectMetadata.PreventVulnerable),
			Public:                   *utilstring.Bool2Str(hp.Spec.HarborProjectMetadata.Public),
		},
	}

	// create objects for Harbor API from CVE List in Custom Resource
	for _, cve := range hp.Spec.CveAllowList {
		model.CVEAllowlist.Items = append(model.CVEAllowlist.Items, &models.CVEAllowlistItem{CVEID: cve})
	}

	// if ReuseSysCveAllowlist is not explicitly set, set it depending on if project cve allow list is configured
	if hp.Spec.HarborProjectMetadata.ReuseSysCveAllowlist == nil {
		reuse := len(hp.Spec.CveAllowList) == 0
		model.Metadata.ReuseSysCVEAllowlist = utilstring.Bool2Str(&reuse)
	} else {
		model.Metadata.ReuseSysCVEAllowlist = utilstring.Bool2Str(hp.Spec.HarborProjectMetadata.ReuseSysCveAllowlist)
	}

	model.Metadata.Severity = &hp.Spec.HarborProjectMetadata.Severity

	// if set, parse human readable storage quota (e.g. "10Gi") into byte int64 for Harbor API
	if hp.Spec.StorageQuota != "" {
		parsedQuota, err := resource.ParseQuantity(hp.Spec.StorageQuota)
		if err != nil {
			return nil, err
		}

		value := parsedQuota.Value()
		model.StorageLimit = &value
	}

	return model, nil
}
