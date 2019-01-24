package faastagger

import (
	"context"
	"log"
	"net/url"
	"os"

	"github.com/vmware/govmomi/vim25/soap"

	"github.com/vmware/govmomi/vapi/tags"

	"github.com/vmware/govmomi/vapi/rest"

	"github.com/vmware/govmomi/vim25/types"

	"github.com/vmware/govmomi"
)

// Client is the tagging client used to tag a VM
type Client struct {
	vc *govmomi.Client
	// rest *rest.Client
	manager *tags.Manager
	log     Logger
}

// Logger is an interface to make logging configurable
type Logger interface {
	Printf(format string, v ...interface{})
}

// New creates a tagging client with a custom logger. If logger is nil, log.Logger will be used.
func New(ctx context.Context, logger Logger, vc string, vcuser string, vcpass string, insecure bool) (*Client, error) {
	if logger == nil {
		logger = log.New(os.Stdout, "", log.LstdFlags)
	}
	u, err := soap.ParseURL(vc)
	if err != nil {
		log.Printf("could not parse vCenter client URL: %v", err)
		return nil, err
	}

	u.User = url.UserPassword(vcuser, vcpass)
	c, err := govmomi.NewClient(ctx, u, insecure)
	if err != nil {
		log.Printf("could not get vCenter client: %v", err)
		return nil, err
	}

	r := rest.NewClient(c.Client)
	err = r.Login(ctx, u.User)
	if err != nil {
		log.Printf("could not get VAPI REST client: %v", err)
		return nil, err
	}
	tm := tags.NewManager(r)

	return &Client{vc: c, manager: tm, log: logger}, nil
}

// TagVM adds a tag to a virtual machine (ManagedObjectReference)
func (c *Client) TagVM(ctx context.Context, moref *types.ManagedObjectReference, tagID string) error {
	return c.manager.AttachTag(ctx, tagID, moref)
}

// Close closes all connections to vCenter
func (c *Client) Close(ctx context.Context) error {
	err := c.vc.Logout(ctx)
	if err != nil {
		c.log.Printf("could not close connection to vCenter")
		return err
	}

	err = c.manager.Logout(ctx)
	if err != nil {
		c.log.Printf("could not close VAPI REST connection to vCenter")
		return err
	}

	return nil
}
