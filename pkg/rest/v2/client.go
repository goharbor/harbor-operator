package v2

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/client/health"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/client/project"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/client/robotv1"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/models"
	"github.com/goharbor/harbor-operator/pkg/rest/model"
	utilstring "github.com/goharbor/harbor-operator/pkg/utils/strings"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
)

// Client for talking to Harbor V2 API
// Wrap based on sdk v2.
type Client struct {
	// Server info for talking to
	server *model.HarborServer
	// Timeout for client connection
	timeout time.Duration
	// Context for doing client connection
	context context.Context
	// Harbor API client
	harborClient *model.HarborClientV2
	// Logger
	log logr.Logger
}

// New V2 client.
func New() *Client {
	// Initialize with default settings
	return &Client{
		timeout: 30 * time.Second, //nolint:gomnd
		context: context.Background(),
		log:     ctrl.Log.WithName("v2").WithName("client"),
	}
}

// NewWithServer new V2 client with provided server.
func NewWithServer(s *model.HarborServer) (*Client, error) {
	var err error
	// Initialize with default settings
	c := New()
	c.server = s
	c.harborClient, err = s.ClientV2()

	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) WithContext(ctx context.Context) *Client {
	if ctx != nil {
		c.context = ctx
	}

	return c
}

func (c *Client) WithTimeout(timeout time.Duration) *Client {
	c.timeout = timeout

	return c
}

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
	p, err := c.GetProject(name)
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

// GetProject gets the project data.
func (c *Client) GetProject(name string) (*models.Project, error) {
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

// DeleteProject deletes project.
func (c *Client) DeleteProject(name string) error {
	if len(name) == 0 {
		return errors.New("project name is empty")
	}

	if c.harborClient == nil {
		return errors.New("nil harbor client")
	}

	// Get ID first
	p, err := c.GetProject(name)
	if err != nil {
		return fmt.Errorf("delete project error: %w", err)
	}

	params := project.NewDeleteProjectParamsWithContext(c.context).
		WithTimeout(c.timeout).
		WithProjectNameOrID(string(p.ProjectID))

	if _, err = c.harborClient.Client.Project.DeleteProject(c.context, params); err != nil {
		return err
	}

	return nil
}

func (c *Client) CheckHealth() (*models.OverallHealthStatus, error) {
	params := health.NewGetHealthParams().
		WithTimeout(c.timeout)

	res, err := c.harborClient.Client.Health.GetHealth(c.context, params)
	if err != nil {
		return nil, err
	}

	return res.Payload, nil
}

func (c *Client) CreateRobotAccount(projectID string) (*models.Robot, error) {
	if len(projectID) == 0 {
		return nil, errors.New("empty project id")
	}

	if c.harborClient == nil {
		return nil, errors.New("nil harbor client")
	}

	params := robotv1.NewCreateRobotV1Params().
		WithTimeout(c.timeout).
		WithProjectNameOrID(projectID).
		WithRobot(&models.RobotCreateV1{
			Access: []*models.Access{
				{
					Action:   "push",
					Resource: fmt.Sprintf("/project/%s/repository", projectID),
				},
			},
			Description: "automated by harbor automation operator",
			ExpiresAt:   -1, // never
			Name:        utilstring.RandomName("4k8s"),
		})

	res, err := c.harborClient.Client.Robotv1.CreateRobotV1(c.context, params)
	if err != nil {
		return nil, err
	}

	rid, err := utilstring.ExtractID(res.Location)
	if err != nil {
		// ignore this error that should never happen
		c.log.Error(err, "location", res.Location)
	}

	return &models.Robot{
		ID:     rid,
		Name:   res.Payload.Name,
		Secret: res.Payload.Secret,
	}, nil
}

func (c *Client) DeleteRobotAccount(projectID, robotID int64) error {
	if projectID <= 0 {
		return errors.New("invalid project id")
	}

	if robotID <= 0 {
		return errors.New("invalid robot id")
	}

	if c.harborClient == nil {
		return errors.New("nil harbor client")
	}

	params := robotv1.NewDeleteRobotV1Params().
		WithTimeout(c.timeout).
		WithProjectNameOrID(fmt.Sprintf("%d", projectID)).
		WithRobotID(robotID)

	if _, err := c.harborClient.Client.Robotv1.DeleteRobotV1(c.context, params); err != nil {
		return err
	}

	return nil
}

func (c *Client) GetRobotAccount(projectID, robotID int64) (*models.Robot, error) {
	if projectID <= 0 {
		return nil, errors.New("invalid project id")
	}

	if robotID <= 0 {
		return nil, errors.New("invalid robot id")
	}

	if c.harborClient == nil {
		return nil, errors.New("nil harbor client")
	}

	params := robotv1.NewGetRobotByIDV1Params().
		WithTimeout(c.timeout).
		WithProjectNameOrID(fmt.Sprintf("%d", projectID)).
		WithRobotID(robotID)

	res, err := c.harborClient.Client.Robotv1.GetRobotByIDV1(c.context, params)
	if err != nil {
		return nil, err
	}

	return &models.Robot{
		ID:   robotID,
		Name: res.Payload.Name,
	}, nil
}
