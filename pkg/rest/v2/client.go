package v2

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/client/health"
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
