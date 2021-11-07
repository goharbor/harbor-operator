package legacy

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/legacy/client/products"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/legacy/models"
	"github.com/goharbor/harbor-operator/pkg/rest/model"
	"github.com/goharbor/harbor-operator/pkg/utils/strings"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
)

// Client for talking to Harbor API
// Wrap based on legacy sdk.
type Client struct {
	// Server info for talking to
	server *model.HarborServer
	// Timeout for client connection
	timeout time.Duration
	// Context for doing client connection
	context context.Context
	// Harbor API client
	harborClient *model.HarborLegacyClient
	// Logger
	log logr.Logger
}

// New client.
func New() *Client {
	// Initialize with default settings.
	return &Client{
		timeout: 30 * time.Second, // nolint:gomnd
		context: context.Background(),
		log:     ctrl.Log.WithName("legacy").WithName("client"),
	}
}

// NewWithServer new client with provided server.
func NewWithServer(s *model.HarborServer) *Client {
	// Initialize with default settings
	c := New()
	c.server = s
	c.harborClient = s.ClientLegacy()

	return c
}

func (c *Client) WithServer(s *model.HarborServer) *Client {
	if s != nil {
		c.server = s
		c.harborClient = s.ClientLegacy()
	}

	return c
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

func (c *Client) CheckHealth() (*models.OverallHealthStatus, error) {
	params := products.NewGetHealthParamsWithContext(c.context).
		WithTimeout(c.timeout)

	res, err := c.harborClient.Client.Products.GetHealth(params, c.harborClient.Auth)
	if err != nil {
		return nil, err
	}

	return res.Payload, nil
}

func (c *Client) CreateRobotAccount(projectID int64) (*model.Robot, error) {
	if projectID <= 0 {
		return nil, errors.New("invalid project id")
	}

	if c.harborClient == nil {
		return nil, errors.New("nil harbor client")
	}

	params := products.NewPostProjectsProjectIDRobotsParamsWithContext(c.context).
		WithTimeout(c.timeout).
		WithProjectID(projectID).
		WithRobot(&models.RobotAccountCreate{
			Access: []*models.RobotAccountAccess{
				{
					Action:   "push",
					Resource: fmt.Sprintf("/project/%d/repository", projectID),
				},
			},
			Description: "automated by harbor automation operator",
			ExpiresAt:   -1, // never
			Name:        strings.RandomName("4k8s"),
		})

	res, err := c.harborClient.Client.Products.PostProjectsProjectIDRobots(params, c.harborClient.Auth)
	if err != nil {
		return nil, err
	}

	rid, err := strings.ExtractID(res.Location)
	if err != nil {
		// ignore this error that should never happen
		c.log.Error(err, "location", res.Location)
	}

	return &model.Robot{
		ID:    rid,
		Name:  res.Payload.Name,
		Token: res.Payload.Token,
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

	params := products.NewDeleteProjectsProjectIDRobotsRobotIDParamsWithContext(c.context).
		WithTimeout(c.timeout).
		WithProjectID(projectID).
		WithRobotID(robotID)

	if _, err := c.harborClient.Client.Products.DeleteProjectsProjectIDRobotsRobotID(params, c.harborClient.Auth); err != nil {
		return err
	}

	return nil
}

func (c *Client) GetRobotAccount(projectID, robotID int64) (*model.Robot, error) {
	if projectID <= 0 {
		return nil, errors.New("invalid project id")
	}

	if robotID <= 0 {
		return nil, errors.New("invalid robot id")
	}

	if c.harborClient == nil {
		return nil, errors.New("nil harbor client")
	}

	params := products.NewGetProjectsProjectIDRobotsRobotIDParamsWithContext(c.context).
		WithTimeout(c.timeout).
		WithProjectID(projectID).
		WithRobotID(robotID)

	res, err := c.harborClient.Client.Products.GetProjectsProjectIDRobotsRobotID(params, c.harborClient.Auth)
	if err != nil {
		return nil, err
	}

	return &model.Robot{
		ID:   robotID,
		Name: res.Payload.Name,
	}, nil
}
