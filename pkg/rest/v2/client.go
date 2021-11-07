package v2

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/goharbor/go-client/pkg/sdk/v2.0/client/project"
	v2models "github.com/goharbor/go-client/pkg/sdk/v2.0/models"
	"github.com/goharbor/harbor-operator/pkg/rest/model"
	utilstring "github.com/goharbor/harbor-operator/pkg/utils/strings"
	"github.com/pkg/errors"
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
}

// New V2 client.
func New() *Client {
	// Initialize with default settings
	return &Client{
		timeout: 30 * time.Second, // nolint:gomnd
		context: context.Background(),
	}
}

// NewWithServer new V2 client with provided server.
func NewWithServer(s *model.HarborServer) *Client {
	// Initialize with default settings
	c := New()
	c.server = s
	c.harborClient = s.ClientV2()

	return c
}

func (c *Client) WithServer(s *model.HarborServer) *Client {
	if s != nil {
		c.server = s
		c.harborClient = s.ClientV2()
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
		WithProject(&v2models.ProjectReq{
			ProjectName: name,
			Metadata: &v2models.ProjectMetadata{
				Public: "false",
			},
		})

	cp, err := c.harborClient.Client.Project.CreateProject(cparams, c.harborClient.Auth)
	if err != nil {
		return -1, fmt.Errorf("ensure project error: %w", err)
	}

	return utilstring.ExtractID(cp.Location)
}

// GetProject gets the project data.
func (c *Client) GetProject(name string) (*v2models.Project, error) {
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

	res, err := c.harborClient.Client.Project.ListProjects(params, c.harborClient.Auth)
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
		WithProjectID((int64)(p.ProjectID))
	if _, err = c.harborClient.Client.Project.DeleteProject(params, c.harborClient.Auth); err != nil {
		return err
	}

	return nil
}
